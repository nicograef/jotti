.DEFAULT_GOAL := help

.PHONY: init dev dev-up down restart logs status \
	test test-frontend test-integration test-all \
       lint-backend lint-backend-full lint-frontend lint \
       fmt-backend fmt-frontend fmt \
       build-backend build-frontend build \
       sqlc \
       prod-init prod-up prod-down prod-logs prod-reset-db prod-reset-and-seed \
       jotti-rocks-init jotti-rocks-up jotti-rocks-down jotti-rocks-logs \
       local-up local-down local-logs \
       db-shell seed rebuild-projections \
       clean \
	check-tools check-backend check-relay check-frontend check-integration check check-full verify \
       website \
       help

# ──────────────────────────────────────────────
# Development                                   
# ──────────────────────────────────────────────

init: ## .env erzeugen (idempotent, sichere Secrets)
	./scripts/init-env.sh

dev: ## Dev-Stack starten (docker compose, detached)
	docker compose up --build -d

dev-up: ## Dev-Stack starten (Vordergrund, mit Logs)
	docker compose up --build

down: ## Dev-Stack stoppen
	docker compose down

restart: down dev ## Dev-Stack neu starten

logs: ## Dev-Stack Logs folgen
	docker compose logs -f

status: ## Status aller Dev-Container anzeigen
	docker compose ps

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

build-relay: ## Print-Relay-Binary kompilieren
	cd cmd/relay && go build ./...

build-frontend: ## Frontend kompilieren
	cd frontend && pnpm build

build: build-backend build-frontend ## Backend + Frontend kompilieren

# ──────────────────────────────────────────────
# Code-Generierung                              
# ──────────────────────────────────────────────

sqlc: ## sqlc Code generieren (aus SQL-Queries)
	cd backend && sqlc generate

# ──────────────────────────────────────────────
# Produktion                                    
# ──────────────────────────────────────────────

prod-init: ## Ersteinrichtung Produktion (Zertifikate, Stack)
	./scripts/prod-init.sh

prod-up: ## Produktions-Stack starten/aktualisieren (wendet nginx-Config-Änderungen an)
	docker compose -f docker-compose.prod.yml up -d --build
	docker compose -f docker-compose.prod.yml up -d --no-deps --force-recreate reverse-proxy

prod-down: ## Produktions-Stack stoppen
	docker compose -f docker-compose.prod.yml down

prod-logs: ## Produktions-Stack Logs folgen
	docker compose -f docker-compose.prod.yml logs -f

prod-reset-db: ## Prod-DB zurücksetzen (Zertifikate bleiben erhalten)
	docker compose -f docker-compose.prod.yml down
	docker volume rm $$(docker volume ls -q --filter name=_postgres-data | head -1)
	docker compose -f docker-compose.prod.yml up -d --build

prod-reset-and-seed: ## Prod-DB resetten, Seed importieren, Projektionen neu aufbauen (SSL bleibt erhalten)
	./scripts/prod-reset-and-seed.sh --yes

# ──────────────────────────────────────────────
# jotti.rocks Deployment                        
# ──────────────────────────────────────────────

jotti-rocks-init: ## jotti.rocks Ersteinrichtung (Zertifikate für alle Domains, Stack)
	./scripts/jotti-rocks-init.sh

jotti-rocks-up: ## jotti.rocks Stack starten/aktualisieren (Landing + Demo App, inkl. nginx-Config)
	docker compose -f docker-compose.prod.yml -f docker-compose.jotti-rocks.yml up -d --build
	docker compose -f docker-compose.prod.yml -f docker-compose.jotti-rocks.yml up -d --no-deps --force-recreate reverse-proxy

jotti-rocks-down: ## jotti.rocks Stack stoppen
	docker compose -f docker-compose.prod.yml -f docker-compose.jotti-rocks.yml down

jotti-rocks-logs: ## jotti.rocks Stack Logs folgen
	docker compose -f docker-compose.prod.yml -f docker-compose.jotti-rocks.yml logs -f

# ──────────────────────────────────────────────
# Lokaler Betrieb (LAN, HTTPS selbstsigniert)   
# ──────────────────────────────────────────────

local-up: ## Lokalen LAN-Stack starten/aktualisieren (HTTPS, selbstsigniertes Zertifikat) — siehe docs/betrieb/leitfaden-hosting.md
	docker compose -f docker-compose.local.yml up -d --build
	docker compose -f docker-compose.local.yml up -d --no-deps --force-recreate reverse-proxy

local-down: ## Lokalen LAN-Stack stoppen
	docker compose -f docker-compose.local.yml down

local-logs: ## Lokalen LAN-Stack Logs folgen
	docker compose -f docker-compose.local.yml logs -f

# ──────────────────────────────────────────────
# Datenbank                                     
# ──────────────────────────────────────────────

db-shell: ## psql-Shell im Dev-Postgres öffnen
	docker exec -it jotti-postgres-dev psql -U $${POSTGRES_USER:-admin} -d jotti

PG_CONTAINER ?= jotti-postgres-dev
BACKEND_CONTAINER ?= jotti-backend-dev
seed: ## Demo-Daten einspielen und Projektionen aufbauen
	docker exec -i $(PG_CONTAINER) psql -U $${POSTGRES_USER:-admin} -d jotti < database/seed.sql
	@$(MAKE) rebuild-projections

rebuild-projections: ## table_state-Projektionen aus Events neu aufbauen
	docker exec $(BACKEND_CONTAINER) go run ./main.go rebuild-projections

# ──────────────────────────────────────────────
# Aufräumen                                     
# ──────────────────────────────────────────────

clean: down ## Dev-Stack stoppen und Volumes entfernen
	docker compose down -v

# ──────────────────────────────────────────────
# Qualitätsprüfung (CI-nah)                     
# ──────────────────────────────────────────────

check-tools: ## Prüfen, ob lokale Verify-Tools installiert sind
	@for tool in golangci-lint goimports pnpm; do \
		if ! command -v $$tool >/dev/null 2>&1; then \
			echo "Fehlendes Tool: $$tool"; \
			echo "Installiere es mit scripts/setup-dev-tools.sh oder folge der README-Anleitung."; \
			exit 1; \
		fi; \
	done

check-backend: ## Backend komplett prüfen (Deps, Format, Lint, Test, Build)
	cd backend && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -tags=unit -count=1 -race ./... && go build ./...

check-relay: ## Print-Relay komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd cmd/relay && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-frontend: ## Frontend komplett prüfen (Format, Lint, Test, Build)
	cd frontend && pnpm format:check && pnpm lint && pnpm test && pnpm build

check-integration: ## Integrationstests gegen echte Datenbank ausführen
	./test-integration.sh

check: check-tools check-backend check-relay check-frontend ## Schnelle Komplettprüfung ohne DB-Integration

check-full: check check-integration ## Vollständige Prüfung inkl. Integrationstests

verify: check-tools check-full ## Alias für vollständige Repo-Prüfung

# ──────────────────────────────────────────────
# Website                                       
# ──────────────────────────────────────────────

website: ## Website Dev-Server (nginx + SSI) starten (http://localhost:8080)
	docker run --rm -p 8080:80 \
	  -v $(CURDIR)/website:/usr/share/nginx/html:ro \
	  -v $(CURDIR)/reverse-proxy/nginx.website-dev.conf:/etc/nginx/conf.d/default.conf:ro \
	  nginx:1.27-alpine

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
