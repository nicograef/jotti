#!/usr/bin/env bash
set -euo pipefail

# jotti — every "…" (German low-high opening quote „, ASCII closing quote ")
# citation of a UI control in docs/leitfaden/** must exist verbatim in
# frontend/src: a citation that used to match a button, menu item or heading
# has to be corrected the moment the frontend text changes underneath it.
#
# Not every "…" citation in the guide is a UI control, though — Windows
# dialogs, a router's own web UI, GitHub, ELSTER, printed receipts and legal
# wording get quoted too. scripts/check-ui-labels.allow names each of those
# with its real source; an entry that no longer matches any citation turns
# the gate red and is to be deleted.
#
# A citation broken across a line wrap (Markdown prose reflow, or a ">"
# blockquote continuation) is joined into one string before matching: a
# leading ">" blockquote marker is stripped from every line, fenced code
# blocks are dropped (quoted console output inside one is not a UI
# citation), and all whitespace — including the wrap itself — collapses to
# single spaces.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-ui-labels.allow"

# Allowlist: one non-UI quote per line, then "# " and its source. Blank
# lines and full-line comments are ignored.
mapfile -t allow_quotes < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST" |
    sed -E 's/[[:space:]]*#.*$//; s/[[:space:]]+$//'
)
allow_hits=()
for _ in "${allow_quotes[@]+"${allow_quotes[@]}"}"; do
  allow_hits+=(0)
done

mapfile -t files < <(git ls-files ':(glob)docs/leitfaden/**/*.md')

violations=0
for file in "${files[@]}"; do
  # Strip a leading "> " blockquote marker per line, drop fenced code
  # blocks, then squeeze the whole file (including every line break) to a
  # single space-separated line — awk/tr rather than perl or python, so the
  # gate stays free of new dependencies (Architectural decisions).
  joined="$(
    awk '
      /^```/ { infence = !infence; next }
      infence { next }
      { sub(/^>[[:space:]]?/, ""); printf "%s ", $0 }
    ' "$file" | tr -s '[:space:]' ' '
  )"

  while IFS= read -r quote; do
    [ -z "$quote" ] && continue

    allowed=0
    if [ "${#allow_quotes[@]}" -gt 0 ]; then
      for i in "${!allow_quotes[@]}"; do
        if [ "$quote" = "${allow_quotes[$i]}" ]; then
          allow_hits[i]=$((allow_hits[i] + 1))
          allowed=1
          break
        fi
      done
    fi
    [ "$allowed" -eq 1 ] && continue

    if ! grep -qrF -- "$quote" frontend/src; then
      error "$file: „$quote\" not found in frontend/src (see $ALLOWLIST to allow a non-UI quote)"
      violations=$((violations + 1))
    fi
  done < <(printf '%s' "$joined" | grep -oE '„[^"]*"' | sed -E 's/^„//; s/"$//')
done

if [ "${#allow_quotes[@]}" -gt 0 ]; then
  for i in "${!allow_quotes[@]}"; do
    if [ "${allow_hits[$i]}" -eq 0 ]; then
      error "$ALLOWLIST: entry matches no citation any more: „${allow_quotes[$i]}\""
      violations=$((violations + 1))
    fi
  done
fi

if [ "$violations" -gt 0 ]; then
  fatal "$violations UI-label citation issue(s) (see $ALLOWLIST to allow a non-UI quote)."
fi

info "Every „…\" UI-label citation in docs/leitfaden/** matches frontend/src or is allowlisted."
