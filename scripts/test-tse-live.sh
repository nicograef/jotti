#!/bin/bash

# Faehrt die TSE-Live-Suite: Wegwerf-Postgres hochziehen, Migrationen
# einspielen, .env.fiskaly-test laden und die Live-Tests (Live-Guard, echte
# fiskaly-TEST-TSS) ausfuehren, danach aufraeumen. Ohne .env.fiskaly-test bricht
# das Skript ab; ohne Credentials skippen die Tests selbst (Guard im Testcode).
#
# Eigener Container/Port, damit die Suite parallel zur normalen
# Integrationstest-DB (scripts/test-integration.sh, Port 5432) laufen kann.
#
# ACHTUNG: legt KEINE TSS an. Der TSS-anlegende Setup-Durchlauf lebt
# ausschliesslich im separaten Opt-in-Target `make test-tse-live-setup`.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTAINER_NAME="jotti-pg-tse-live"
PGPORT=5453
DATABASE_URL="postgres://admin:admin@localhost:${PGPORT}/jotti?sslmode=disable"

if [ ! -f "$ROOT_DIR/.env.fiskaly-test" ]; then
  echo "FEHLER: .env.fiskaly-test fehlt. Vorlage: .env.fiskaly-test.example"
  exit 1
fi

cleanup() {
  echo ""
  echo "🧹 Cleaning up..."

  docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
  docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "🧪 Starting TSE live test environment..."

docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

echo "🐘 Starting PostgreSQL..."
docker run -d \
  --name "$CONTAINER_NAME" \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=jotti \
  -p ${PGPORT}:5432 \
  --health-cmd "pg_isready -U admin -d jotti" \
  --health-interval 2s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:17

echo "⏳ Waiting for PostgreSQL to accept real connections..."
# pg_isready alone is not enough: during initialization postgres:17 runs a
# temporary socket-only server that is restarted afterwards; pg_isready reports
# that server as ready while migrate then fails with "connection reset by
# peer". Only a real query over TCP proves the final server is up.
until docker exec -e PGPASSWORD=admin "$CONTAINER_NAME" \
  psql -h 127.0.0.1 -U admin -d jotti -c "SELECT 1" >/dev/null 2>&1; do
  sleep 2
done

echo "✅ PostgreSQL ready!"

echo "🔄 Running database migrations (forward-only)..."
migrate -path "$ROOT_DIR/database/migrations" -database "$DATABASE_URL" up

echo "✅ Migrations complete!"
echo ""

echo "🏃 Running TSE live tests against the fiskaly TEST-TSS..."
cd "$ROOT_DIR/backend"
# .env.fiskaly-test liefert die FISKALY_TEST_*-Credentials und den Admin-PUK/PIN;
# nie ausgeben oder loggen. Ohne Credentials skippen die Tests (Guard im Code).
set -a
# shellcheck disable=SC1091
. "$ROOT_DIR/.env.fiskaly-test"
set +a

# -run 'LiveSigniert|LiveSuite' schliesst den TSS-anlegenden Setup-Durchlauf
# (TestFiskalySetup_LiveVollerDurchlauf) bewusst aus: Der bleibt allein im
# separaten Target `make test-tse-live-setup`.
POSTGRES_HOST=localhost \
POSTGRES_PORT=${PGPORT} \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
go test -tags=integration -count=1 -v -p 1 -run 'LiveSigniert|LiveSuite' ./api/fiskal/tse_live/ ./repository/tse_repo/

echo "✅ TSE live tests passed!"
