#!/usr/bin/env bash
set -euo pipefail

# jotti — one decidable language rule per area: Windows console output
# stays ASCII, backend Go comments carry real German umlauts, and
# packaging/**/*.cmd stays ASCII throughout (Windows batch files have no
# separate doc-comment channel). Identifiers never change here — only
# string literals, comments and .cmd text.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

# scripts/checklanguage is not part of any Go module (see go.work): passing
# the file directly, not a package path, builds it standalone and sidesteps
# "outside modules listed in go.work".
CHECKER="$SCRIPT_DIR/checklanguage/main.go"

violations=0

# Rule 1a: non-ASCII bytes in Go string literals under windows/**. Comments,
# *.manifest and *.syso are excluded — they carry German prose and never
# reach a Windows console.
mapfile -t windows_go_files < <(git ls-files ':(glob)windows/**/*.go')
if [ "${#windows_go_files[@]}" -gt 0 ]; then
  if hits="$(go run "$CHECKER" windows-strings "${windows_go_files[@]}")"; then
    :
  else
    status=$?
    if [ "$status" -ne 1 ]; then
      fatal "checklanguage windows-strings failed (exit $status)"
    fi
    while IFS= read -r hit; do
      error "$hit"
    done <<<"$hits"
    violations=$((violations + $(printf '%s\n' "$hits" | wc -l)))
  fi
fi

# Rule 1b: non-ASCII bytes anywhere in packaging/**/*.cmd (including REM
# comments — a Windows batch file has no console-vs-doc split like Go does).
mapfile -t cmd_files < <(git ls-files ':(glob)packaging/**/*.cmd')
if [ "${#cmd_files[@]}" -gt 0 ]; then
  if hits="$(go run "$CHECKER" cmd-ascii "${cmd_files[@]}")"; then
    :
  else
    status=$?
    if [ "$status" -ne 1 ]; then
      fatal "checklanguage cmd-ascii failed (exit $status)"
    fi
    while IFS= read -r hit; do
      error "$hit"
    done <<<"$hits"
    violations=$((violations + $(printf '%s\n' "$hits" | wc -l)))
  fi
fi

# Rule 2: transliterated umlaut words (fuer, ueber, koennen, ...) as whole
# words on Go comment lines under backend/**. Identifiers such as the
# `auftraege` parameter in backend/seed are untouched — only *ast.Comment
# text is scanned. backend/sqlc/dbgen/** is excluded like in
# check-prose.sh: it is generated code (AGENTS.md rule 14, "niemals
# editieren") whose comments come from database/migrations/** (frozen) and
# backend/sqlc/queries/**, not from hand-authored Go prose.
mapfile -t backend_go_files < <(git ls-files ':(glob)backend/**/*.go' ':(glob,exclude)backend/sqlc/dbgen/**')
if [ "${#backend_go_files[@]}" -gt 0 ]; then
  if hits="$(go run "$CHECKER" backend-comments "${backend_go_files[@]}")"; then
    :
  else
    status=$?
    if [ "$status" -ne 1 ]; then
      fatal "checklanguage backend-comments failed (exit $status)"
    fi
    while IFS= read -r hit; do
      error "$hit"
    done <<<"$hits"
    violations=$((violations + $(printf '%s\n' "$hits" | wc -l)))
  fi
fi

if [ "$violations" -gt 0 ]; then
  fatal "$violations language rule violation(s) found."
fi

info "Windows output is ASCII, packaging/**/*.cmd is ASCII, backend comments use real umlauts."
