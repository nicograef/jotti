#!/usr/bin/env bash
# check-private-data.sh — repo gate: no club contact or non-public club name in a tracked file
#
# Usage:
#   make check-repo   # or: ./scripts/check-private-data.sh
#
# Reads the search terms from the private register (CRM_DIR, default: the main
# checkout's .crm/) via `crm.py pii-terms`. Without a register, as in CI, it
# passes with a notice. It scans the working tree and every line the unpushed
# commits add. A hit prints only file and line or the commit, never the term.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"
# Case folding of umlauts in grep -i needs a UTF-8 locale.
export LC_ALL=C.UTF-8

CRM_DIR="${CRM_DIR:-$(git rev-parse --path-format=absolute --git-common-dir)/../.crm}"
export CRM_DIR

if [[ ! -f "$CRM_DIR/register.toml" ]]; then
  info "No private register at $CRM_DIR; private-data check skipped."
  exit 0
fi

terms_file="$(mktemp)"
log_file="$(mktemp)"
trap 'rm -f "$terms_file" "$log_file"' EXIT
python3 "$SCRIPT_DIR/crm/crm.py" pii-terms >"$terms_file"
count="$(grep -c . "$terms_file" || true)"
if (( count == 0 )); then
  info "The private register yields no search terms; private-data check skipped."
  exit 0
fi

failed=0

# With pipefail the substitution reports git grep's code: 1 means no match.
set +e
hits="$(git grep -n -I -F -i -w -f "$terms_file" | cut -d: -f1,2)"
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
    failed=1
    ;;
  *)
    fatal "git grep failed with exit code $rc."
    ;;
esac

if upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  range="$upstream..HEAD"
elif git rev-parse -q --verify origin/HEAD >/dev/null && git rev-parse -q --verify HEAD >/dev/null; then
  range="origin/HEAD..HEAD"
else
  range=""
  info "No upstream and no origin/HEAD; unpushed commits not checked."
fi

if [[ -n "$range" ]]; then
  git log -p --no-color --no-ext-diff --no-textconv --format='commit %H' "$range" >"$log_file" \
    || fatal "git log $range failed."
  # Prefixes every added hunk line with its commit, so a hit names the commit and not the term.
  set +e
  commits="$(awk '/^commit [0-9a-f]+$/ { sha = $2; hunk = 0; next }
                  /^diff --git / { hunk = 0; next }
                  /^@@/ { hunk = 1; next }
                  hunk && /^\+/ { print sha "\t" substr($0, 2) }' "$log_file" \
    | grep -F -i -w -f "$terms_file" | cut -f1 | sort -u)"
  rc=$?
  set -e
  case "$rc" in
    1)
      info "No commit in $range adds any of the $count private terms."
      ;;
    0)
      while IFS= read -r sha; do
        error "commit $sha: adds a name or e-mail from the private register"
      done <<<"$commits"
      failed=1
      ;;
    *)
      fatal "grep over $range failed with exit code $rc."
      ;;
  esac
fi

if (( failed )); then
  fatal "Private data in tracked files or unpushed commits; remove it before pushing."
fi
