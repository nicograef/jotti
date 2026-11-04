#!/usr/bin/env bash
# check-private-data.sh — repo gate: no club contact or non-public club name in a tracked file
#
# Usage:
#   make check-repo   # or: ./scripts/check-private-data.sh
#
# Reads the search terms from the private register (CRM_DIR, default: the main
# checkout's .crm/) via `crm.py pii-terms`. Without a register, as in CI, it
# passes with a notice. A hit prints only file and line, never the term.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

CRM_DIR="${CRM_DIR:-$(git rev-parse --path-format=absolute --git-common-dir)/../.crm}"
export CRM_DIR

if [[ ! -f "$CRM_DIR/register.toml" ]]; then
  info "No private register at $CRM_DIR; private-data check skipped."
  exit 0
fi

terms="$(python3 "$SCRIPT_DIR/crm/crm.py" pii-terms)"
count="$(grep -c . <<<"$terms" || true)"
if (( count == 0 )); then
  info "The private register yields no search terms; private-data check skipped."
  exit 0
fi

# With pipefail the substitution reports git grep's code: 1 means no match.
set +e
hits="$(git grep -n -I -F -i -w -f - <<<"$terms" | cut -d: -f1,2)"
rc=$?
set -e

case "$rc" in
  1)
    info "No tracked file contains any of the $count private terms."
    ;;
  0)
    while IFS= read -r hit; do
      error "$hit: contains a name or e-mail from the private register"
    done <<<"$hits"
    fatal "Private data in tracked files; remove it before pushing."
    ;;
  *)
    fatal "git grep failed with exit code $rc."
    ;;
esac
