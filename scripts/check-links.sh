#!/usr/bin/env bash
set -euo pipefail

# jotti — every relative *.md path mentioned anywhere in the tracked tree
# (a markdown link, a backtick code span, a plain path in a script or config
# comment) must resolve to a real tracked file. Checked both relative to the
# mentioning file and relative to the repo root, since docs cross-reference
# each other by sibling path and ops scripts reference docs/ from the root.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-links.allow"

# A relative *.md path: no whitespace or markup delimiters, ending in ".md".
LINK_RE='[^]"'"'"'`()<>[*[:space:]]+\.md'

# Same source and exceptions as scripts/check-prose.sh.
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

# Allowlist: one path per line, an optional trailing "# reason". A listed
# file is skipped entirely (same granularity as scripts/check-prose.allow).
mapfile -t allowed < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST" | awk '{print $1}'
)
is_allowed_file() {
  local needle="$1"
  for a in "${allowed[@]+"${allowed[@]}"}"; do
    [ "$needle" = "$a" ] && return 0
  done
  return 1
}

violations=0
for file in "${files[@]}"; do
  is_allowed_file "$file" && continue
  dir="$(dirname "$file")"
  while IFS=: read -r lineno raw; do
    # Drop a leading "@" (CLAUDE.md's `@AGENTS.md` import syntax). LINK_RE
    # already stops the match at ".md", so raw never carries a trailing
    # anchor or sentence punctuation to strip.
    link="${raw#@}"
    [ -z "$link" ] && continue
    case "$link" in
    //* | http:* | https:* | \$\{*) continue ;; # a URL or a runtime-built path, not a repo path
    esac

    [ -f "$dir/$link" ] && continue
    [ -f "$link" ] && continue

    error "$file:$lineno: dead link to '$link'"
    violations=$((violations + 1))
  done < <(grep -aonE "$LINK_RE" "$file" 2>/dev/null || true)
done

if [ "$violations" -gt 0 ]; then
  fatal "$violations dead markdown link(s) (see $ALLOWLIST to allow a specific one)."
fi

info "All relative *.md references resolve to a tracked file."
