#!/usr/bin/env bash
# check-prose.sh — repo gate: prose describes the current state and keeps the prose caps
#
# Usage:
#   make check-repo                            # or: ./scripts/check-prose.sh
#   PROSE_CAPS=1 ./scripts/check-prose.sh      # also enforce the prose caps
#   ./scripts/check-prose.sh --list-caps       # print every cap violation as file:line: reason
#   Exceptions: scripts/check-prose.allow (one path per line, "# reason").
#
# Rejected are words that frame a statement against a former state and
# session-scoped jargon from a plan or handoff in flight (see PATTERN).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-prose.allow"

PATTERN='(bisher|bisherige[nrs]?|früher|frühere[nrs]?|frueher|fruehere[nrs]?|bislang|vormals|neuerdings|Phase [0-9]+|NEU[0-9]{2}|Muster [0-9]+|Befund #[0-9]*|Design-Handoff|design_handoff|Seit Version [0-9]+|Ab Version [0-9]+)'

PROSE_MAX_WORDS=20
PROSE_MAX_PARA_LINES=3

# prose_files — the tracked files this gate polices. Excluded: the changelog and
# plans, paths frozen by the freeze discipline, and generated or vendored files.
prose_files() {
  git ls-files \
    ':(glob,exclude)CHANGELOG.md' \
    ':(glob,exclude)docs/plans/**' \
    ':(glob,exclude)docs/rechtsquellen/**' \
    ':(glob,exclude)database/migrations/**' \
    ':(glob,exclude)backend/sqlc/dbgen/**' \
    ':(glob,exclude)**/pnpm-lock.yaml' \
    ':(glob,exclude)**/go.sum'
}

# cap_files — the Markdown files under the prose caps. Excluded: plans, legal texts
# binding as worded, and the operator's Verfahrensdokumentation template.
cap_files() {
  git ls-files '*.md' \
    ':(glob,exclude)docs/plans/**' \
    ':(glob,exclude)docs/rechtsquellen/**' \
    ':(exclude)TERMS.md' \
    ':(exclude)CLA.md' \
    ':(exclude)SERVICE.md' \
    ':(exclude)docs/verfahrensdokumentation.md'
}

