.DEFAULT_GOAL := help

.PHONY: init dev dev-up down restart logs status \
       test test-frontend test-integration test-all \
       lint-backend lint-backend-full lint-frontend lint \
       fmt-backend fmt-frontend fmt \
       build-backend build-relay build-resolver build-local-proxy build-frontend build \
       build-starter-windows build-relay-windows starter-syso release-windows \
       sqlc \
       prod-init prod-up prod-update prod-down prod-logs prod-backup prod-restore prod-harden \
       rocks-init rocks-up rocks-down rocks-logs rocks-reset-db rocks-reset-and-seed \
       local-up local-down local-logs local-reset-db local-reset-and-seed \
       db-shell seed rebuild-projections \
       clean \
       check-tools check-backend check-relay check-starter check-resolver check-local-proxy check-frontend check-integration check check-full verify \
       website website-check website-fmt \
       website-dev website-build website-verify \
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
	./scripts/test-integration.sh

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

# Version-String fuer die Windows-Exes (per ldflags einkompiliert). Der
# Release-Workflow ruft die Targets mit VERSION=<tag> auf.
VERSION ?= dev

# Verzeichnis-/Dateinamen des Release-ZIPs (dist/ ist gitignored).
RELEASE_NAME := jotti-windows-$(VERSION)
RELEASE_DIR := dist/$(RELEASE_NAME)

build-backend: ## Backend kompilieren
	cd backend && go build ./...

build-relay: ## Print-Relay-Binary kompilieren
	cd windows/relay && go build ./...

build-resolver: ## DNS-Resolver-Binary kompilieren
	cd resolver && go build ./...

build-local-proxy: ## Lokales Proxy-Entrypoint-Binary kompilieren
	cd reverse-proxy && go build ./...

build-starter-windows: ## Windows-Starter (jotti-start.exe) cross-kompilieren (VERSION=… fuer die Versionszeile)
	cd windows/starter && GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o jotti-start.exe .

build-relay-windows: ## Windows-Relay (jotti-relay.exe) cross-kompilieren (VERSION=… fuer die Versionszeile)
	cd windows/relay && GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o jotti-relay.exe .

starter-syso: ## Windows-Manifest (rsrc_windows_amd64.syso) aus jotti-start.manifest neu erzeugen (selten noetig)
	cd windows/starter && go run github.com/akavel/rsrc@v0.10.2 -manifest jotti-start.manifest -arch amd64 -o rsrc_windows_amd64.syso

release-windows: build-starter-windows build-relay-windows ## Release-ZIP (Exes + Release-Compose + Doku) unter dist/ bauen (VERSION=… setzen; baut KEINE Images)
	rm -rf "$(RELEASE_DIR)"
	mkdir -p "$(RELEASE_DIR)"
	cp windows/starter/jotti-start.exe "$(RELEASE_DIR)/"
	cp windows/relay/jotti-relay.exe "$(RELEASE_DIR)/"
	cp packaging/windows/jotti-stop.cmd "$(RELEASE_DIR)/"
	cp packaging/windows/jotti-restore.cmd "$(RELEASE_DIR)/"
	cp packaging/windows/jotti-repair.cmd "$(RELEASE_DIR)/"
	cp packaging/windows/KURZANLEITUNG.md "$(RELEASE_DIR)/"
	cp .env.example "$(RELEASE_DIR)/"
	cp docker-compose.release.yml "$(RELEASE_DIR)/"
	# Migrationen werden NICHT mehr ins ZIP kopiert — sie sind ins
	# jotti-migrate-Image gebacken (siehe database/migrate/Dockerfile).
	# Image-Tag im gestageten Compose auf die konkrete Version pinnen (die
	# eingecheckte Datei bleibt ein Template mit RELEASE_VERSION-Platzhalter).
	sed -i 's|:RELEASE_VERSION|:$(VERSION)|g' "$(RELEASE_DIR)/docker-compose.release.yml"
	cd dist && zip -qr "$(RELEASE_NAME).zip" "$(RELEASE_NAME)"
	@echo "Release-ZIP erstellt: dist/$(RELEASE_NAME).zip"

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

prod-init: ## Ersteinrichtung Produktion (.env prüfen, Images ziehen, Caddy Auto-TLS, Stack)
	./scripts/prod-init.sh

prod-up: ## Produktions-Stack starten/aktualisieren (zieht gepinnte Images, kein Build)
	docker compose -f docker-compose.prod.yml pull
	docker compose -f docker-compose.prod.yml up -d

prod-update: ## Sicheres Update (Pre-Update-Backup, Images ziehen, Migrationen, Health-Check, Rollback-Anleitung)
	./scripts/prod-update.sh

prod-down: ## Produktions-Stack stoppen
	docker compose -f docker-compose.prod.yml down

prod-logs: ## Produktions-Stack Logs folgen
	docker compose -f docker-compose.prod.yml logs -f

prod-backup: ## Datenbank-Backup ziehen (pg_dump, gzip, rotiert BACKUP_KEEP)
	./scripts/prod-backup.sh

prod-restore: ## Datenbank aus Backup wiederherstellen (destruktiv, mit Bestätigung)
	./scripts/prod-restore.sh

prod-harden: ## Optionale Server-Härtung (ufw-Firewall, fail2ban) — opt-in, idempotent
	./scripts/prod-harden.sh

