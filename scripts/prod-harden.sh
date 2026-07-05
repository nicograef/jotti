#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Optional Server Hardening (self-hosted production, Weg B)
#
# Opt-in, idempotent host hardening for a public VPS running jotti. Configures a
# ufw firewall that allows only SSH plus the jotti web ports (80/443), denies
# everything else inbound, and optionally enables fail2ban's sshd jail. This is
# NOT part of prod-init.sh — run it deliberately, after the stack is up. Steps:
#   1. Require root (or sudo) and detect the SSH port (never lock yourself out)
#   2. ufw: allow SSH FIRST, then 80/443(+udp), default-deny inbound, enable
#   3. Optional fail2ban sshd jail (skip with SKIP_FAIL2BAN=1)
#   4. Reminder about unattended-upgrades (printed, not installed)
#
# Postgres is never exposed: docker-compose.prod.yml only publishes 80/443 on
# the host, so the database stays on the internal Docker network.
#
# Configuration (environment):
#   SSH_PORT        SSH port to keep open (default: auto-detected, then 22)
#   SKIP_FAIL2BAN   set to 1 to skip the fail2ban step
#
# Usage: sudo ./scripts/prod-harden.sh  (or `make prod-harden`)
# =============================================================================

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# apt_install PACKAGE — install a package on Debian/Ubuntu. Returns non-zero when
# apt-get is unavailable or the install fails (caller decides fatal vs. skip).
apt_install() {
  command -v apt-get &>/dev/null || return 1
  $SUDO apt-get update -qq && $SUDO env DEBIAN_FRONTEND=noninteractive apt-get install -y "$1"
}

# ---------------------------------------------------------------------------
# Step 1 — Require root (or sudo) and detect the SSH port
# ---------------------------------------------------------------------------
# Every privileged command is prefixed with $SUDO so the script works both as
# root (SUDO empty) and as a sudo-capable user.
SUDO=""
if [[ "$(id -u)" -ne 0 ]]; then
  if command -v sudo &>/dev/null; then
    SUDO="sudo"
  else
    fatal "Run as root or install sudo: ufw/fail2ban changes need root privileges."
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

info "SSH port to keep open: $SSH_PORT"

echo ""
warn "This will configure a ufw firewall on THIS host:"
warn "  allow $SSH_PORT/tcp (SSH), 80/tcp, 443/tcp, 443/udp; deny all other inbound."
read -r -p "Continue? Type 'yes' to proceed: " answer
[[ "$answer" == "yes" ]] || fatal "Aborted by user. Nothing was changed."

# ---------------------------------------------------------------------------
# Step 2 — ufw firewall (SSH allowed before enable; idempotent rules)
# ---------------------------------------------------------------------------
if ! command -v ufw &>/dev/null; then
  info "ufw not found, installing..."
  apt_install ufw || fatal "Could not install ufw automatically. Install it, then re-run: sudo apt-get install -y ufw"
fi

# `ufw allow` skips rules that already exist, so re-running is a no-op. The
# default-deny policy only takes effect on `enable`, which happens last — after
# SSH is already allowed.
$SUDO ufw allow "$SSH_PORT/tcp" comment 'SSH'
$SUDO ufw allow 80/tcp comment 'jotti HTTP'
$SUDO ufw allow 443/tcp comment 'jotti HTTPS'
$SUDO ufw allow 443/udp comment 'jotti HTTP/3'
$SUDO ufw default deny incoming
$SUDO ufw default allow outgoing
$SUDO ufw --force enable
info "ufw active (SSH + 80/443 allowed, everything else denied inbound)."

# ---------------------------------------------------------------------------
# Step 3 — fail2ban sshd jail (optional)
# ---------------------------------------------------------------------------
configure_fail2ban() {
  $SUDO mkdir -p /etc/fail2ban/jail.d
  $SUDO tee /etc/fail2ban/jail.d/jotti-sshd.local >/dev/null <<EOF
# Managed by jotti scripts/prod-harden.sh — protects SSH from brute-force.
[sshd]
enabled = true
port    = $SSH_PORT
EOF
  if command -v systemctl &>/dev/null; then
    $SUDO systemctl enable fail2ban &>/dev/null || true
    $SUDO systemctl restart fail2ban
    info "fail2ban sshd jail active (port $SSH_PORT)."
  else
    warn "fail2ban jail written, but no systemctl found — start the service manually."
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

# ---------------------------------------------------------------------------
# Step 4 — unattended-upgrades reminder (printed, not installed)
# ---------------------------------------------------------------------------
echo ""
info "Recommended: enable automatic security updates (this script does NOT install them):"
echo "    sudo apt-get install -y unattended-upgrades"
echo "    sudo dpkg-reconfigure -plow unattended-upgrades"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti — Server Hardening Complete"
echo "=========================================="
echo ""
echo "  Firewall (ufw): SSH ($SSH_PORT/tcp), 80/tcp, 443/tcp, 443/udp allowed;"
echo "                  all other inbound denied. Postgres stays internal."
if [[ "$fail2ban_enabled" == true ]]; then
  echo "  fail2ban:       sshd jail active."
else
  echo "  fail2ban:       not configured."
fi
echo ""
warn "Before logging out, open a SECOND SSH session to confirm you are not locked out."
echo "=========================================="
