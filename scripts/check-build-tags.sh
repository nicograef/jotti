#!/usr/bin/env bash
set -euo pipefail

# jotti — every backend/**/*_test.go must declare exactly one //go:build tag,
# either "unit" or "integration". Without this, a file without a tag runs
# under every build (both `go test -tags=unit` and `-tags=integration`), and
# a file with the wrong tag silently skips golangci-lint or `make test`.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

violations=0

# `:(glob)` makes `/**/` mean "zero or more directories". Without the magic
# git matches `**` like a plain `*`, which needs at least one directory and
# so skips a test file lying directly in backend/.
while IFS= read -r file; do
  count="$(grep -c '^//go:build' "$file" || true)"
  if [ "$count" -ne 1 ]; then
    error "$file: expected exactly one //go:build line, found $count"
    violations=$((violations + 1))
    continue
  fi

  tag="$(grep '^//go:build' "$file")"
  if [ "$tag" != "//go:build unit" ] && [ "$tag" != "//go:build integration" ]; then
    error "$file: //go:build line must be 'unit' or 'integration', found: $tag"
    violations=$((violations + 1))
  fi
done < <(git ls-files ':(glob)backend/**/*_test.go')

if [ "$violations" -gt 0 ]; then
  fatal "$violations backend test file(s) missing a valid //go:build tag."
fi

info "All backend test files declare exactly one //go:build tag (unit or integration)."
