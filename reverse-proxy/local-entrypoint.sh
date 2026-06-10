#!/bin/sh
set -eu

CERT_DIR="/etc/nginx/certs"
CERT_FILE="${CERT_DIR}/selfsigned.crt"
KEY_FILE="${CERT_DIR}/selfsigned.key"
LAN_IP_FILE="${CERT_DIR}/.lan-ip"

info() {
  printf '[local-tls] %s\n' "$1"
}

detect_lan_ip() {
  default_gateway="$(ip route | awk '/^default / {print $3; exit}')"
  gateway_interface="$(ip route | awk '/^default / {print $5; exit}')"

  if [ -z "$default_gateway" ] || [ -z "$gateway_interface" ]; then
    return 1
  fi

  lan_ip="$(ip -4 addr show dev "$gateway_interface" | awk '/inet / {print $2; exit}' | cut -d/ -f1)"
  if [ -z "$lan_ip" ]; then
    return 1
  fi

  printf '%s' "$lan_ip"
}

create_certificate() {
  lan_ip="$1"

  tmp_config="$(mktemp)"
  cat >"$tmp_config" <<EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_req

[dn]
CN = ${lan_ip}

[v3_req]
subjectAltName = @alt_names

[alt_names]
IP.1 = ${lan_ip}
IP.2 = 127.0.0.1
DNS.1 = localhost
EOF

  openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -config "$tmp_config"

  rm -f "$tmp_config"
  printf '%s\n' "$lan_ip" >"$LAN_IP_FILE"
  chmod 600 "$KEY_FILE"

  info "Generated self-signed certificate for ${lan_ip}."
}

mkdir -p "$CERT_DIR"

if ! command -v openssl >/dev/null 2>&1; then
  info "Installing openssl package..."
  apk add --no-cache openssl >/dev/null
fi

current_lan_ip="$(detect_lan_ip)"
stored_lan_ip=""

if [ -f "$LAN_IP_FILE" ]; then
  stored_lan_ip="$(cat "$LAN_IP_FILE" 2>/dev/null || true)"
fi

if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ] && [ "$stored_lan_ip" = "$current_lan_ip" ]; then
  info "Certificate exists for ${current_lan_ip}. Reusing existing certificate."
else
  info "Certificate missing or LAN IP changed. Regenerating certificate."
  create_certificate "$current_lan_ip"
fi

exec nginx -g 'daemon off;'