#!/usr/bin/env bash
# check-crm.sh — repo gate: unit tests of the private-register tool scripts/crm/crm.py
#
# Usage:
#   make check-repo   # or: ./scripts/check-crm.sh
#
# The tests use invented fixtures only; the real register in .crm/ is never read.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

if ! python3 -c 'import sys; sys.exit(sys.version_info < (3, 11))' 2>/dev/null; then
  fatal "python3 >= 3.11 is required for scripts/crm (tomllib); see scripts/setup-dev-tools.sh."
fi

python3 -m unittest -q scripts/crm/test_crm.py
info "scripts/crm tests pass."
