#!/usr/bin/env bash
# check-links.sh — repo gate: every relative *.md path resolves to a tracked file
#
# Usage:
#   make check-repo   # or: ./scripts/check-links.sh
#   Exceptions: scripts/check-links.allow ("path line-fragment  # reason").
#
# A path is checked relative to the mentioning file and to the repo root: docs
# cross-reference each other by sibling path, ops scripts reference docs/ from
# the root.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-links.allow"

# A relative *.md path: no whitespace or markup delimiters, ending in ".md".
LINK_RE='[^]"'"'"'`()<>[*[:space:]]+\.md'

# link_files — every tracked file whose *.md mentions must resolve, AGENTS.md and
# .claude/** included. Out: history (changelog, plans, gitleaks fingerprints), frozen,
# generated and vendored files, lockfiles, and this gate's own files, which quote paths as examples.
link_files() {
  git ls-files \
    ':(exclude)CHANGELOG.md' \
    ':(exclude).gitleaksignore' \
    ':(glob,exclude)docs/plans/**' \
    ':(glob,exclude)docs/rechtsquellen/**' \
    ':(glob,exclude)database/migrations/**' \
    ':(glob,exclude)backend/sqlc/dbgen/**' \
    ':(glob,exclude)**/pnpm-lock.yaml' \
    ':(glob,exclude)**/go.sum' \
    ':(exclude)scripts/check-links.sh' \
    ':(exclude)scripts/check-links.allow'
}

# Allowlist: "path  fragment  # reason"; a dead link on a line of path that
# contains fragment is exempt. An entry that exempts nothing fails the gate.
load_allowlist "$ALLOWLIST" fragment

mapfile -t files < <(link_files)

violations=0
for file in "${files[@]}"; do
  dir="$(dirname "$file")"
  while IFS=: read -r lineno raw; do
    # Drop a leading "@" (CLAUDE.md's `@AGENTS.md` import syntax). LINK_RE stops
    # the match at ".md", so nothing trailing needs stripping.
    link="${raw#@}"
    [[ -z "$link" ]] && continue
    case "$link" in
    //* | http:* | https:* | \$\{*) continue ;; # a URL or a runtime-built path, not a repo path
    esac

    [[ -f "$dir/$link" ]] && continue
    [[ -f "$link" ]] && continue
    allowlisted "$file" "$(sed -n "${lineno}p" "$file")" && continue

    error "$file:$lineno: dead link to '$link'"
    violations=$((violations + 1))
  done < <(grep -aonE "$LINK_RE" "$file" 2>/dev/null || true)
done

violations=$((violations + $(report_stale_allowlist)))

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations dead markdown link(s) (see $ALLOWLIST to allow a specific line)."
fi

info "All relative *.md references resolve to a tracked file."
