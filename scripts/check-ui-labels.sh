#!/usr/bin/env bash
# check-ui-labels.sh — repo gate: every UI label the leitfaden cites exists in frontend/src
#
# Usage:
#   make check-repo   # or: ./scripts/check-ui-labels.sh
#   Exceptions: scripts/check-ui-labels.allow (one non-UI quote per line, "# source").
#
# Every „…" citation in docs/leitfaden/** must exist verbatim in frontend/src, so a
# citation is corrected the moment the frontend text changes. Citations of
# something else (Windows dialogs, a router's web UI, GitHub, ELSTER, printed
# receipts, legal wording) are allowlisted with their real source; an entry that
# matches no citation any more turns the gate red.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-ui-labels.allow"

# Allowlist: one non-UI quote per line, then "# " and its source.
load_allowlist "$ALLOWLIST" quote

mapfile -t files < <(git ls-files ':(glob)docs/leitfaden/**/*.md')

violations=0
for file in "${files[@]}"; do
  # Strip a leading "> " blockquote marker, drop fenced code blocks (quoted
  # console output is not a UI citation), then squeeze the file including every
  # line break to single spaces, so a citation broken across a line wrap still
  # matches — awk/tr, so this gate needs no interpreter its siblings don't.
  # Fences here sit indented inside list items, so the marker is matched after
  # optional leading whitespace.
  joined="$(
    awk '
      /^[[:space:]]*```/ { infence = !infence; next }
      infence { next }
      { sub(/^>[[:space:]]?/, ""); printf "%s ", $0 }
    ' "$file" | tr -s '[:space:]' ' '
  )"

  while IFS= read -r quote; do
    [[ -z "$quote" ]] && continue

    allowlisted "$quote" && continue

    if ! grep -qrF -- "$quote" frontend/src; then
      error "$file: „$quote\" not found in frontend/src (see $ALLOWLIST to allow a non-UI quote)"
      violations=$((violations + 1))
    fi
  done < <(printf '%s' "$joined" | grep -oE '„[^"]*"' | sed -E 's/^„//; s/"$//')
done

violations=$((violations + $(report_stale_allowlist 'entry matches no citation any more: „%s"')))

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations UI-label citation issue(s) (see $ALLOWLIST to allow a non-UI quote)."
fi

info "Every „…\" UI-label citation in docs/leitfaden/** matches frontend/src or is allowlisted."
