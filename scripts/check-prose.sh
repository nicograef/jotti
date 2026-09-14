#!/usr/bin/env bash
set -euo pipefail

# jotti — AGENTS.md rule 18 ("nur der aktuelle Stand"): prose describes the
# current state only. Rejected are words that frame a statement against a former
# state and session-scoped jargon from a plan or handoff in flight (see PATTERN).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-prose.allow"

PATTERN='(bisher|bisherige[nrs]?|früher|frühere[nrs]?|frueher|fruehere[nrs]?|bislang|vormals|neuerdings|Phase [0-9]+|NEU[0-9]{2}|Muster [0-9]+|Befund #[0-9]*|Design-Handoff|design_handoff|Seit Version [0-9]+|Ab Version [0-9]+)'

# Excluded: paths frozen by the freeze discipline, rule texts that quote the
# banned words themselves, and generated or vendored files.
mapfile -t files < <(git ls-files \
  ':(glob,exclude)CHANGELOG.md' \
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

# Allowlist: one path per line, an optional trailing "# reason".
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

  if hits="$(grep -inwE "$PATTERN" "$file" 2>/dev/null)"; then
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
