#!/usr/bin/env bash
# prod-harden.sh — optional, idempotent hardening of a public VPS host
#
# Usage:
#   make prod-harden   # after the stack is up; SKIP_FAIL2BAN=1 skips fail2ban
#
# What it does (after the handbook's setup-server.sh):
#   1. ufw: SSH rate-limited, 80/443 allowed, everything else denied inbound.
#   2. A fail2ban sshd jail on the systemd journal.
#   3. unattended-upgrades for daily security updates.
#   4. An sshd drop-in that allows key logins only and no root login.
# Not part of prod-init.sh. Postgres is never exposed: docker-compose.prod.yml
# publishes only 80/443.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# apt_install PACKAGE — install a package on Debian/Ubuntu. Returns non-zero when
# apt-get is unavailable or the install fails (caller decides fatal vs. skip).
apt_install() {
  command -v apt-get &>/dev/null || return 1
  $SUDO apt-get update -qq && $SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y "$1"
}

# has_systemd — true when systemd is PID 1 (sd_booted); false in a plain container.
has_systemd() {
  [[ -d /run/systemd/system ]]
}

# Every privileged command is prefixed with $SUDO so the script works both as
# root (SUDO empty) and as a sudo-capable user.
SUDO=""
if [[ "$(id -u)" -ne 0 ]]; then
  if command -v sudo &>/dev/null; then
    SUDO="sudo"
  else
    fatal "Run as root or install sudo: the hardening steps need root privileges."
  fi
fi

# Detect the SSH port from the active session first (most reliable), then from
# sshd_config, falling back to 22. Allowing this port BEFORE enabling the
# default-deny firewall is what prevents locking yourself out.
detect_ssh_port() {
  local port=""
  if [[ -n "${SSH_CONNECTION:-}" ]]; then
    port="$(awk '{print $4}' <<<"$SSH_CONNECTION")"
  fi
  if [[ -z "$port" && -r /etc/ssh/sshd_config ]]; then
    port="$(grep -iE '^[[:space:]]*Port[[:space:]]+[0-9]+' /etc/ssh/sshd_config | tail -n1 | awk '{print $2}')"
  fi
  [[ "$port" =~ ^[0-9]+$ ]] || port=22
  printf '%s\n' "$port"
}

SSH_PORT="${SSH_PORT:-$(detect_ssh_port)}"
[[ "$SSH_PORT" =~ ^[0-9]+$ ]] || fatal "SSH_PORT must be a number (got: $SSH_PORT)."

# key_sudo_user — prints the first non-root member of group sudo whose
# authorized_keys holds a key; the drop-in disables root login, so that user is the way back in.
key_sudo_user() {
  local user home
  for user in $(getent group sudo | cut -d: -f4 | tr ',' ' '); do
    [[ "$user" == root ]] && continue
    home="$(getent passwd "$user" | cut -d: -f6)"
    if [[ -n "$home" ]] && $SUDO grep -qsE '(^|[[:space:]])(ssh-|ecdsa-|sk-)' "$home/.ssh/authorized_keys"; then
      printf '%s\n' "$user"
      return 0
    fi
  done
  return 1
}

if [[ -f /etc/ssh/sshd_config ]]; then
  if ! SSH_USER="$(key_sudo_user)"; then
    error "Kein Benutzer in der Gruppe sudo mit SSH-Schlüssel gefunden. Die Härtung sperrt die Root-Anmeldung, ihr würdet euch aussperren."
    error "Behebung: Benutzer anlegen (adduser NAME), in die Gruppe sudo aufnehmen (usermod -aG sudo NAME),"
    error "euren Schlüssel kopieren (ssh-copy-id NAME@SERVER), Anmeldung als NAME testen, dann erneut ausführen."
    fatal "Abgebrochen. Es wurde nichts geändert."
  fi
  info "SSH login after hardening: $SSH_USER (key, sudo)"
fi

info "SSH port to keep open: $SSH_PORT"

echo ""
warn "This will harden THIS host:"
warn "  ufw: limit $SSH_PORT/tcp (SSH), allow 80/tcp, 443/tcp, 443/udp; deny all other inbound."
warn "  fail2ban sshd jail, unattended-upgrades, SSH logins by key only, no root login."
read -r -p "Continue? Type 'yes' to proceed: " answer
[[ "$answer" == "yes" ]] || fatal "Aborted by user. Nothing was changed."

if ! command -v ufw &>/dev/null; then
  info "ufw not found, installing..."
  apt_install ufw || fatal "Could not install ufw automatically. Install it, then re-run: sudo apt-get install -y ufw"
fi

