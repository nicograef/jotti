#!/usr/bin/env bash
# check-lib.sh — fixture test for the lib.sh helpers the prod scripts depend on
#
# Usage:
#   make check-repo   # or: ./scripts/check-lib.sh
#
# What it does:
#   1. Writes .env fixtures into a throwaway directory.
#   2. Asserts read_env's value for each (quotes, CRLF, "=" in the value, last line wins).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

FIXTURE_DIR="$(mktemp -d)"
trap 'rm -rf "$FIXTURE_DIR"' EXIT
cd "$FIXTURE_DIR"

failures=0

# expect_read_env NAME ENV_CONTENT KEY WANT — ENV_CONTENT is a printf format.
expect_read_env() {
  local name="$1" content="$2" key="$3" want="$4" got
  # shellcheck disable=SC2059
  printf "$content" >.env
  got="$(read_env "$key")"
  if [[ "$got" != "$want" ]]; then
    error "read_env $name: want '$want', got '$got'"
    failures=$((failures + 1))
  fi
}

expect_read_env "double quotes + CRLF" 'KEY="x"\r\n' KEY x
expect_read_env "single quotes" "KEY='x'\n" KEY x
expect_read_env "no quotes" 'KEY=x\n' KEY x
expect_read_env "no quotes + CRLF" 'KEY=x\r\n' KEY x
expect_read_env "= in the value" 'KEY=a=b\n' KEY a=b
expect_read_env "= in a quoted value" 'KEY="a=b"\n' KEY a=b
expect_read_env "one quote pair only" 'KEY=""x""\n' KEY '"x"'
expect_read_env "unbalanced quote kept" 'KEY="x\n' KEY '"x'
expect_read_env "surrounding blanks" 'KEY=  x  \n' KEY x
expect_read_env "empty value" 'KEY=\n' KEY ''
expect_read_env "missing key" 'OTHER=x\n' KEY ''
expect_read_env "prefix key not matched" 'KEYX=y\nKEY=x\n' KEY x
expect_read_env "last line wins" 'KEY=old\nKEY=new\n' KEY new

if ((failures > 0)); then
  fatal "$failures lib.sh fixture(s) failed."
fi

info "lib.sh fixtures pass: read_env strips blanks, CRLF and one pair of quotes."
