#!/usr/bin/env bash
set -euo pipefail

# jotti — one pinned version per third-party image, across every stack and every
# test harness. Two stacks on different versions of the same image behave
# differently while claiming to be the same deployment, and a harness that
# starts another version tests something no stack ever runs.
#
# Scanned are the tracked places that pin a version:
#   - `image:` lines in docker-compose*.yml
#   - `FROM` lines in every Dockerfile
#   - the `packageManager` fields of frontend, website and e2e
# jotti's own images (ghcr.io/nicograef/jotti-*) are exempt: their tag is a
# variable on purpose, so every stack follows the release it was shipped with.
#
# Compared is, per image name, the version part of the tag up to the first "-":
# caddy:2.11.4-builder and caddy:2.11.4 are the same version. Two versions for
# one name are an error unless scripts/check-pins.allow names the image with a
# reason; an entry that matches nothing turns the gate red and is to be deleted.
# The packageManager fields must match literally, sha512 hash included.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

cd "$PROJECT_ROOT"

ALLOWLIST="scripts/check-pins.allow"
OWN_IMAGE_PREFIX="ghcr.io/nicograef/jotti-"
PACKAGE_JSONS=(frontend/package.json website/package.json e2e/package.json)

mapfile -t allow_entries < <(
  [ -f "$ALLOWLIST" ] && grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST"
)

allow_names=()
allow_hits=()
for entry in "${allow_entries[@]+"${allow_entries[@]}"}"; do
  name="$(printf '%s\n' "$entry" | awk '{print $1}')"
  reason="$(printf '%s\n' "$entry" | awk '{print $2}')"
  # Without this an image name carrying a space would be cut at that space by
  # the field split and exempt a shorter name than the entry spells out.
  if [ "${reason:0:1}" != "#" ]; then
    fatal "$ALLOWLIST: image name with a space, or reason missing: $entry"
  fi
  allow_names+=("$name")
  allow_hits+=(0)
done

# `:(glob)` makes `**/` mean "zero or more directories", so a Dockerfile in the
# repository root is matched as well.
mapfile -t compose_files < <(git ls-files ':(glob)docker-compose*.yml')
mapfile -t dockerfiles < <(git ls-files ':(glob)**/Dockerfile*')

# collect_pins prints one "<image reference><TAB><file>:<line>" per pinned image.
collect_pins() {
  local file
  for file in "${compose_files[@]+"${compose_files[@]}"}"; do
    awk -v file="$file" '
      match($0, /^[[:space:]]*image:[[:space:]]*/) {
        ref = substr($0, RSTART + RLENGTH)
        sub(/[[:space:]#].*$/, "", ref)
        gsub(/["'\'']/, "", ref)
        if (ref != "") print ref "\t" file ":" FNR
      }
    ' "$file"
  done
  for file in "${dockerfiles[@]+"${dockerfiles[@]}"}"; do
    awk -v file="$file" '
      toupper($1) == "FROM" {
        ref = $2
        if (ref ~ /^--/) ref = $3
        gsub(/["'\'']/, "", ref)
        if (ref != "") print ref "\t" file ":" FNR
      }
    ' "$file"
  done
}

declare -A version_count=()
declare -A version_seen=()
declare -A version_detail=()
violations=0

while IFS=$'\t' read -r ref loc; do
  case "$ref" in
    "$OWN_IMAGE_PREFIX"*) continue ;;
  esac

  name="${ref%:*}"
  tag="${ref##*:}"
  if [ "$name" = "$ref" ]; then
    error "$loc: image without a pinned tag: $ref"
    violations=$((violations + 1))
    continue
  fi

  version="${tag%%-*}"
  key="$name"$'\t'"$version"
  if [ -z "${version_seen[$key]:-}" ]; then
    version_seen["$key"]=1
    version_count["$name"]=$(( ${version_count[$name]:-0} + 1 ))
    version_detail["$name"]="${version_detail[$name]:-}"$'\n'"  $version at $loc"
  fi
done < <(collect_pins)

# Sorted, so the report is the same on every run.
mapfile -t names < <(printf '%s\n' "${!version_count[@]}" | sort)

for name in "${names[@]+"${names[@]}"}"; do
  [ "${version_count[$name]}" -le 1 ] && continue

  exempt=0
  if [ "${#allow_names[@]}" -gt 0 ]; then
    for i in "${!allow_names[@]}"; do
      if [ "$name" = "${allow_names[$i]}" ]; then
        allow_hits[i]=$((allow_hits[i] + 1))
        exempt=1
        break
      fi
    done
  fi
  [ "$exempt" -eq 1 ] && continue

  error "$name is pinned to ${version_count[$name]} versions (allow it in $ALLOWLIST):"
  # The detail lines are collected in the order the sources were read, so the
  # report names the first location of every version exactly once.
  while IFS= read -r line; do
    [ -n "$line" ] && error "$line"
  done <<<"${version_detail[$name]}"
  violations=$((violations + 1))
done

if [ "${#allow_names[@]}" -gt 0 ]; then
  for i in "${!allow_names[@]}"; do
    if [ "${allow_hits[$i]}" -eq 0 ]; then
      error "$ALLOWLIST: exception matches nothing any more: ${allow_names[$i]}"
      violations=$((violations + 1))
    fi
  done
fi

# packageManager: one literal value in all three package.json. A missing file is
# a rename that must turn the gate red, not silently shrink its scope.
pm_files=()
pm_values=()
for file in "${PACKAGE_JSONS[@]}"; do
  if [ ! -f "$file" ]; then
    fatal "$file is missing: the packageManager comparison needs all three package.json."
  fi
  value="$(awk '
    match($0, /"packageManager"[[:space:]]*:[[:space:]]*"[^"]*"/) {
      field = substr($0, RSTART, RLENGTH)
      sub(/^"packageManager"[[:space:]]*:[[:space:]]*"/, "", field)
      sub(/"$/, "", field)
      print field
      exit
    }
  ' "$file")"
  if [ -z "$value" ]; then
    error "$file: no packageManager field."
    violations=$((violations + 1))
    continue
  fi
  pm_files+=("$file")
  pm_values+=("$value")
done

if [ "${#pm_values[@]}" -gt 1 ]; then
  for i in "${!pm_values[@]}"; do
    if [ "${pm_values[$i]}" != "${pm_values[0]}" ]; then
      error "${pm_files[$i]}: packageManager is ${pm_values[$i]}, ${pm_files[0]} pins ${pm_values[0]}."
      violations=$((violations + 1))
    fi
  done
fi

if [ "$violations" -gt 0 ]; then
  fatal "$violations version pin violation(s) found."
fi

info "Every third-party image and all three packageManager fields carry one pinned version."
