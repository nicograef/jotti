#!/usr/bin/env bash
set -euo pipefail

# jotti local prerequisite setup for quality gates (make check / make verify)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

ensure_cmd() {
  local cmd="$1"
  local hint="$2"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    fatal "Missing required command '$cmd'. $hint"
  fi
}

# CI reference (.github/workflows/ci.yml):
# - Go: 1.26.0
# - Node: 24
# - pnpm: 10
# - golangci-lint: latest (action)

info "Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

info "Checking base runtimes..."
ensure_cmd go "Install Go >= 1.26.0 (CI uses 1.26.0)."
ensure_cmd node "Install Node >= 24 (CI uses 24)."

info "Ensuring goimports is available..."
if command -v goimports >/dev/null 2>&1; then
  info "goimports already installed: $(goimports -V 2>/dev/null || echo 'version unknown')"
else
  info "Installing goimports via 'go install golang.org/x/tools/cmd/goimports@latest'"
  go install golang.org/x/tools/cmd/goimports@latest
fi

GO_BIN_PATH="$(go env GOPATH)/bin"
export PATH="$GO_BIN_PATH:$PATH"
if ! command -v goimports >/dev/null 2>&1; then
  fatal "goimports is still not on PATH. Add '$GO_BIN_PATH' to your PATH and rerun this script."
fi

info "Ensuring golangci-lint is available..."
if command -v golangci-lint >/dev/null 2>&1; then
  info "golangci-lint already installed: $(golangci-lint --version | head -n 1)"
else
  ensure_cmd curl "Install curl to bootstrap golangci-lint."
  info "Installing golangci-lint with official install script"
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b "$GO_BIN_PATH" latest
fi

if ! command -v golangci-lint >/dev/null 2>&1; then
  fatal "golangci-lint installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

info "Ensuring sqlc is available..."
if command -v sqlc >/dev/null 2>&1; then
  info "sqlc already installed: $(sqlc version)"
else
  info "Installing sqlc via 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest'"
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
fi

if ! command -v sqlc >/dev/null 2>&1; then
  fatal "sqlc installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

# Matches CI: .github/workflows/ci.yml (Install golang-migrate step)
MIGRATE_VERSION="v4.19.1"
info "Ensuring golang-migrate ($MIGRATE_VERSION) is available..."
if command -v migrate >/dev/null 2>&1; then
  info "golang-migrate already installed: $(migrate -version 2>&1 || echo 'version unknown')"
else
  ensure_cmd curl "Install curl to bootstrap golang-migrate."
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) fatal "Unsupported architecture: $ARCH" ;;
  esac
  MIGRATE_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${OS}-${ARCH}.tar.gz"
  info "Downloading golang-migrate from $MIGRATE_URL"
  curl -fsSL "$MIGRATE_URL" | tar -xz -C "$GO_BIN_PATH" migrate
fi

if ! command -v migrate >/dev/null 2>&1; then
  fatal "golang-migrate installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

info "Ensuring pnpm (v10) is available..."
if command -v pnpm >/dev/null 2>&1; then
  info "pnpm already installed: $(pnpm --version)"
else
  if command -v corepack >/dev/null 2>&1; then
    info "Activating pnpm@10 via corepack"
    corepack enable
    corepack prepare pnpm@10 --activate
  else
    fatal "pnpm not found and corepack is unavailable. Install pnpm v10 manually."
  fi
fi

if ! command -v pnpm >/dev/null 2>&1; then
  fatal "pnpm installation failed. Install pnpm v10 manually and rerun."
fi

info "Installing frontend dependencies..."
cd "$PROJECT_ROOT/frontend" && pnpm install
cd "$PROJECT_ROOT"

info "Tool summary"
echo "  go:             $(go version)"
echo "  node:           $(node --version)"
echo "  pnpm:           $(pnpm --version)"
echo "  goimports:      $(goimports -V 2>/dev/null || echo 'installed')"
echo "  golangci-lint:  $(golangci-lint --version | head -n 1)"
echo "  sqlc:           $(sqlc version)"
echo "  migrate:        $(migrate -version 2>&1 || echo 'installed')"

info "All verify-relevant tools are available."
info "Next step: make verify"
