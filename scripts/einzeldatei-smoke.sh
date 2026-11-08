#!/usr/bin/env bash
# einzeldatei-smoke.sh — smoke test and measurement of jotti.exe on a Windows host (Git Bash)
#
# Usage:
#   bash scripts/einzeldatei-smoke.sh start <jotti.exe>
#   bash scripts/einzeldatei-smoke.sh messung <jotti.exe> <tsv>
#   bash scripts/einzeldatei-smoke.sh auswerten <tsv>
#
# Modes:
#   start      starts the file with a pipe on stdin, waits until https://127.0.0.1/api/health
#              answers 200, https://127.0.0.1/ serves the app's index.html (both -k: internal CA)
#              and the file prompts for Enter, then sends Enter and requires exit 0 with no
#              postgres.exe left. An Enter sent during the start must not end the file.
#   messung    requires a host jotti never ran on, runs `start` twice and appends erststart_s
#              and folgestart_s (process start until health and app answer) and groesse_bytes
#              to <tsv>, one "<key>\t<value>" line each.
#   auswerten  prints the values of <tsv> and fails when one is missing or over its threshold:
#              groesse_bytes 120000000, erststart_s 60, folgestart_s 15,
#              defender_echtzeitschutz not true, defender_funde over 0.
#
# Environment:
#   EINZELDATEI_TIMEOUT   seconds to wait for health, app and the Enter prompt, default 180
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

TIMEOUT="${EINZELDATEI_TIMEOUT:-180}"

usage() {
  fatal "Usage: $0 start <jotti.exe> | messung <jotti.exe> <tsv> | auswerten <tsv>"
}

# now_s prints the wall clock in seconds with milliseconds.
now_s() {
  date +%s.%3N
}

# since_s START prints the seconds since START (from now_s) with one decimal.
since_s() {
  awk -v s="$1" -v n="$(now_s)" 'BEGIN { printf "%.1f", n - s }'
}

# postgres_running succeeds while any postgres.exe process exists.
postgres_running() {
  tasklist //FI "IMAGENAME eq postgres.exe" //NH 2>/dev/null | grep -qi 'postgres.exe'
}

# app_serves succeeds once health answers 200 and / serves the app's index.html.
app_serves() {
  curl -kfsS -o /dev/null --max-time 5 https://127.0.0.1/api/health 2>/dev/null \
    && curl -kfsS --max-time 5 https://127.0.0.1/ 2>/dev/null | grep -q '<title>jotti</title>'
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

# give_up LOG RELEASE MESSAGE — ends a file that did not get ready and fails with its log.
give_up() {
  local log="$1" release="$2" message="$3"
  # A file stuck before its Enter prompt would never read the release.
  touch "$release"
  kill "$EXE_PID" 2>/dev/null || true
  wait "$EXE_PID" || true
  cat "$log" >&2
  fatal "$message"
}

# keep_waiting EXE LOG RELEASE STARTED WHAT — sleeps half a second; gives up when EXE ended or
# TIMEOUT passed since STARTED before it WHAT.
keep_waiting() {
  local exe="$1" log="$2" release="$3" started="$4" what="$5"
  kill -0 "$EXE_PID" 2>/dev/null || give_up "$log" "$release" "$exe ended before it $what."
  if awk -v s="$started" -v n="$(now_s)" -v t="$TIMEOUT" 'BEGIN { exit !(n - s >= t) }'; then
    give_up "$log" "$release" "$exe had not $what after ${TIMEOUT}s."
  fi
  sleep 0.5
}

# run_once EXE — one start-serve-stop cycle as described for `start`; sets SERVED_S to the
# seconds from process start until health and app answered.
run_once() {
  local exe="$1" work log release started status=0
  [[ -f "$exe" ]] || fatal "$exe not found."
  work="$(mktemp -d)"
  log="$work/jotti.log"
  release="$work/enter"

  info "Starting $exe ..."
  started="$(now_s)"
  start_exe "$exe" "$log" "$release"

  until app_serves; do
    keep_waiting "$exe" "$log" "$release" "$started" 'served health and app'
  done
  SERVED_S="$(since_s "$started")"
  info "Health and app answer after ${SERVED_S}s."
  until grep -q 'Enter beendet jotti' "$log"; do
    keep_waiting "$exe" "$log" "$release" "$started" 'prompted for Enter'
  done
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

mode_messung() {
  local exe="$1" tsv="$2" dir erststart
  [[ -f "$exe" ]] || fatal "$exe not found."
  # Folder IDs 35 and 38 are %PROGRAMDATA% and %ProgramFiles%; an earlier run there would turn
  # the first start into a later one.
  for dir in "$(cygpath -F 35)/jotti" "$(cygpath -F 38)/jotti"; do
    [[ ! -e "$dir" ]] || fatal "$dir exists; messung needs a host jotti never ran on."
  done

  run_once "$exe"
  erststart="$SERVED_S"
  run_once "$exe"
  {
    printf 'erststart_s\t%s\n' "$erststart"
    printf 'folgestart_s\t%s\n' "$SERVED_S"
    printf 'groesse_bytes\t%s\n' "$(wc -c <"$exe" | tr -d '[:space:]')"
  } >>"$tsv"
  info "First start ${erststart}s, later start ${SERVED_S}s."
}

mode_auswerten() {
  local tsv="$1"
  [[ -f "$tsv" ]] || fatal "$tsv not found."
  # A line written by PowerShell may end in \r.
  awk -F'\t' '
    function fail(msg) { print "Threshold failed: " msg > "/dev/stderr"; bad = 1 }
    function over(key, max) {
      if (!(key in v) || v[key] !~ /^[0-9]+(\.[0-9]+)?$/) fail(key " is missing or not a number")
      else if (v[key] + 0 > max) fail(key " " v[key] " exceeds " max)
    }
    { sub(/\r$/, "", $2); v[$1] = $2; printf "%s\t%s\n", $1, $2 }
    END {
      over("groesse_bytes", 120000000)
      over("erststart_s", 60)
      over("folgestart_s", 15)
      over("defender_funde", 0)
      if (v["defender_echtzeitschutz"] != "true") fail("defender_echtzeitschutz is \"" v["defender_echtzeitschutz"] "\", not true")
      exit bad
    }' "$tsv"
}

[[ $# -ge 1 ]] || usage
case "$1" in
  start)
    [[ $# -eq 2 ]] || usage
    run_once "$2"
    ;;
  messung)
    [[ $# -eq 3 ]] || usage
    mode_messung "$2" "$3"
    ;;
  auswerten)
    [[ $# -eq 2 ]] || usage
    mode_auswerten "$2"
    ;;
  *) usage ;;
esac
