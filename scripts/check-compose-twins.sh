#!/usr/bin/env bash
# check-compose-twins.sh — repo gate: docker-compose.release.yml stays the twin of docker-compose.local.yml
#
# Usage:
#   make check-repo   # or: ./scripts/check-compose-twins.sh
#
# Clubs run the release file from the Windows ZIP, while `make local-up` tests the
# local file. The two may differ only in comments, a source build versus a GHCR image
# per buildable service, and the local migrations bind mount. Anything else is drift.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

# normalize FILE — prints FILE without comments and blank lines, with every build
# block and jotti GHCR image as "image: <IMAGE>" and without the migrations mount.
normalize() {
  awk '
    function indent_of(s) { match(s, /^ */); return RLENGTH }
    /^[[:space:]]*(#|$)/ { next }
    skip_deeper >= 0 {
      if (indent_of($0) > skip_deeper) next
      skip_deeper = -1
    }
    /^ *build:$/ {
      skip_deeper = indent_of($0)
      lines[++count] = substr($0, 1, skip_deeper) "image: <IMAGE>"
      next
    }
    /^ *image: ghcr\.io\/nicograef\/jotti-[a-z-]+:RELEASE_VERSION$/ {
      sub(/image: .*/, "image: <IMAGE>")
    }
    /^ *- \.\/database\/migrations:\/migrations$/ { next }
    { lines[++count] = $0 }
    END {
      # A service-level volumes key left without entries carried only the migrations mount.
      for (i = 1; i <= count; i++) {
        if (lines[i] ~ /^    volumes:$/ && (i == count || indent_of(lines[i + 1]) <= 4)) continue
        print lines[i]
      }
    }
  ' skip_deeper=-1 "$1"
}

if ! drift="$(diff <(normalize docker-compose.local.yml) <(normalize docker-compose.release.yml))"; then
  error "docker-compose.release.yml has drifted from docker-compose.local.yml (< local, > release):"
  printf '%s\n' "$drift" >&2
  exit 1
fi
info "docker-compose.release.yml matches docker-compose.local.yml."
