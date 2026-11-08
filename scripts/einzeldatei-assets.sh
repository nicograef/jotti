#!/usr/bin/env bash
# einzeldatei-assets.sh — fills windows/einzeldatei/assets/files/ for the jotti.exe build
#
# Usage:
#   make build-einzeldatei-windows VERSION=v1.2.3   # or: VERSION=v1.2.3 ./scripts/einzeldatei-assets.sh
#
# Environment:
#   VERSION                  version baked into the app build, default dev
#   EINZELDATEI_CACHE_DIR    where the Postgres ZIP is cached, default ~/.cache/jotti
#
# What it does:
#   1. Downloads the EDB Postgres ZIP for Windows once and checks its SHA-256.
#   2. Packs the allowlisted binaries, DLLs and share/ into files/postgres.zip.
#   3. Builds the app with VERSION into files/app/ and copies the migrations into files/migrations/.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# check-pins.sh holds this version against the postgres: image tag.
POSTGRES_WINDOWS_VERSION="17.11"
POSTGRES_WINDOWS_BUILD="1"
POSTGRES_WINDOWS_SHA256="6eabdf00d2893713b75db4336a23c3fdf505f056e217ec6e2e95d901750cfea3"

# Executables the single file runs, and the DLLs they and the two server libraries import.
# plpgsql and dict_snowball are the libraries initdb loads; everything else in lib\ is an extension.
POSTGRES_ALLOWLIST=(
  bin/postgres.exe bin/pg_ctl.exe bin/initdb.exe bin/pg_dump.exe bin/pg_restore.exe
  bin/icudt67.dll bin/icuin67.dll bin/icuuc67.dll
  bin/libcrypto-3-x64.dll bin/libssl-3-x64.dll bin/libpq.dll
  bin/libiconv-2.dll bin/libintl-9.dll bin/libwinpthread-1.dll
  bin/liblz4.dll bin/libzstd.dll bin/libxml2.dll bin/zlib1.dll
  lib/plpgsql.dll lib/dict_snowball.dll
  'share/*'
)

VERSION="${VERSION:-dev}"
CACHE_DIR="${EINZELDATEI_CACHE_DIR:-$HOME/.cache/jotti}"
FILES_DIR="$PROJECT_ROOT/windows/einzeldatei/assets/files"
ZIP_NAME="postgresql-${POSTGRES_WINDOWS_VERSION}-${POSTGRES_WINDOWS_BUILD}-windows-x64-binaries.zip"
ZIP_URL="https://get.enterprisedb.com/postgresql/${ZIP_NAME}"

for tool in curl sha256sum unzip zip pnpm; do
  command -v "$tool" >/dev/null 2>&1 || fatal "Missing tool: $tool"
done

mkdir -p "$CACHE_DIR"
zip_path="$CACHE_DIR/$ZIP_NAME"
if ! echo "$POSTGRES_WINDOWS_SHA256  $zip_path" | sha256sum -c --status - 2>/dev/null; then
  info "Downloading $ZIP_URL ..."
  curl -fsSL -o "$zip_path.part" "$ZIP_URL"
  echo "$POSTGRES_WINDOWS_SHA256  $zip_path.part" | sha256sum -c --quiet - \
    || fatal "SHA-256 of $ZIP_NAME does not match the pin."
  mv "$zip_path.part" "$zip_path"
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

info "Packing the Postgres allowlist ..."
unzip -q "$zip_path" "${POSTGRES_ALLOWLIST[@]/#/pgsql/}" -d "$work"
find "$FILES_DIR" -mindepth 1 ! -name .gitkeep -delete
(cd "$work/pgsql" && zip -qr -9 "$FILES_DIR/postgres.zip" bin lib share)

info "Building the app ($VERSION) ..."
(cd "$PROJECT_ROOT/frontend" && pnpm install --frozen-lockfile --prefer-offline && VERSION="$VERSION" pnpm build)
cp -r "$PROJECT_ROOT/frontend/dist" "$FILES_DIR/app"

mkdir -p "$FILES_DIR/migrations"
cp "$PROJECT_ROOT"/database/migrations/*.up.sql "$FILES_DIR/migrations/"

info "Assets ready in $FILES_DIR."
