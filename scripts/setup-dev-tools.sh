#!/usr/bin/env bash
# setup-dev-tools.sh — installs the pinned tools behind make check / make verify (idempotent)
#
# Usage:
#   bash scripts/setup-dev-tools.sh
#
# What it does:
#   1. Requires Go and Node.
#   2. Installs golangci-lint, sqlc, golang-migrate and actionlint at their pins via
#      `go install`, and shellcheck via apt-get.
#   3. Installs pnpm at the packageManager pin via npm, then every package's
#      dependencies.
set -euo pipefail

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

info "Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

info "Checking base runtimes..."
ensure_cmd go "Install Go >= 1.27.1 (CI uses 1.27.1)."
ensure_cmd node "Install Node >= 24 (CI uses 24)."

GO_BIN_PATH="$(go env GOPATH)/bin"
export PATH="$GO_BIN_PATH:$PATH"

# Every Go tool below is built with `go install` via the module proxy: GitHub
# release downloads are blocked behind some proxies.
#
# golangci-lint is built with the module's own toolchain (backend/go.mod): it
# refuses to run when the Go it was built with is older than the module's.
GO_TOOLCHAIN="$(cd "$PROJECT_ROOT/backend" && go env GOVERSION)"

# Version of the module an installed Go tool was built from, read from its
# build info (`go version -m`) rather than a tool flag: a `go install`ed
# migrate reports "dev". Prints nothing when the command is missing or was
# built from another module.
installed_mod_version() {
  local bin buildinfo
  bin="$(command -v "$1")" || return 0
  buildinfo="$(go version -m "$bin" 2>/dev/null)" || return 0
  awk -v module="$2" '$1 == "mod" && $2 == module {print $3}' <<<"$buildinfo"
}

# Matches CI: .github/workflows/ci.yml pins the golangci-lint action to this
# version so a green CI and a green `make verify` mean the same thing.
GOLANGCI_LINT_VERSION="v2.14.0"
info "Ensuring golangci-lint ($GOLANGCI_LINT_VERSION) is available..."

# A version match alone is not enough for golangci-lint: compare the Go version
# recorded in the binary (`go version -m`) against $GO_TOOLCHAIN too.
golangci_lint_built_with() {
  go version -m "$1" 2>/dev/null | awk 'NR==1 {print $2}'
}

INSTALLED_GOLANGCI=""
INSTALLED_GOLANGCI_BUILT_WITH=""
if command -v golangci-lint >/dev/null 2>&1; then
  INSTALLED_GOLANGCI="v$(golangci-lint version --short 2>/dev/null || echo 'unknown')"
  INSTALLED_GOLANGCI_BUILT_WITH="$(golangci_lint_built_with "$(command -v golangci-lint)")"
fi

if [[ "$INSTALLED_GOLANGCI" = "$GOLANGCI_LINT_VERSION" ]] && [[ "$INSTALLED_GOLANGCI_BUILT_WITH" = "$GO_TOOLCHAIN" ]]; then
  info "golangci-lint already installed: $INSTALLED_GOLANGCI (built with $INSTALLED_GOLANGCI_BUILT_WITH)"
else
  if [[ -n "$INSTALLED_GOLANGCI" ]] && [[ "$INSTALLED_GOLANGCI" != "$GOLANGCI_LINT_VERSION" ]]; then
    info "Replacing golangci-lint $INSTALLED_GOLANGCI with the pinned $GOLANGCI_LINT_VERSION"
  elif [[ -n "$INSTALLED_GOLANGCI" ]]; then
    info "Rebuilding golangci-lint $GOLANGCI_LINT_VERSION: built with $INSTALLED_GOLANGCI_BUILT_WITH, module now targets $GO_TOOLCHAIN"
  fi
  info "Building golangci-lint $GOLANGCI_LINT_VERSION with $GO_TOOLCHAIN into $GO_BIN_PATH"
  GOTOOLCHAIN="$GO_TOOLCHAIN" GOBIN="$GO_BIN_PATH" \
    go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_LINT_VERSION"
fi

if ! command -v golangci-lint >/dev/null 2>&1; then
  fatal "golangci-lint installation failed. Ensure '$GO_BIN_PATH' is on PATH (before any system golangci-lint) and rerun."
fi

# Pinned to the version that generated the checked-in backend/sqlc/dbgen/ (see
# the "versions:" header in those files) and to the sqlc diff step in
# .github/workflows/ci.yml; a different sqlc can reformat the generated code
# and make `make sqlc` or `make check-sqlc` fail.
SQLC_VERSION="v1.31.1"
info "Ensuring sqlc ($SQLC_VERSION) is available..."
INSTALLED_SQLC="$(installed_mod_version sqlc github.com/sqlc-dev/sqlc)"
if [[ "$INSTALLED_SQLC" = "$SQLC_VERSION" ]]; then
  info "sqlc already installed: $INSTALLED_SQLC"
else
  info "Installing sqlc $SQLC_VERSION into $GO_BIN_PATH (installed: ${INSTALLED_SQLC:-none})"
  GOBIN="$GO_BIN_PATH" go install "github.com/sqlc-dev/sqlc/cmd/sqlc@$SQLC_VERSION"
