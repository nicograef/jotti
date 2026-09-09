#!/usr/bin/env bash
set -euo pipefail

# jotti — Kategorie (backend/domain/druckstation, backend/domain/produkt) and
# Steuersatz (backend/domain/steuer) are closed value sets; the domain
# constants are their one source. A hand-written literal outside the domain
# packages drifts silently the moment a value is renamed or a station added.
#
# The check derives the literal set from the const blocks in
# backend/domain/druckstation/druckstation.go and backend/domain/steuer/steuer.go
# (druckstation's five Kategorie values are a superset of produkt's three),
# then scans every tracked non-test *.go file under backend/ outside
# backend/domain/** and the generated backend/sqlc/dbgen/** for a Go string
# literal equal to one of those values, on any line, standalone — no
# co-occurring "Kategorie"/"Steuersatz" identifier is required, so a bare map
# key (druckstationen["abholbon"]) or a value split across lines from its
# OneOf( call is caught the same as an inline literal.
#
# scripts/check-domain-enums.allow lists exceptions: a value that
# legitimately shares text with a Kategorie/Steuersatz literal without being
# one (e.g. a spelling-correction word list). An entry that matches nothing
# turns the gate red and is to be deleted.

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

# Allowlist: path, then a space-free code fragment, then "#" and the reason.
mapfile -t allow_entries < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST"
)

allow_paths=()
allow_needles=()
allow_hits=()
for entry in "${allow_entries[@]+"${allow_entries[@]}"}"; do
  path="$(printf '%s\n' "$entry" | awk '{print $1}')"
  needle="$(printf '%s\n' "$entry" | awk '{print $2}')"
  reason="$(printf '%s\n' "$entry" | awk '{print $3}')"
  if [ -z "$needle" ] || [ "${needle:0:1}" = "#" ]; then
    fatal "$ALLOWLIST: entry without a code fragment: $entry"
  fi
  # Without this the awk split would silently cut a fragment at its first
  # space and exempt more lines than the entry names.
  if [ "${reason:0:1}" != "#" ]; then
    fatal "$ALLOWLIST: code fragment with a space, or reason missing: $entry"
  fi
  allow_paths+=("$path")
  allow_needles+=("$needle")
  allow_hits+=(0)
done

# `:(glob)` makes `/**/` mean "zero or more directories" for git pathspecs.
mapfile -t files < <(git ls-files ':(glob)backend/**/*.go' \
  ':(exclude,glob)backend/domain/**' \
  ':(exclude,glob)backend/sqlc/dbgen/**' \
  ':(exclude,glob)backend/**/*_test.go')

violations=0
for file in "${files[@]}"; do
  hits="$(grep -nE "\"(${literal_pattern})\"" "$file" || true)"
  [ -z "$hits" ] && continue

  while IFS= read -r hit; do
    code="${hit#*:}"

    exempt=0
    if [ "${#allow_paths[@]}" -gt 0 ]; then
      for i in "${!allow_paths[@]}"; do
        if [ "$file" = "${allow_paths[$i]}" ] && [[ "$code" == *"${allow_needles[$i]}"* ]]; then
          allow_hits[i]=$((allow_hits[i] + 1))
          exempt=1
          break
        fi
      done
    fi
    [ "$exempt" -eq 1 ] && continue

    error "$file: Kategorie/Steuersatz literal outside backend/domain/**: ${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [ "${#allow_paths[@]}" -gt 0 ]; then
  for i in "${!allow_paths[@]}"; do
    if [ "${allow_hits[$i]}" -eq 0 ]; then
      error "$ALLOWLIST: exception matches nothing any more: ${allow_paths[$i]} ${allow_needles[$i]}"
      violations=$((violations + 1))
    fi
  done
fi

if [ "$violations" -gt 0 ]; then
  fatal "$violations domain enum literal(s) found outside backend/domain/** (allow one in $ALLOWLIST)."
fi

info "No Kategorie or Steuersatz literal outside backend/domain/** (non-test backend/**/*.go)."
