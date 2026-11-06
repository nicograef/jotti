#!/usr/bin/env bash
# check-domain-enums.sh — repo gate: Kategorie and Steuersatz literals live in backend/domain only
#
# Usage:
#   make check-repo   # or: ./scripts/check-domain-enums.sh
#   Exceptions: scripts/check-domain-enums.allow ("path fragment  # reason").
#
# Kategorie (backend/domain/druckstation, backend/domain/produkt) and Steuersatz
# (backend/domain/steuer) are closed value sets owned by the domain constants; a
# literal copied elsewhere drifts silently the moment a value is renamed or a
# station added. druckstation's five Kategorie values are a superset of produkt's
# three, so druckstation.go alone defines the set. A match needs no co-occurring
# "Kategorie"/"Steuersatz" identifier, so a bare map key is caught like an inline
# literal.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

KATEGORIE_SRC="backend/domain/druckstation/druckstation.go"
STEUERSATZ_SRC="backend/domain/steuer/steuer.go"
ALLOWLIST="scripts/check-domain-enums.allow"

mapfile -t kategorien < <(
  grep -oE '\bKategorie\s*=\s*"[a-z]+"' "$KATEGORIE_SRC" | grep -oE '"[a-z]+"' | tr -d '"'
)
mapfile -t steuersaetze < <(
  grep -oE '\bSteuersatz\s*=\s*"[a-z]+"' "$STEUERSATZ_SRC" | grep -oE '"[a-z]+"' | tr -d '"'
)

if [[ "${#kategorien[@]}" -eq 0 ]]; then
  fatal "no Kategorie literals found in $KATEGORIE_SRC — did the const block move?"
fi
if [[ "${#steuersaetze[@]}" -eq 0 ]]; then
  fatal "no Steuersatz literals found in $STEUERSATZ_SRC — did the const block move?"
fi

literal_pattern="$(
  {
    printf '%s\n' "${kategorien[@]}"
    printf '%s\n' "${steuersaetze[@]}"
  } | paste -sd '|' -
)"

load_allowlist "$ALLOWLIST" fragment

# `:(glob)` makes `/**/` mean "zero or more directories" for git pathspecs.
mapfile -t files < <(git ls-files ':(glob)backend/**/*.go' \
  ':(exclude,glob)backend/domain/**' \
  ':(exclude,glob)backend/sqlc/dbgen/**' \
  ':(exclude,glob)backend/**/*_test.go')

violations=0
for file in "${files[@]}"; do
  hits="$(grep -nE "\"(${literal_pattern})\"" "$file" || true)"
  [[ -z "$hits" ]] && continue

  while IFS= read -r hit; do
    code="${hit#*:}"

    allowlisted "$file" "$code" && continue

    error "$file: Kategorie/Steuersatz literal outside backend/domain/**: ${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

violations=$((violations + $(report_stale_allowlist)))

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations domain enum literal(s) found outside backend/domain/** (allow one in $ALLOWLIST)."
fi

info "No Kategorie or Steuersatz literal outside backend/domain/** (non-test backend/**/*.go)."
