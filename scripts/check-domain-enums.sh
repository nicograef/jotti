#!/usr/bin/env bash
set -euo pipefail

# jotti — Kategorie (backend/domain/druckstation, backend/domain/produkt) and
# Steuersatz (backend/domain/steuer) are closed value sets; the domain
# constants are their one source. A hand-written literal outside the domain
# packages drifts silently the moment a value is renamed or a station added
# (backend/domain/druckstation/druckstation.go's AlleKategorien() is exactly
# the fix for the two OneOf calls this gate used to catch).
#
# The check derives the literal set from the const blocks in
# backend/domain/druckstation/druckstation.go and backend/domain/steuer/steuer.go
# (druckstation's five Kategorie values are a superset of produkt's three), then
# scans every tracked non-test *.go file under backend/ outside backend/domain/**
# and the generated backend/sqlc/dbgen/** for a line that names "Kategorie" or
# "Steuersatz" (case-insensitive — covers the Go field/type and the lowerCamelCase
# JSON/variable spelling) and also quotes one of those literals. Test files are
# out of scope: literal category/tax-rate fixtures in *_test.go are existing,
# reviewed practice throughout the suite. The line-based identifier match is a
# heuristic, not a type check — a literal reached only through an unrelated map
# key (no "kategorie"/"steuersatz" on the same line) would not be caught.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

KATEGORIE_SRC="backend/domain/druckstation/druckstation.go"
STEUERSATZ_SRC="backend/domain/steuer/steuer.go"

mapfile -t kategorien < <(
  grep -oE '\bKategorie\s*=\s*"[a-z]+"' "$KATEGORIE_SRC" | grep -oE '"[a-z]+"' | tr -d '"'
)
mapfile -t steuersaetze < <(
  grep -oE '\bSteuersatz\s*=\s*"[a-z]+"' "$STEUERSATZ_SRC" | grep -oE '"[a-z]+"' | tr -d '"'
)

if [ "${#kategorien[@]}" -eq 0 ]; then
  fatal "no Kategorie literals found in $KATEGORIE_SRC — did the const block move?"
fi
if [ "${#steuersaetze[@]}" -eq 0 ]; then
  fatal "no Steuersatz literals found in $STEUERSATZ_SRC — did the const block move?"
fi

literal_pattern="$(
  {
    printf '%s\n' "${kategorien[@]}"
    printf '%s\n' "${steuersaetze[@]}"
  } | paste -sd '|' -
)"

# `:(glob)` makes `/**/` mean "zero or more directories" for git pathspecs.
mapfile -t files < <(git ls-files ':(glob)backend/**/*.go' \
  ':(exclude,glob)backend/domain/**' \
  ':(exclude,glob)backend/sqlc/dbgen/**' \
  ':(exclude,glob)backend/**/*_test.go')

violations=0
for file in "${files[@]}"; do
  hits="$(grep -niE "(kategorie|steuersatz)" "$file" | grep -E "\"(${literal_pattern})\"" || true)"
  [ -z "$hits" ] && continue

  while IFS= read -r hit; do
    error "$file: Kategorie/Steuersatz literal outside backend/domain/**: ${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [ "$violations" -gt 0 ]; then
  fatal "$violations domain enum literal(s) found outside backend/domain/**."
fi

info "No Kategorie or Steuersatz literal outside backend/domain/** (non-test backend/**/*.go)."
