#!/usr/bin/env bash
set -euo pipefail

# jotti — one pinned version per third-party image across the stacks: two stacks
# on different versions of the same image behave differently while claiming to be
# the same deployment. Read are `image:` lines in docker-compose*.yml and
# .github/workflows/*.yml, `FROM` lines in every Dockerfile, literal references
# to one of those image names in scripts/*.sh and windows/**/*.go (the
# `docker run` calls of the test scripts, the starter's helper image) and the
# packageManager fields of frontend, website and e2e. jotti's own images (ghcr.io/nicograef/jotti-*) are
# exempt: their tag is a variable on purpose, so every stack follows the release
# it was shipped with. Compared per image name is the tag up to the first "-", so
# caddy:2.11.4-builder and caddy:2.11.4 are one version; a digest pin counts as
# its own version and collides with a tag pin of the same image.

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
mapfile -t workflow_files < <(git ls-files ':(glob).github/workflows/*.yml')
# This script names example images in its comments, so it does not scan itself.
mapfile -t literal_files < <(
  git ls-files ':(glob)scripts/*.sh' ':(glob)windows/**/*.go' | grep -vxF 'scripts/check-pins.sh'
)

# collect_pins prints one "<image reference><TAB><file>:<line>" per pinned image.
collect_pins() {
  local file
  for file in "${compose_files[@]+"${compose_files[@]}"}" "${workflow_files[@]+"${workflow_files[@]}"}"; do
    awk -v file="$file" '
      match($0, /^[[:space:]]*image:[[:space:]]*/) {
        ref = substr($0, RSTART + RLENGTH)
        sub(/[[:space:]#].*$/, "", ref)
        gsub(/["'\'']/, "", ref)
        if (ref != "") print ref "\t" file ":" FNR
      }
    ' "$file"
  done
  # One awk process per file, so `stages` never leaks into the next Dockerfile.
  for file in "${dockerfiles[@]+"${dockerfiles[@]}"}"; do
    awk -v file="$file" '
      toupper($1) == "FROM" {
        ref = $2
        if (ref ~ /^--/) ref = $3
        gsub(/["'"'"']/, "", ref)
        # Not a pinned third-party image: the empty base image, a ref built from
        # a variable, or an earlier build stage of this same file.
        skip = (ref == "" || tolower(ref) == "scratch" || ref ~ /\$/ || tolower(ref) in stages)
        for (i = 3; i < NF; i++) {
          if (toupper($i) == "AS") stages[tolower($(i + 1))] = 1
        }
        if (!skip) print ref "\t" file ":" FNR
      }
    ' "$file"
  done
}

# image_name prints a reference without its digest and tag.
image_name() {
  local base="${1%%@*}" segment
  segment="${base##*/}"
  case "$segment" in
    *:*) printf '%s\n' "${base%:"${segment##*:}"}" ;;
    *) printf '%s\n' "$base" ;;
  esac
}

# collect_literal_pins prints the same lines for references in literal_files to
# an image name that collect_pins found ($1: those names as one ERE
# alternation). A match is the name, not preceded by a character that could
# extend it or by the "@" of a URL host (postgres://...@postgres:5432), plus a
# tag that starts with a digit, optionally after "v".
collect_literal_pins() {
  local file
  [ -n "$1" ] || return 0
  for file in "${literal_files[@]+"${literal_files[@]}"}"; do
    NAMES_RE="$1" awk -v file="$file" '
      match($0, "(^|[^A-Za-z0-9./_@-])(" ENVIRON["NAMES_RE"] "):v?[0-9][A-Za-z0-9._-]*") {
        ref = substr($0, RSTART, RLENGTH)
        sub(/^[^A-Za-z0-9]/, "", ref)
        print ref "\t" file ":" FNR
      }
    ' "$file"
  done
}

mapfile -t pins < <(collect_pins)
names_re="$(
  for line in "${pins[@]+"${pins[@]}"}"; do
    ref="${line%%$'\t'*}"
    case "$ref" in
      "$OWN_IMAGE_PREFIX"*) continue ;;
    esac
    image_name "$ref"
  done | sort -u | sed 's/[.]/\\./g' | paste -sd '|' -
)"

declare -A version_count=()
declare -A version_seen=()
declare -A version_detail=()
violations=0

while IFS=$'\t' read -r ref loc; do
  case "$ref" in
    "$OWN_IMAGE_PREFIX"*) continue ;;
  esac

  # A digest pin carries its version after the "@"; what precedes it may still
  # carry a tag. Split the tag off the last path segment only, so the port of a
  # registry host (`registry:5000/image`) is not mistaken for one.
  digest=""
  base="$ref"
  case "$ref" in
    *@*)
      digest="${ref#*@}"
      base="${ref%%@*}"
      ;;
  esac

  segment="${base##*/}"
  case "$segment" in
    *:*)
      tag="${segment##*:}"
      name="${base%:"$tag"}"
      ;;
    *)
      tag=""
      name="$base"
      ;;
  esac

  if [ -n "$digest" ]; then
    version="$digest"
  elif [ -z "$tag" ]; then
    error "$loc: image without a pinned tag: $ref"
    violations=$((violations + 1))
    continue
  else
    version="${tag%%-*}"
  fi
  key="$name"$'\t'"$version"
  if [ -z "${version_seen[$key]:-}" ]; then
    version_seen["$key"]=1
    version_count["$name"]=$(( ${version_count[$name]:-0} + 1 ))
    version_detail["$name"]="${version_detail[$name]:-}"$'\n'"  $version at $loc"
  fi
done < <(
  if [ "${#pins[@]}" -gt 0 ]; then
    printf '%s\n' "${pins[@]}"
  fi
  collect_literal_pins "$names_re"
)

# Sorted, so the report is the same on every run. The guard matters: printf
# without arguments would emit one empty line and turn it into an empty name.
names=()
if [ "${#version_count[@]}" -gt 0 ]; then
  mapfile -t names < <(printf '%s\n' "${!version_count[@]}" | sort)
fi

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
