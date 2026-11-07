#!/usr/bin/env bash
# check-e2e-assertions.sh — repo gate: no `?? 0` fallback on a measured e2e value
#
# Usage:
#   make check-repo   # or: ./scripts/check-e2e-assertions.sh
#
# A `?? 0` fallback lets a failed measurement pass silently: a missing value then
# compares as 0 (see e2e/support/viewport.ts's erwarteKeinenHorizontalenUeberlauf).
# The networkidle wait is the oxlint rule playwright/no-networkidle. Pure `//`
# comment lines are skipped, so documenting the pattern does not trip the gate.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

# `:(glob)` makes `/**/` mean "zero or more directories" for git pathspecs.
mapfile -t files < <(git ls-files \
  ':(glob)e2e/tests/**/*.ts' \
  ':(glob)e2e/support/**/*.ts')

violations=0
for file in "${files[@]}"; do
  hits="$(awk '
    /^[[:space:]]*\/\// { next }
    /\?\?[[:space:]]*0([^0-9]|$)/ { printf "%d:%s\n", FNR, $0 }
  ' "$file")"
  [[ -z "$hits" ]] && continue

  while IFS= read -r hit; do
    error "$file:${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations \`?? 0\` fallback(s) found under e2e/tests and e2e/support."
fi

info "No \`?? 0\` fallback in e2e/tests or e2e/support."