# prose_scan FILE — prints one cap violation per line as "file:line: reason".
#
# Port of prose_scan in the handbook's scripts/check-repo.sh (front matter, code, comments,
# headings, tables skipped). German additions: single letters ("z. B.") and the German
# entries in abbrev() do not end a sentence.
prose_scan() {
  LC_ALL=C awk -v file="$1" -v maxwords="$PROSE_MAX_WORDS" -v maxpara="$PROSE_MAX_PARA_LINES" '
    function clean(s,   pre, mid, post) {
      gsub(/`[^`]*`/, " ", s)
      gsub(/!\[[^]]*\]\([^)]*\)/, " ", s)
      while (match(s, /\[[^]]*\]\([^)]*\)/)) {
        pre = substr(s, 1, RSTART - 1)
        mid = substr(s, RSTART, RLENGTH)
        post = substr(s, RSTART + RLENGTH)
        sub(/\]\([^)]*\)$/, "", mid)
        sub(/^\[/, "", mid)
        s = pre mid post
      }
      gsub(/<[^ <>]*>/, " ", s)
      gsub(/[*_]/, "", s)
      return s
    }

    function words(s,   n, i, a, c) {
      n = split(s, a, /[ \t]+/)
      c = 0
      for (i = 1; i <= n; i++)
        if (a[i] ~ /[A-Za-z0-9]/) c++
      return c
    }

    function abbrev(s) {
      if (s ~ /(^|[ (.])[A-Za-z]\.$/) return 1
      return (s ~ /(^|[ (])(e\.g|i\.e|etc|vs|cf|approx|resp|Dr|Mr|Ms|No|bzw|ggf|inkl|exkl|zzgl|vgl|usw|sog|evtl|bspw|bzgl|ca|Nr|Abs|Art|Kap|Tz|Rz|Anh|Anl|gem|lt|max|min|Std|Min|Mio|Mrd|Tsd|ff)\.$/)
    }

    function sentences(s, a,   i, c, cur, n, nxt) {
      n = 0
      cur = ""
      for (i = 1; i <= length(s); i++) {
        c = substr(s, i, 1)
        cur = cur c
        if (c != "." && c != "!" && c != "?") continue
        nxt = substr(s, i + 1, 1)
        if (nxt != "" && nxt != " ") continue
        if (abbrev(cur)) continue
        a[++n] = cur
        cur = ""
      }
      if (cur ~ /[A-Za-z0-9]/) a[++n] = cur
      return n
    }

    function snippet(s,   n, i, a, out) {
      sub(/^[ \t]+/, "", s)
      n = split(s, a, /[ \t]+/)
      out = ""
      for (i = 1; i <= n && i <= 8; i++) out = (out == "") ? a[i] : out " " a[i]
      return (n > 8) ? out " ..." : out
    }

    function flushpara() {
      if (para > maxpara)
        printf "%s:%d: paragraph of %d lines (cap %d)\n", file, parastart, para, maxpara
      para = 0
    }

    function checkblock(   n, i, a, w) {
      if (block ~ /[A-Za-z]/) {
        n = sentences(block, a)
        for (i = 1; i <= n; i++) {
          w = words(a[i])
          if (w > maxwords)
            printf "%s:%d: sentence of %d words (cap %d): %s\n", file, blockline, w, maxwords, snippet(a[i])
        }
      }
      block = ""
    }

    BEGIN { fm = 0; fence = 0; comment = 0; para = 0; parastart = 0; block = ""; blockline = 0; prevtype = "" }

    { raw = $0 }

    NR == 1 && raw ~ /^---[ \t]*$/ { fm = 1; next }
    fm { if (raw ~ /^---[ \t]*$/) fm = 0; next }

    raw ~ /^[ \t]*(```|~~~)/ { flushpara(); checkblock(); prevtype = ""; fence = !fence; next }
    fence { next }

    # Code spans go first so a backticked "<!--" stays prose; "@" keeps the line non-blank.
    { gsub(/`[^`]*`/, "@", raw) }

    {
      if (comment) {
        if (raw ~ /-->/) { sub(/^.*-->/, "", raw); comment = 0 }
        else next
      }
      gsub(/<!--.*-->/, " ", raw)
      if (index(raw, "<!--") > 0) {
        raw = substr(raw, 1, index(raw, "<!--") - 1)
        comment = 1
      }
    }

    raw ~ /^[ \t]*$/ { flushpara(); checkblock(); prevtype = ""; next }
    raw ~ /^#{1,6} / { flushpara(); checkblock(); prevtype = ""; next }
    raw ~ /^[ \t]*\|/ { flushpara(); checkblock(); prevtype = ""; next }
    raw ~ /^[ \t]*(-{3,}|\*{3,}|_{3,})[ \t]*$/ { flushpara(); checkblock(); prevtype = ""; next }

    {
      body = raw
      if (raw ~ /^[ \t]*>/) {
        type = "quote"
        sub(/^[ \t]*>[ \t]*/, "", body)
      } else if (raw ~ /^[ \t]*([-*+]|[0-9]+[.)])[ \t]/) {
        type = "bullet"
        sub(/^[ \t]*([-*+]|[0-9]+[.)])[ \t]+/, "", body)
      } else if (raw ~ /^[ \t]/) {
        type = "cont"
      } else {
        type = "para"
      }

      text = clean(body)

      # Paragraph runs count source lines; blockquotes and indented lines continue their block.
      if (type == "para") {
        if (prevtype != "para") { flushpara(); checkblock(); parastart = NR }
        para++
      } else if (type != "cont" && !(type == "quote" && prevtype == "quote")) {
        flushpara(); checkblock()
      }
      prevtype = type

      if (text !~ /[A-Za-z]/) next
      if (block == "") blockline = NR
      block = (block == "") ? text : block " " text
    }

    END { flushpara(); checkblock() }
  ' "$1"
}

cap_violations() {
  local file
  while IFS= read -r file; do
    prose_scan "$file"
  done < <(cap_files)
}

if [[ "${1:-}" == "--list-caps" ]]; then
  list="$(cap_violations)"
  [[ -z "$list" ]] && exit 0
  printf '%s\n' "$list"
  exit 1
fi

mapfile -t files < <(prose_files)

# Allowlist: one path per line, an optional trailing "# reason".
mapfile -t allowed < <(
  [[ -f "$ALLOWLIST" ]] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST" | awk '{print $1}'
)

violations=0
for file in "${files[@]}"; do
  skip=0
  for a in "${allowed[@]+"${allowed[@]}"}"; do
    if [[ "$file" = "$a" ]]; then
      skip=1
      break
    fi
  done
  [[ "$skip" -eq 1 ]] && continue

  if hits="$(grep -inwE "$PATTERN" "$file" 2>/dev/null)"; then
    while IFS= read -r hit; do
      error "$file:$hit"
    done <<<"$hits"
    violations=$((violations + $(printf '%s\n' "$hits" | wc -l)))
  fi
done

if [[ "${PROSE_CAPS:-0}" == "1" ]]; then
  while IFS= read -r violation; do
    error "$violation"
    violations=$((violations + 1))
  done < <(cap_violations)
fi

if [[ "$violations" -gt 0 ]]; then
  fatal "$violations violation(s) of the current-state rule or the prose caps (see $ALLOWLIST to allow a file for the current-state rule)."
fi

info "No historical or handoff prose found outside $ALLOWLIST."
