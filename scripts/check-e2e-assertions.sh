#!/usr/bin/env bash
set -euo pipefail

# jotti — two weak assertions that let a failed e2e measurement pass silently:
# `page.waitForLoadState('networkidle')` (Playwright marks it DISCOURAGED — it
# waits for a quiet network, not for the DOM state a spec depends on) and a
# `?? 0` fallback on a measured value (a missing or failed measurement then
# compares as 0 and satisfies the assertion instead of failing it, see
# e2e/support/viewport.ts's erwarteKeinenHorizontalenUeberlauf). Pure `//`
# comment lines are skipped, so documenting a forbidden pattern does not trip
# the gate.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

# `:(glob)` makes `/**/` mean "zero or more directories" for git pathspecs.
mapfile -t files < <(git ls-files \
  ':(glob)e2e/tests/**/*.ts' \
  ':(glob)e2e/support/**/*.ts' \
  ':(glob)e2e/helpers/**/*.ts')

violations=0
for file in "${files[@]}"; do
  hits="$(awk '
    /^[[:space:]]*\/\// { next }
    /networkidle/ || /\?\?[[:space:]]*0([^0-9]|$)/ { printf "%d:%s\n", FNR, $0 }
  ' "$file")"
  [ -z "$hits" ] && continue

  while IFS= read -r hit; do
    error "$file:${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [ "$violations" -gt 0 ]; then
  fatal "$violations weak e2e assertion(s) found (networkidle or \`?? 0\`) under e2e/tests, e2e/support and e2e/helpers."
fi

info "No networkidle wait or \`?? 0\` fallback in e2e/tests, e2e/support or e2e/helpers."
