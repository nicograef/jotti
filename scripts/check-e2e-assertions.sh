#!/usr/bin/env bash
set -euo pipefail

# jotti — two weak-assertion patterns that let a failed e2e measurement pass
# silently: `page.waitForLoadState('networkidle')` (Playwright itself marks
# this DISCOURAGED — it waits for quiet network, not for the DOM state a spec
# actually depends on) and a `?? 0` fallback on a measured value (a missing or
# failed measurement then compares as 0, which usually satisfies the
# assertion instead of failing it — see e2e/support/viewport.ts's
# erwarteKeinenHorizontalenUeberlauf, the case this gate was written for).
#
# Scanned are the tracked TypeScript files under e2e/tests and e2e/support —
# the Playwright specs and their shared helpers. e2e/website/**(*.mjs) is out
# of scope: those are static-site smoke scripts (CSP check, screenshots) for
# the marketing site, not app specs, and a different concern than this gate.
# A pure `//` comment line is skipped, so a line documenting the forbidden
# pattern (as this file's own header, or a fix's explanatory comment, does)
# does not trip the gate itself.

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
    /networkidle/ || /\?\?[[:space:]]*0([^0-9]|$)/ { printf "%d:%s\n", FNR, $0 }
  ' "$file")"
  [ -z "$hits" ] && continue

  while IFS= read -r hit; do
    error "$file:${hit}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [ "$violations" -gt 0 ]; then
  fatal "$violations weak e2e assertion(s) found (networkidle or \`?? 0\`) under e2e/tests and e2e/support."
fi

info "No networkidle wait or \`?? 0\` fallback in e2e/tests or e2e/support."