# `ufw limit` and `ufw allow` update a rule that already exists, so re-running
# is a no-op. `limit` denies an address after six connections within 30 s. The
# default-deny policy only takes effect on `enable`, which happens last.
$SUDO ufw limit "$SSH_PORT/tcp" comment 'SSH'
$SUDO ufw allow 80/tcp comment 'jotti HTTP'
$SUDO ufw allow 443/tcp comment 'jotti HTTPS'
$SUDO ufw allow 443/udp comment 'jotti HTTP/3'
$SUDO ufw default deny incoming
$SUDO ufw default allow outgoing
$SUDO ufw --force enable
info "ufw active (SSH rate-limited, 80/443 allowed, everything else denied inbound)."

# Debian and Ubuntu log sshd to the journal only, so the jail reads it there.
configure_fail2ban() {
  $SUDO mkdir -p /etc/fail2ban/jail.d
  $SUDO tee /etc/fail2ban/jail.d/jotti-sshd.local >/dev/null <<EOF
# Managed by jotti scripts/prod-harden.sh — protects SSH from brute-force.
[sshd]
enabled      = true
port         = $SSH_PORT
backend      = systemd
journalmatch = _SYSTEMD_UNIT=ssh.service + _COMM=sshd + _COMM=sshd-session
EOF
  if has_systemd; then
    $SUDO systemctl enable fail2ban &>/dev/null || true
    $SUDO systemctl restart fail2ban
    info "fail2ban sshd jail active (port $SSH_PORT)."
  else
    warn "fail2ban jail written, but systemd is not running — start the service manually."
  fi
}

fail2ban_enabled=false
if [[ "${SKIP_FAIL2BAN:-}" == "1" ]]; then
  info "Skipping fail2ban (SKIP_FAIL2BAN=1)."
else
  if ! command -v fail2ban-client &>/dev/null; then
    info "fail2ban not found, installing..."
    apt_install fail2ban || warn "Could not install fail2ban automatically. Skipping (install later: sudo apt-get install -y fail2ban)."
  fi
  if command -v fail2ban-client &>/dev/null; then
    configure_fail2ban
    fail2ban_enabled=true
  fi
fi

if ! command -v unattended-upgrade &>/dev/null; then
  info "unattended-upgrades not found, installing..."
  apt_install unattended-upgrades || fatal "Could not install unattended-upgrades automatically. Install it, then re-run: sudo apt-get install -y unattended-upgrades"
fi
# The file `dpkg-reconfigure -plow unattended-upgrades` writes: a daily package
# list update and upgrade run. Reboots stay manual.
$SUDO tee /etc/apt/apt.conf.d/20auto-upgrades >/dev/null <<'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
EOF
info "unattended-upgrades active (daily security updates)."

# sshd keeps the first value it reads and cloud images ship 50-cloud-init.conf,
# so the 00- prefix wins.
configure_sshd() {
  local dropin=/etc/ssh/sshd_config.d/00-jotti-hardening.conf
  local include='Include /etc/ssh/sshd_config.d/*.conf'
  if ! $SUDO grep -qxF "$include" /etc/ssh/sshd_config; then
    $SUDO sed -i "1i $include" /etc/ssh/sshd_config
  fi
  $SUDO install -d -m 0755 /etc/ssh/sshd_config.d /run/sshd
  $SUDO tee "$dropin" >/dev/null <<'EOF'
# Managed by jotti scripts/prod-harden.sh — SSH logins by key only, no root login.
PubkeyAuthentication yes
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin no
EOF
  if ! $SUDO sshd -t; then
    $SUDO rm -f "$dropin"
    fatal "sshd rejected the drop-in, so it was removed again. The SSH configuration is unchanged."
  fi
  if has_systemd; then
    $SUDO systemctl try-reload-or-restart ssh
  fi
}

ssh_key_only=false
if [[ -f /etc/ssh/sshd_config ]]; then
  configure_sshd
  ssh_key_only=true
  info "sshd allows key logins only, root login is off."
else
  warn "No /etc/ssh/sshd_config found — skipping the key-only SSH drop-in."
fi

echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti — Server Hardening Complete"
echo "=========================================="
echo ""
echo "  Firewall (ufw): SSH ($SSH_PORT/tcp) rate-limited; 80/tcp, 443/tcp, 443/udp"
echo "                  allowed; all other inbound denied. Postgres stays internal."
if [[ "$fail2ban_enabled" == true ]]; then
  echo "  fail2ban:       sshd jail active."
else
  echo "  fail2ban:       not configured."
fi
echo "  Updates:        unattended-upgrades daily; reboots after kernel updates stay manual."
if [[ "$ssh_key_only" == true ]]; then
  echo "  SSH:            key logins only, no root login ($SSH_USER has sudo)."
else
  echo "  SSH:            unchanged (no sshd found)."
fi
echo ""
warn "Before logging out, open a SECOND SSH session to confirm you are not locked out."
echo "=========================================="
