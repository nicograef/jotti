#!/usr/bin/env bash
set -euo pipefail

# jotti — AGENTS.md rule 18 ("nur der aktuelle Stand"): prose describes the
# current state only. This gate rejects two things: history words that frame
# a statement relative to a former state ("früher", "bisher", ...), and
# session-scoped jargon that only makes sense while a plan or design handoff
# is in flight ("Phase 3", "Muster 05", "Befund #1", "Design-Handoff",
# "Seit Version 0.17").

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-prose.allow"

# Whole words only (grep -w): "bisher" must not flag "bisherige", and the
# structural patterns ("Phase [0-9]+", ...) must not flag a longer identifier
# that merely contains them.
PATTERN='(früher|frueher|vormals|bisher|bislang|neuerdings|Phase [0-9]+|NEU[0-9]{2}|Muster [0-9]+|Befund #[0-9]*|Design-Handoff|design_handoff|Seit Version [0-9]+|Ab Version [0-9]+)'

# Paths frozen by the freeze discipline, rule texts that quote the banned
# words themselves, and generated/vendored files that were never authored
# as prose.
mapfile -t files < <(git ls-files \
  ':(glob,exclude)CHANGELOG.md' \
  ':(glob,exclude)docs/adrs/**' \
  ':(glob,exclude)docs/plans/**' \
  ':(glob,exclude)docs/rechtsquellen/**' \
  ':(glob,exclude)database/migrations/**' \
  ':(glob,exclude)backend/sqlc/dbgen/**' \
  ':(glob,exclude)AGENTS.md' \
  ':(glob,exclude).github/copilot-instructions.md' \
  ':(glob,exclude).github/instructions/**' \
  ':(glob,exclude).claude/**' \
  ':(glob,exclude)reverse-proxy/caddyfile.go' \
  ':(glob,exclude)**/pnpm-lock.yaml' \
  ':(glob,exclude)**/go.sum')

# Allowlist: one path per line, an optional trailing "# reason", blank lines
# and full-line comments ignored.
mapfile -t allowed < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST" | awk '{print $1}'
)

violations=0
for file in "${files[@]}"; do
  skip=0
  for a in "${allowed[@]+"${allowed[@]}"}"; do
    if [ "$file" = "$a" ]; then
      skip=1
      break
    fi
  done
  [ "$skip" -eq 1 ] && continue

  if hits="$(grep -nwE "$PATTERN" "$file" 2>/dev/null)"; then
    while IFS= read -r hit; do
      error "$file:$hit"
    done <<<"$hits"
    violations=$((violations + $(printf '%s\n' "$hits" | wc -l)))
  fi
done

if [ "$violations" -gt 0 ]; then
  fatal "$violations line(s) with historical or handoff prose (see $ALLOWLIST to allow a specific file)."
fi

info "No historical or handoff prose found outside $ALLOWLIST."