fi

if ! command -v sqlc >/dev/null 2>&1; then
  fatal "sqlc installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

# Matches MIGRATE_VERSION in database/migrate/Dockerfile, which CI runs. The
# postgres build tag adds the one database driver jotti needs; the file source
# behind `-path` is always built in.
MIGRATE_VERSION="v4.20.1"
info "Ensuring golang-migrate ($MIGRATE_VERSION) is available..."
INSTALLED_MIGRATE="$(installed_mod_version migrate github.com/golang-migrate/migrate/v4)"
if [[ "$INSTALLED_MIGRATE" = "$MIGRATE_VERSION" ]]; then
  info "golang-migrate already installed: $INSTALLED_MIGRATE"
else
  info "Installing golang-migrate $MIGRATE_VERSION into $GO_BIN_PATH (installed: ${INSTALLED_MIGRATE:-none})"
  GOBIN="$GO_BIN_PATH" go install -tags postgres "github.com/golang-migrate/migrate/v4/cmd/migrate@$MIGRATE_VERSION"
fi

if ! command -v migrate >/dev/null 2>&1; then
  fatal "golang-migrate installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

# CI runs the shellcheck of the GitHub runner image, which is not pinned either,
# so the distribution package is close enough; apt covers Debian, Ubuntu and
# the devcontainer, other systems get the hint.
info "Ensuring shellcheck is available..."
if ! command -v shellcheck >/dev/null 2>&1 && command -v apt-get >/dev/null 2>&1; then
  sudo_cmd=()
  if [[ "$(id -u)" -ne 0 ]] && command -v sudo >/dev/null 2>&1; then
    sudo_cmd=(sudo)
  fi
  info "Installing shellcheck via apt-get"
  if ! { "${sudo_cmd[@]+"${sudo_cmd[@]}"}" apt-get update -qq &&
    "${sudo_cmd[@]+"${sudo_cmd[@]}"}" apt-get install -y -qq shellcheck; }; then
    warn "apt-get could not install shellcheck."
  fi
fi
ensure_cmd shellcheck "Install shellcheck with your package manager (apt-get install shellcheck, brew install shellcheck)."

# Matches the "Lint workflows" step of .github/workflows/ci.yml.
ACTIONLINT_VERSION="v1.7.12"
info "Ensuring actionlint ($ACTIONLINT_VERSION) is available..."
INSTALLED_ACTIONLINT="$(installed_mod_version actionlint github.com/rhysd/actionlint)"
if [[ "$INSTALLED_ACTIONLINT" = "$ACTIONLINT_VERSION" ]]; then
  info "actionlint already installed: $INSTALLED_ACTIONLINT"
else
  info "Installing actionlint $ACTIONLINT_VERSION into $GO_BIN_PATH (installed: ${INSTALLED_ACTIONLINT:-none})"
  GOBIN="$GO_BIN_PATH" go install "github.com/rhysd/actionlint/cmd/actionlint@$ACTIONLINT_VERSION"
fi

if ! command -v actionlint >/dev/null 2>&1; then
  fatal "actionlint installation failed. Ensure '$GO_BIN_PATH' is on PATH and rerun."
fi

# npm, not Corepack: Node 25+ no longer ships Corepack. The version is the
# packageManager pin that scripts/check-pins.sh keeps equal across packages.
PNPM_VERSION="$(sed -n 's/.*"packageManager": *"pnpm@\([^+"]*\).*/\1/p' "$PROJECT_ROOT/frontend/package.json")"
[[ -n "$PNPM_VERSION" ]] || fatal "No pnpm packageManager pin in frontend/package.json."
info "Ensuring pnpm ($PNPM_VERSION) is available..."
if command -v pnpm >/dev/null 2>&1; then
  info "pnpm already installed: $(pnpm --version)"
else
  info "Installing pnpm $PNPM_VERSION with npm"
  npm install -g "pnpm@$PNPM_VERSION"
  hash -r
fi

if ! command -v pnpm >/dev/null 2>&1; then
  fatal "pnpm is not on PATH after 'npm install -g'. Add '$(npm prefix -g)/bin' to your PATH and rerun."
fi

for project in frontend website e2e; do
  info "Installing $project dependencies..."
  (cd "$PROJECT_ROOT/$project" && pnpm install --frozen-lockfile)
done

info "Tool summary"
echo "  go:             $(go version)"
echo "  node:           $(node --version)"
echo "  pnpm:           $(pnpm --version)"
echo "  golangci-lint:  $(golangci-lint --version | head -n 1)"
echo "  sqlc:           $(sqlc version)"
echo "  migrate:        $(installed_mod_version migrate github.com/golang-migrate/migrate/v4)"
echo "  shellcheck:     $(shellcheck --version | awk '/^version:/ {print $2}')"
echo "  actionlint:     $(installed_mod_version actionlint github.com/rhysd/actionlint)"

info "All verify-relevant tools are available."
info "Next step: make verify"
