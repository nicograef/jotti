#!/usr/bin/env bash
set -euo pipefail

# jotti local prerequisite setup for quality gates (make check / make verify)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

ensure_cmd() {
  local cmd="$1"
  local hint="$2"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    fatal "Missing required command '$cmd'. $hint"
  fi
}

# CI reference (.github/workflows/ci.yml):
# - Go: 1.26.5
# - Node: 24
# - pnpm: 11
# - golangci-lint: pinned, see GOLANGCI_LINT_VERSION below

info "Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

info "Checking base runtimes..."
ensure_cmd go "Install Go >= 1.26.5 (CI uses 1.26.5)."
ensure_cmd node "Install Node >= 24 (CI uses 24)."

# Matches CI: .github/workflows/ci.yml pins goimports to this version in every
# "Check format" step, so local formatting matches CI (D13). goimports across
# versions can reformat imports differently, so @latest would drift from CI.
GOIMPORTS_VERSION="v0.40.0"
info "Ensuring goimports ($GOIMPORTS_VERSION) is available..."
if command -v goimports >/dev/null 2>&1; then
  info "goimports already installed: $(goimports -V 2>/dev/null || echo 'version unknown')"
else
  info "Installing goimports via 'go install golang.org/x/tools/cmd/goimports@$GOIMPORTS_VERSION'"
  go install "golang.org/x/tools/cmd/goimports@$GOIMPORTS_VERSION"
fi

GO_BIN_PATH="$(go env GOPATH)/bin"
export PATH="$GO_BIN_PATH:$PATH"
if ! command -v goimports >/dev/null 2>&1; then
  fatal "goimports is still not on PATH. Add '$GO_BIN_PATH' to your PATH and rerun this script."
fi

# Matches CI: .github/workflows/ci.yml pins the golangci-lint action to this
# version so a green CI and a green `make verify` mean the same thing (D13).
GOLANGCI_LINT_VERSION="v2.11.4"
info "Ensuring golangci-lint ($GOLANGCI_LINT_VERSION) is available..."

# golangci-lint refuses to run when the Go it was built with is older than the
# version the module targets (backend/go.mod), so it must be built with the
# module's own toolchain. Prebuilt GitHub release binaries are additionally
# unreachable through the cloud-session proxy, so `go install` via the
# (allowlisted) module proxy is the one method that works locally and in cloud.
GO_TOOLCHAIN="$(cd "$PROJECT_ROOT/backend" && go env GOVERSION)"

INSTALLED_GOLANGCI=""
if command -v golangci-lint >/dev/null 2>&1; then
  INSTALLED_GOLANGCI="v$(golangci-lint version --short 2>/dev/null || echo 'unknown')"
fi

if [ "$INSTALLED_GOLANGCI" = "$GOLANGCI_LINT_VERSION" ]; then
  info "golangci-lint already installed: $INSTALLED_GOLANGCI"
else
  [ -n "$INSTALLED_GOLANGCI" ] && info "Replacing golangci-lint $INSTALLED_GOLANGCI with the pinned $GOLANGCI_LINT_VERSION"
  info "Building golangci-lint $GOLANGCI_LINT_VERSION with $GO_TOOLCHAIN into $GO_BIN_PATH"
  GOTOOLCHAIN="$GO_TOOLCHAIN" GOBIN="$GO_BIN_PATH" \
    go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION"
fi

# Cloud sessions ship an older golangci-lint at /usr/local/bin — on the default
# PATH, ahead of "$GO_BIN_PATH" — that would shadow the pinned build in `make`.
# The container is ephemeral, so point that copy at the pinned build too.
if [ "${CLAUDE_CODE_REMOTE:-}" = "true" ] && [ -w /usr/local/bin/golangci-lint ] \
   && [ "$(/usr/local/bin/golangci-lint version --short 2>/dev/null)" != "${GOLANGCI_LINT_VERSION#v}" ]; then
  info "Cloud session: replacing the base-image golangci-lint at /usr/local/bin with $GOLANGCI_LINT_VERSION"
  cp "$GO_BIN_PATH/golangci-lint" /usr/local/bin/golangci-lint
fi

if ! command -v golangci-lint >/dev/null 2>&1; then
  fatal "golangci-lint installation failed. Ensure '$GO_BIN_PATH' is on PATH (before any system golangci-lint) and rerun."
fi

# Pinned to the version that generated the checked-in backend/sqlc/dbgen/ (see
# the "versions:" header in those files); a different sqlc can reformat the
# generated code and make `make sqlc` dirty the working tree.
SQLC_VERSION="v1.31.1"
info "Ensuring sqlc ($SQLC_VERSION) is available..."
if command -v sqlc >/dev/null 2>&1; then
  info "sqlc already installed: $(sqlc version)"
else
  info "Installing sqlc via 'go install github.com/sqlc-dev/sqlc/cmd/sqlc@$SQLC_VERSION'"
  go install "github.com/sqlc-dev/sqlc/cmd/sqlc@$SQLC_VERSION"
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

info "Ensuring pnpm (v11) is available..."
if command -v pnpm >/dev/null 2>&1; then
  info "pnpm already installed: $(pnpm --version)"
else
  if command -v corepack >/dev/null 2>&1; then
    info "Activating pnpm@11 via corepack"
    corepack enable
    corepack prepare pnpm@11 --activate
  else
    fatal "pnpm not found and corepack is unavailable. Install pnpm v11 manually."
  fi
fi

if ! command -v pnpm >/dev/null 2>&1; then
  fatal "pnpm installation failed. Install pnpm v11 manually and rerun."
fi

info "Installing frontend dependencies..."
cd "$PROJECT_ROOT/frontend" && pnpm install
cd "$PROJECT_ROOT"

info "Tool summary"
echo "  go:             $(go version)"
echo "  node:           $(node --version)"
echo "  pnpm:           $(pnpm --version)"
echo "  goimports:      $(goimports -V 2>/dev/null || echo 'version unknown')"
echo "  golangci-lint:  $(golangci-lint --version | head -n 1)"
echo "  sqlc:           $(sqlc version)"
echo "  migrate:        $(migrate -version 2>&1 || echo 'version unknown')"

info "All verify-relevant tools are available."
info "Next step: make verify"
