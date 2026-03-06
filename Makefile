.PHONY: dev down logs test test-integration lint-backend lint-frontend lint prod-up prod-down

dev:
	docker compose -f docker-compose.dev.yml up --build -d

down:
	docker compose -f docker-compose.dev.yml down

logs:
	docker compose -f docker-compose.dev.yml logs -f

test:
	cd backend && go test -tags=unit -race ./...

test-integration:
	./test-integration.sh

lint-backend:
	cd backend && go vet ./... && goimports -l .

lint-frontend:
	cd frontend && pnpm lint

lint: lint-backend lint-frontend

prod-up:
	docker compose up -d --build

prod-down:
	docker compose down
