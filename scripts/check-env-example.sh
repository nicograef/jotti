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

mapfile -t allow_entries < <(
  [[ -f "$ALLOWLIST" ]] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST"
)

allow_names=()
allow_hits=()
for entry in "${allow_entries[@]+"${allow_entries[@]}"}"; do
  name="$(awk '{print $1}' <<<"$entry")"
  reason="$(awk '{print $2}' <<<"$entry")"
  [[ "${reason:0:1}" == "#" ]] || fatal "$ALLOWLIST: reason missing: $entry"
  allow_names+=("$name")
  allow_hits+=(0)
done

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

  if [[ "${#allow_names[@]}" -gt 0 ]]; then
    for i in "${!allow_names[@]}"; do
      if [[ "$var" == "${allow_names[$i]}" ]]; then
        allow_hits[i]=$((allow_hits[i] + 1))
        continue 2
      fi
    done
  fi

  if ! grep -qE "^#?[[:space:]]*${var}=" "$TEMPLATE"; then
    error "$loc: \${$var} is missing from $TEMPLATE (add it, commented out if optional)."
    violations=$((violations + 1))
  fi
done

if [[ "${#allow_names[@]}" -gt 0 ]]; then
  for i in "${!allow_names[@]}"; do
    if [[ "${allow_hits[$i]}" -eq 0 ]]; then
      error "$ALLOWLIST: no compose file reads ${allow_names[$i]} any more."
      violations=$((violations + 1))
    fi
  done
fi

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations compose variable(s) without an entry in $TEMPLATE (see $ALLOWLIST for intended gaps)."
fi

info "Every variable the compose files read is in $TEMPLATE or allowlisted."
