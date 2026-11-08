#!/usr/bin/env bash
# einzeldatei-smoke.sh — smoke test of jotti.exe on a Windows host (Git Bash)
#
# Usage:
#   bash scripts/einzeldatei-smoke.sh start <jotti.exe>
#
# Modes:
#   start   starts the file with a pipe on stdin, waits until https://127.0.0.1/api/health
#           answers 200, https://127.0.0.1/ serves the app's index.html (both -k: internal CA)
#           and the file prompts for Enter, then sends Enter and requires exit 0 with no
#           postgres.exe left. An Enter sent during the start must not end the file.
#
# Environment:
#   EINZELDATEI_TIMEOUT   seconds to wait for health and app, default 180
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

TIMEOUT="${EINZELDATEI_TIMEOUT:-180}"

usage() {
  fatal "Usage: $0 start <jotti.exe>"
}

# postgres_running succeeds while any postgres.exe process exists.
postgres_running() {
  tasklist //FI "IMAGENAME eq postgres.exe" //NH 2>/dev/null | grep -qi 'postgres.exe'
}

# app_ready LOG succeeds once health answers 200, / serves the app's index.html and LOG
# shows the Enter prompt.
app_ready() {
  curl -kfsS -o /dev/null --max-time 5 https://127.0.0.1/api/health 2>/dev/null \
    && curl -kfsS --max-time 5 https://127.0.0.1/ 2>/dev/null | grep -q '<title>jotti</title>' \
    && grep -q 'Enter beendet jotti' "$1"
}

# start_exe EXE LOG RELEASE — runs EXE with stdin fed by one Enter during the start, which EXE
# must ignore, and one more as soon as the file RELEASE exists; sets EXE_PID to the pipeline's
# last process.
start_exe() {
  local exe="$1" log="$2" release="$3"
  (
    echo
    while [[ ! -e "$release" ]]; do sleep 1; done
    echo
  ) | "$exe" >"$log" 2>&1 &
  EXE_PID=$!
}

mode_start() {
  local exe="$1" work log release status=0 waited=0
  [[ -f "$exe" ]] || fatal "$exe not found."
  work="$(mktemp -d)"
  log="$work/jotti.log"
  release="$work/enter"

  info "Starting $exe ..."
  start_exe "$exe" "$log" "$release"

  until app_ready "$log"; do
    if ! kill -0 "$EXE_PID" 2>/dev/null; then
      touch "$release"
      cat "$log" >&2
      fatal "$exe ended before it served the app."
    fi
    if [[ "$waited" -ge "$TIMEOUT" ]]; then
      # A file stuck before its Enter prompt would never read the release.
      touch "$release"
      kill "$EXE_PID" 2>/dev/null || true
      wait "$EXE_PID" || true
      cat "$log" >&2
      fatal "$exe did not serve health and app within ${TIMEOUT}s."
    fi
    sleep 1
    waited=$((waited + 1))
  done
  info "Health and app answer after ${waited}s."
  sleep 3
  if ! kill -0 "$EXE_PID" 2>/dev/null; then
    cat "$log" >&2
    fatal "$exe ended on the Enter sent during its start."
  fi

  touch "$release"
  wait "$EXE_PID" || status=$?
  cat "$log" >&2
  [[ "$status" -eq 0 ]] || fatal "$exe exited with $status after Enter."
  if postgres_running; then
    fatal "postgres.exe still runs after $exe ended."
  fi
  info "$exe ended cleanly; no postgres.exe left."
}

[[ $# -ge 1 ]] || usage
case "$1" in
  start)
    [[ $# -eq 2 ]] || usage
    mode_start "$2"
    ;;
  *) usage ;;
esac
