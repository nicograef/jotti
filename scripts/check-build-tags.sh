#!/usr/bin/env bash
# check-build-tags.sh — repo gate: unit tests carry no Go build tag
#
# Usage:
#   make check-repo   # or: ./scripts/check-build-tags.sh
#
# The only allowed constraint in backend/ is `//go:build integration` on a
# *_test.go file. Any other tag (a typo, a leftover `unit`) silently drops the
# file from `go test ./...` and golangci-lint.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

violations=0

# `:(glob)` makes `/**/` mean "zero or more directories"; without it git matches
# `**` like `*`, which needs one directory and skips a file in backend/.
while IFS= read -r file; do
  tags="$(grep '^//go:build' "$file" || true)"
  [[ -z "$tags" ]] && continue

  case "$file" in
    *_test.go) ;;
    *)
      error "$file: only test files may carry a //go:build line, found: $tags"
      violations=$((violations + 1))
      continue
      ;;
  esac

  if [[ "$tags" != "//go:build integration" ]]; then
    error "$file: the only allowed //go:build line is 'integration', found: $tags"
    violations=$((violations + 1))
  fi
done < <(git ls-files ':(glob)backend/**/*.go')

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations backend file(s) carry a build tag other than a single //go:build integration."
fi

info "Backend build tags are valid: unit tests untagged, integration tests tagged //go:build integration."
