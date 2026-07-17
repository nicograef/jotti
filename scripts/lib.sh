#!/usr/bin/env bash
# jotti — shared shell helpers for scripts/*.sh: color vars, log helpers and
# .env reading. Source, don't execute:
#
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   # shellcheck source=scripts/lib.sh
#   . "$SCRIPT_DIR/lib.sh"
#
# All log helpers write to stderr, so a caller's stdout stays free for a
# machine-readable return value (see ops-smoke.sh's TSV protocol).

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1" >&2; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1" >&2; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# read_env KEY — read a single value from .env without executing the file
# (passwords may contain shell-special characters). Returns the last match,
# trimmed of surrounding whitespace.
read_env() {
  local key="$1"
  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
}