# ──────────────────────────────────────────────
# jotti.rocks Deployment
# ──────────────────────────────────────────────

rocks-init: ## jotti.rocks Ersteinrichtung (Zertifikate für alle Domains, Stack)
	./scripts/rocks-init.sh

rocks-up: ## jotti.rocks Stack starten/aktualisieren (Landing + Demo App, inkl. nginx-Config)
	docker compose -f docker-compose.rocks.yml up -d --build
	docker compose -f docker-compose.rocks.yml up -d --no-deps --force-recreate reverse-proxy

rocks-down: ## jotti.rocks Stack stoppen
	docker compose -f docker-compose.rocks.yml down

rocks-logs: ## jotti.rocks Stack Logs folgen
	docker compose -f docker-compose.rocks.yml logs -f

rocks-reset-db: ## jotti.rocks-DB zurücksetzen (Zertifikate bleiben erhalten) — nur Demo/Staging
	docker compose -f docker-compose.rocks.yml down
	docker volume rm jotti_postgres-data
	docker compose -f docker-compose.rocks.yml up -d --build

rocks-reset-and-seed: ## jotti.rocks-DB resetten + Seed einspielen (SSL bleibt erhalten) — nur Demo/Staging
	./scripts/reset-and-seed.sh rocks --yes

# ──────────────────────────────────────────────
# Lokaler Betrieb (LAN, HTTPS via Caddy)
# ──────────────────────────────────────────────

local-up: ## Lokalen LAN-Stack starten/aktualisieren (HTTPS via lokal.jotti.rocks + interner CA-Fallback) — siehe docs/leitfaden.md
	@LAN_IP="$$(ip route get 1.1.1.1 2>/dev/null | awk '{for (i = 1; i <= NF; i++) if ($$i == "src") { print $$(i + 1); exit }}')"; \
	echo "Host-LAN-IP: $${LAN_IP:-<nicht erkannt>}"; \
	LAN_IP="$$LAN_IP" docker compose -f docker-compose.local.yml up -d --build; \
	LAN_IP="$$LAN_IP" docker compose -f docker-compose.local.yml up -d --no-deps --force-recreate reverse-proxy; \
	echo "Status & Zugangsadresse: http://localhost:8484"

local-down: ## Lokalen LAN-Stack stoppen
	docker compose -f docker-compose.local.yml down

local-logs: ## Lokalen LAN-Stack Logs folgen
	docker compose -f docker-compose.local.yml logs -f

local-reset-db: ## Lokale DB zurücksetzen (Caddy-Zertifikate bleiben erhalten)
	docker compose -f docker-compose.local.yml down
	docker volume rm jotti-local_postgres-data
	docker compose -f docker-compose.local.yml up -d --build

local-reset-and-seed: ## Lokale DB resetten + Seed einspielen (Caddy-Zertifikate bleiben erhalten)
	./scripts/reset-and-seed.sh local --yes

# ──────────────────────────────────────────────
# Datenbank                                     
# ──────────────────────────────────────────────

db-shell: ## psql-Shell im Dev-Postgres öffnen
	docker exec -it jotti-postgres-dev psql -U $${POSTGRES_USER:-admin} -d jotti

BACKEND_CONTAINER ?= jotti-backend-dev
seed: ## Demo-Daten per Seeder-Subkommando einspielen (Guard + Projektions-Rebuild inklusive)
	docker exec $(BACKEND_CONTAINER) go run ./main.go seed

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
	cd windows/relay && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-starter: ## Windows-Starter komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd windows/starter && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -count=1 -race ./... && go build ./...

check-resolver: ## DNS-Resolver komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd resolver && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-local-proxy: ## Lokales Proxy-Entrypoint-Binary komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd reverse-proxy && go mod tidy -diff && golangci-lint run && if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then echo "Go files are not properly formatted:"; goimports -l .; exit 1; fi && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-frontend: ## Frontend komplett prüfen (Format, Lint, Test, Build)
	cd frontend && pnpm format:check && pnpm lint && pnpm test && pnpm build

check-integration: ## Integrationstests gegen echte Datenbank ausführen
	./scripts/test-integration.sh

check: check-tools check-backend check-relay check-starter check-resolver check-local-proxy check-frontend ## Schnelle Komplettprüfung ohne DB-Integration

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

# head-seo.html ist für Prettier kein valides HTML (SSI-Echo in Attributwerten)
# und wird deshalb von Format-Check und -Write ausgenommen.
website-check: ## Website prüfen (Links, Assets, SSI, CSS-Klassen, Format)
	./scripts/check-website.sh
	cd frontend && pnpm exec prettier --check --ignore-unknown ../website '!../website/partials/head-seo.html'

website-fmt: ## Website formatieren (Prettier via Frontend-Installation)
	cd frontend && pnpm exec prettier --write --ignore-unknown ../website '!../website/partials/head-seo.html'

# Neue Astro/Starlight-Website (website/). Setzt einmaliges `cd website && pnpm install`
# voraus. Ersetzt in der letzten Phase die statischen Targets oben.
website-dev: ## Astro Dev-Server starten (http://localhost:4321), liest docs/ live
	cd website && pnpm dev

website-build: ## Website bauen (Astro Build → website/dist)
	cd website && pnpm build

website-verify: ## Website prüfen (astro check + Build inkl. Link-Validierung)
	cd website && pnpm check

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
