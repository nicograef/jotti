#!/bin/bash

echo "🧪 Starting integration test environment..."

# Start PostgreSQL container
echo "🐘 Starting PostgreSQL..."
docker run -d \
  --name jotti-postgres-test \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=jotti \
  -p 5432:5432 \
  --health-cmd "pg_isready -U admin -d jotti" \
  --health-interval 2s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:17

# Wait for PostgreSQL to be healthy
echo "⏳ Waiting for PostgreSQL to be ready..."
until docker exec jotti-postgres-test pg_isready -U admin -d jotti > /dev/null 2>&1; do
  sleep 2
done

echo "✅ PostgreSQL ready!"
sleep 2

# Run database migrations
echo "🔄 Running database migrations..."
cd database
chmod +x migrate/migrate || true
./migrate/migrate -path ./migrations -database "postgres://admin:admin@localhost:5432/jotti?sslmode=disable" up
cd ..

echo "✅ Migrations complete!"
echo ""
sleep 2

echo "🏃 Running integration tests..."

# Run integration tests
cd backend
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
JWT_SECRET=test-secret \
go test -tags=integration -count=1 -race ./... || true

echo ""
echo "🧹 Cleaning up..."
cd ..

# Run migrations down
cd database
./migrate/migrate -path ./migrations -database "postgres://admin:admin@localhost:5432/jotti?sslmode=disable" down -all || true
cd ..

# Stop and remove container
docker stop jotti-postgres-test > /dev/null 2>&1
docker rm jotti-postgres-test > /dev/null 2>&1

