#!/usr/bin/env bash
# check-timezone.sh — repo gate: every printed or exported timestamp names its zone
#
# Usage:
#   make check-repo   # or: ./scripts/check-timezone.sh
#   Exceptions: scripts/check-timezone.allow ("path fragment  # reason").
#
# The containers run in UTC: a .Format() that does not pass through .In(zone)
# shifts receipts, work slips, file names and reports by up to two hours, late in
# the evening by a day. The check is line-based: a zone on an earlier .Format() of
# the same line does not shield a later one, but a call chain spanning two lines
# passes unnoticed. Pure comment lines are skipped.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-timezone.allow"

load_allowlist "$ALLOWLIST" fragment

# `:(glob)` makes `/**/` mean "zero or more directories". Without the magic
# git matches `**` like a plain `*`, which needs at least one directory.
mapfile -t files < <(git ls-files \
  ':(glob)backend/api/druck/**/*.go' \
  ':(glob)backend/api/fiskal/**/*.go' \
  ':(glob)backend/api/reporting/**/*.go')

violations=0
for file in "${files[@]}"; do
  # awk rather than grep: the decision needs the position of every .In( relative
  # to every .Format( on the line. A line is reported at its first unzoned call.
  hits="$(awk '
    /^[[:space:]]*\/\// { next }
    {
      start = 1
      while ((offset = index(substr($0, start), ".Format(")) > 0) {
        call = start + offset - 1
        if (index(substr($0, start, call - start), ".In(") == 0) {
          printf "%d:%s\n", FNR, $0
          break
        }
        start = call + length(".Format(")
      }
    }
  ' "$file")"
  [[ -z "$hits" ]] && continue

  while IFS= read -r hit; do
    lineno="${hit%%:*}"
    code="${hit#*:}"

    allowlisted "$file" "$code" && continue

    error "$file:$lineno: .Format() without .In(zone):${code}"
    violations=$((violations + 1))
  done <<<"$hits"
done

violations=$((violations + $(report_stale_allowlist)))

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations time zone rule violation(s) found (allow one in $ALLOWLIST)."
fi

info "Every .Format() in the print, fiskal and reporting paths goes through .In(zone)."
