#!/usr/bin/env bash
set -euo pipefail

# jotti — every printed, exported or served timestamp carries a deliberately
# chosen zone. The containers run in UTC: a .Format() on a time value that does
# not pass through .In(zone) shifts receipts, work slips, file names and reports
# by up to two hours — and late in the evening by a day.
#
# Scanned are all Go files under backend/api/druck, backend/api/fiskal and
# backend/api/reporting. A line containing .Format() passes when the same line
# carries .In( before it. The check is line-based: no call chain there spans two
# lines. Pure comment lines (leading //) are skipped.
#
# Exceptions live in scripts/check-timezone.allow: per line the path, then a
# fragment of the exempt code line (no spaces), then "# reason". An exception
# applies only to lines containing that fragment; one that matches nothing turns
# the gate red and is to be deleted.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-timezone.allow"

mapfile -t allow_entries < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST"
)

allow_paths=()
allow_needles=()
allow_hits=()
for entry in "${allow_entries[@]+"${allow_entries[@]}"}"; do
  path="$(printf '%s\n' "$entry" | awk '{print $1}')"
  needle="$(printf '%s\n' "$entry" | awk '{print $2}')"
  if [ -z "$needle" ] || [ "${needle:0:1}" = "#" ]; then
    fatal "$ALLOWLIST: entry without a code fragment: $entry"
  fi
  allow_paths+=("$path")
  allow_needles+=("$needle")
  allow_hits+=(0)
done

# `:(glob)` makes `/**/` mean "zero or more directories". Without the magic
# git matches `**` like a plain `*`, which needs at least one directory.
mapfile -t files < <(git ls-files \
  ':(glob)backend/api/druck/**/*.go' \
  ':(glob)backend/api/fiskal/**/*.go' \
  ':(glob)backend/api/reporting/**/*.go')

violations=0
for file in "${files[@]}"; do
  # awk rather than grep: the decision needs the order of .In( and .Format( in
  # the same line, not just their presence.
  hits="$(awk '
    /^[[:space:]]*\/\// { next }
    {
      pos = index($0, ".Format(")
      if (pos == 0) next
      zone = index($0, ".In(")
      if (zone > 0 && zone < pos) next
      printf "%d:%s\n", FNR, $0
    }
  ' "$file")"
  [ -z "$hits" ] && continue

  while IFS= read -r hit; do
    lineno="${hit%%:*}"
    code="${hit#*:}"

    exempt=0
    if [ "${#allow_paths[@]}" -gt 0 ]; then
      for i in "${!allow_paths[@]}"; do
        if [ "$file" = "${allow_paths[$i]}" ] && [[ "$code" == *"${allow_needles[$i]}"* ]]; then
          allow_hits[i]=$((allow_hits[i] + 1))
          exempt=1
          break
        fi
      done
    fi
    [ "$exempt" -eq 1 ] && continue

    error "$file:$lineno: .Format() without .In(zone):${code}"
    violations=$((violations + 1))
  done <<<"$hits"
done

if [ "${#allow_paths[@]}" -gt 0 ]; then
  for i in "${!allow_paths[@]}"; do
    if [ "${allow_hits[$i]}" -eq 0 ]; then
      error "$ALLOWLIST: exception matches nothing any more: ${allow_paths[$i]} ${allow_needles[$i]}"
      violations=$((violations + 1))
    fi
  done
fi

if [ "$violations" -gt 0 ]; then
  fatal "$violations time zone rule violation(s) found (allow one in $ALLOWLIST)."
fi

info "Every .Format() in the print, fiskal and reporting paths goes through .In(zone)."
