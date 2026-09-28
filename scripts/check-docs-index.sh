#!/usr/bin/env bash
# check-docs-index.sh — repo gate: every doc in docs/ is linked from docs/README.md
#
# Usage:
#   make check-repo   # or: ./scripts/check-docs-index.sh
#
# A file docs/NAME.md needs a link "](NAME.md"; a directory docs/DIR/ needs a link
# into it, "](DIR/".
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

INDEX="docs/README.md"
[[ -f "$INDEX" ]] || fatal "$INDEX is missing."

# One entry per top-level doc: "NAME.md" for a file, "DIR/" for a directory.
mapfile -t entries < <(
  git ls-files docs | awk -F/ 'NF == 2 && $2 ~ /\.md$/ { print $2 } NF > 2 { print $2 "/" }' | sort -u
)

violations=0
for entry in "${entries[@]}"; do
  [[ "$entry" == "README.md" ]] && continue
  if ! grep -qF "](${entry}" "$INDEX"; then
    error "$INDEX: no link to docs/$entry."
    violations=$((violations + 1))
  fi
done

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations doc(s) missing from $INDEX."
fi

info "Every file and directory in docs/ is linked from $INDEX."
