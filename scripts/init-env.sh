#!/usr/bin/env bash
# init-env.sh — create .env from .env.example with fresh random secrets
#
# Usage:
#   make init
#   ENV_FILE=.env.test ./scripts/init-env.sh
#
# What it does:
#   1. Leaves an existing ENV_FILE untouched (idempotent).
#   2. Copies ENV_TEMPLATE_FILE and replaces POSTGRES_PASSWORD, JWT_SECRET and
#      RELAY_AUTH_TOKEN with `openssl rand -hex 32` values.
#   3. Writes the result with mode 600.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"
cd "$PROJECT_ROOT"

ENV_FILE="${ENV_FILE:-.env}"
ENV_TEMPLATE_FILE="${ENV_TEMPLATE_FILE:-.env.example}"

if [[ -f "$ENV_FILE" ]]; then
  info "$ENV_FILE already exists; nothing changed."
  exit 0
fi

if [[ ! -f "$ENV_TEMPLATE_FILE" ]]; then
  fatal "Template missing: $ENV_TEMPLATE_FILE"
fi

if ! command -v openssl >/dev/null 2>&1; then
  fatal "openssl not found. Install it and rerun."
fi

generate_secret() {
  openssl rand -hex 32
}

postgres_password="$(generate_secret)"
jwt_secret="$(generate_secret)"
relay_auth_token="$(generate_secret)"

tmp_file="$(mktemp)"
cleanup() {
  rm -f "$tmp_file"
}
trap cleanup EXIT

if ! awk \
  -v postgres_password="$postgres_password" \
  -v jwt_secret="$jwt_secret" \
  -v relay_auth_token="$relay_auth_token" '
  BEGIN {
    found_postgres_password = 0
    found_jwt_secret = 0
    found_relay_auth_token = 0
  }
  /^POSTGRES_PASSWORD=/ {
    print "POSTGRES_PASSWORD=" postgres_password
    found_postgres_password = 1
    next
  }
  /^JWT_SECRET=/ {
    print "JWT_SECRET=" jwt_secret
    found_jwt_secret = 1
    next
  }
  /^RELAY_AUTH_TOKEN=/ {
    print "RELAY_AUTH_TOKEN=" relay_auth_token
    found_relay_auth_token = 1
    next
  }
  {
    print
  }
  END {
    if (!found_postgres_password || !found_jwt_secret || !found_relay_auth_token) {
      exit 1
    }
  }
' "$ENV_TEMPLATE_FILE" >"$tmp_file"; then
  fatal "Template is incomplete. Expected keys: POSTGRES_PASSWORD, JWT_SECRET, RELAY_AUTH_TOKEN"
fi

mv "$tmp_file" "$ENV_FILE"
chmod 600 "$ENV_FILE" 2>/dev/null || true

info "$ENV_FILE created."
