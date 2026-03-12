.PHONY: dev dev-up down restart logs status \
       test test-integration \
       lint-backend lint-frontend lint \
       fmt-backend fmt-frontend fmt \
       build-backend build-frontend build \
       sqlc \
       staging staging-down staging-logs \
       prod-up prod-down prod-reset-db init \
       db-shell \
       clean \
       help

# ──────────────────────────────────────────────
# Development                                   
# ──────────────────────────────────────────────

dev: ## Dev-Stack starten (docker compose, detached)
	docker compose -f docker-compose.dev.yml up --build -d

dev-up: ## Dev-Stack starten (Vordergrund, mit Logs)
	docker compose -f docker-compose.dev.yml up --build

down: ## Dev-Stack stoppen
	docker compose -f docker-compose.dev.yml down

restart: down dev ## Dev-Stack neu starten

logs: ## Dev-Stack Logs folgen
	docker compose -f docker-compose.dev.yml logs -f

status: ## Status aller Dev-Container anzeigen
	docker compose -f docker-compose.dev.yml ps

# ──────────────────────────────────────────────
# Tests                                         
# ──────────────────────────────────────────────

test: ## Backend Unit-Tests ausführen
	cd backend && go test -tags=unit -race ./...

test-frontend: ## Frontend Tests ausführen
	cd frontend && pnpm test

test-integration: ## Integrationstests ausführen
	./test-integration.sh

test-all: test test-frontend ## Alle Unit-Tests (Backend + Frontend)

# ──────────────────────────────────────────────
# Linting                                       
# ──────────────────────────────────────────────

lint-backend: ## Backend Linting (go vet + goimports)
	cd backend && go vet ./... && goimports -l .

lint-backend-full: ## Backend Linting mit golangci-lint
	cd backend && golangci-lint run

lint-frontend: ## Frontend Linting (ESLint)
	cd frontend && pnpm lint

lint: lint-backend lint-frontend ## Backend + Frontend Linting

# ──────────────────────────────────────────────
# Formatierung                                  
# ──────────────────────────────────────────────

fmt-backend: ## Backend Code formatieren (goimports)
	cd backend && goimports -w .

fmt-frontend: ## Frontend Code formatieren (Prettier)
	cd frontend && pnpm format

fmt: fmt-backend fmt-frontend ## Backend + Frontend formatieren

# ──────────────────────────────────────────────
# Build                                         
# ──────────────────────────────────────────────

build-backend: ## Backend kompilieren
	cd backend && go build ./...

build-frontend: ## Frontend kompilieren
	cd frontend && pnpm build

build: build-backend build-frontend ## Backend + Frontend kompilieren

# ──────────────────────────────────────────────
# Code-Generierung                              
# ──────────────────────────────────────────────

sqlc: ## sqlc Code generieren (aus SQL-Queries)
	cd backend && sqlc generate

# ──────────────────────────────────────────────
# Staging                                       
# ──────────────────────────────────────────────

staging: ## Staging-Stack starten (Vordergrund)
	docker compose -f docker-compose.staging.yml up --build

staging-down: ## Staging-Stack stoppen
	docker compose -f docker-compose.staging.yml down

staging-logs: ## Staging-Stack Logs folgen
	docker compose -f docker-compose.staging.yml logs -f

# ──────────────────────────────────────────────
# Produktion                                    
# ──────────────────────────────────────────────

prod-up: ## Produktions-Stack starten
	docker compose up -d --build

prod-down: ## Produktions-Stack stoppen
	docker compose down

prod-reset-db: ## Prod-DB zurücksetzen (Zertifikate bleiben erhalten)
	docker compose down
	docker volume rm $$(docker volume ls -q --filter name=_postgres-data | head -1)
	docker compose up -d --build

init: ## Ersteinrichtung Produktion (Zertifikate, Stack)
	./scripts/prod-init.sh

# ──────────────────────────────────────────────
# Datenbank                                     
# ──────────────────────────────────────────────

db-shell: ## psql-Shell im Dev-Postgres öffnen
	docker exec -it jotti-postgres-dev psql -U $${POSTGRES_USER:-admin} -d jotti

# ──────────────────────────────────────────────
# Aufräumen                                     
# ──────────────────────────────────────────────

clean: down ## Dev-Stack stoppen und Volumes entfernen
	docker compose -f docker-compose.dev.yml down -v

# ──────────────────────────────────────────────
# Qualitätsprüfung (CI-nah)                     
# ──────────────────────────────────────────────

check-backend: ## Backend komplett prüfen (tidy, lint, test, build)
	cd backend && go mod tidy && golangci-lint run && goimports -w . && go vet ./... && go test -tags=unit -count=1 ./... && go build ./...

check-frontend: ## Frontend komplett prüfen (format, lint, build)
	cd frontend && pnpm format && pnpm lint && pnpm build

check: check-backend check-frontend ## Alles prüfen (Backend + Frontend)

# ──────────────────────────────────────────────
# Hilfe                                         
# ──────────────────────────────────────────────

help: ## Alle verfügbaren Targets anzeigen
	@echo ""
	@echo "Verfügbare Make-Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk -F ':.*## ' '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
