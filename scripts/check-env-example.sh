#!/usr/bin/env bash
# check-env-example.sh — repo gate: every variable a compose file reads is in .env.example
#
# Usage:
#   make check-repo   # or: ./scripts/check-env-example.sh
#   Exceptions: scripts/check-env-example.allow (variable name, "# reason").
#
# A key counts as present when .env.example sets it or names it commented out
# ("# KEY="). Full comment lines and "$$" escapes in the compose files are skipped.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

TEMPLATE=".env.example"
ALLOWLIST="scripts/check-env-example.allow"

load_allowlist "$ALLOWLIST" name

mapfile -t compose_files < <(git ls-files ':(glob)docker-compose*.yml')
[[ "${#compose_files[@]}" -gt 0 ]] || fatal "no docker-compose*.yml found."

# "<VAR><TAB><file>:<line>" for every ${VAR} or $VAR a compose file interpolates.
mapfile -t uses < <(
  for file in "${compose_files[@]}"; do
    awk -v file="$file" '
      /^[[:space:]]*#/ { next }
      {
        line = $0
        while (match(line, /(^|[^$])\$\{?[A-Za-z_][A-Za-z0-9_]*/)) {
          ref = substr(line, RSTART, RLENGTH)
          sub(/^[^$]*\$\{?/, "", ref)
          print ref "\t" file ":" FNR
          line = substr(line, RSTART + RLENGTH)
        }
      }
    ' "$file"
  done
)

violations=0
declare -A seen=()
for use in "${uses[@]+"${uses[@]}"}"; do
  var="${use%%$'\t'*}"
  loc="${use#*$'\t'}"
  [[ -n "${seen[$var]:-}" ]] && continue
  seen["$var"]=1

  allowlisted "$var" && continue

  if ! grep -qE "^#?[[:space:]]*${var}=" "$TEMPLATE"; then
    error "$loc: \${$var} is missing from $TEMPLATE (add it, commented out if optional)."
    violations=$((violations + 1))
  fi
done

violations=$((violations + $(report_stale_allowlist "no compose file reads %s any more.")))

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations compose variable(s) without an entry in $TEMPLATE (see $ALLOWLIST for intended gaps)."
fi

info "Every variable the compose files read is in $TEMPLATE or allowlisted."
