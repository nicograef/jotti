.DEFAULT_GOAL := help

.PHONY: init up up-attached down restart logs status \
       test test-frontend test-integration test-all test-e2e test-tse-live test-tse-live-setup fuzz \
       lint-backend lint-backend-full lint-frontend lint-e2e lint \
       fmt-backend fmt-frontend fmt-repo fmt \
       build-backend build-relay build-resolver build-local-proxy build-frontend build \
       build-starter-windows build-relay-windows starter-syso release-windows \
       sqlc \
       prod-init prod-up prod-update prod-down prod-logs prod-backup prod-restore prod-backup-verify prod-harden \
       rocks-init rocks-up rocks-down rocks-logs rocks-reset-db rocks-reset-and-seed rocks-backup \
       local-up local-down local-logs \
       db-shell seed rebuild-projections \
       clean \
       check-tools check-tools-integration check-backend check-sqlc check-relay check-starter check-resolver check-local-proxy check-frontend check-e2e check-shell check-workflows check-format check-repo check-integration check check-full verify \
       website-dev website-build website-test website-check website-screenshots \
       help

# Development

init: ## .env erzeugen (idempotent, sichere Secrets)
	./scripts/init-env.sh

up: ## Dev-Stack starten (docker compose, detached)
	docker compose up --build -d

up-attached: ## Dev-Stack starten (Vordergrund, mit Logs)
	docker compose up --build

down: ## Dev-Stack stoppen
	docker compose down

restart: down up ## Dev-Stack neu starten

logs: ## Dev-Stack Logs folgen
	docker compose logs -f

status: ## Status aller Dev-Container anzeigen
	docker compose ps

# Tests

test: ## Backend Unit-Tests ausführen
	cd backend && go test -race ./...

test-frontend: ## Frontend Tests ausführen
	cd frontend && pnpm test

test-integration: ## Integrationstests ausführen
	./scripts/test-integration.sh

test-tse-live: ## TSE-Live-Suite gegen die fiskaly-TEST-TSS (Wegwerf-Postgres + Migrationen + Live-Tests, legt keine TSS an)
	./scripts/test-tse-live.sh

test-tse-live-setup: ## Voller TSE-Setup-Durchlauf gegen fiskaly-TEST (ACHTUNG: legt eine unlöschbare TSS im TEST-Konto an)
	@test -f .env.fiskaly-test || { echo "FEHLER: .env.fiskaly-test fehlt. Vorlage: .env.fiskaly-test.example"; exit 1; }
	cd backend && set -a && . ../.env.fiskaly-test && set +a && JOTTI_TSE_LIVE=1 go test -tags=integration -count=1 -v -run 'LiveVollerDurchlauf' ./repository/tse_repo/

test-all: test test-frontend ## Alle Unit-Tests (Backend + Frontend)

test-e2e: ## E2E-Tests (Playwright) gegen E2E_BASE_URL (Default Dev-Stack http://localhost)
	cd e2e && pnpm install --frozen-lockfile && pnpm exec playwright install chromium && pnpm test

fuzz: ## Fuzz-Targets länger laufen lassen (je Target 90s; kein CI-Dauerlauf)
	cd backend && go test -run='^$$' -fuzz='FuzzApplyEvent$$' -fuzztime=90s ./domain/kasse/
	cd backend && go test -run='^$$' -fuzz='FuzzPositionEventDataRoundtrip$$' -fuzztime=90s ./domain/kasse/
	cd backend && go test -run='^$$' -fuzz='FuzzSerializeCSV$$' -fuzztime=90s ./api/fiskal/dsfinvk/
	cd backend && go test -run='^$$' -fuzz='FuzzFormatKassenbeleg$$' -fuzztime=90s ./api/druck/bondruck/application/escpos/

# Linting

# goimports is a tool of backend/go.mod; go.work makes it the one pin for every Go module.
# $(goimports-check) fails when goimports would rewrite a file below the current directory.
goimports-check = if [ -n "$$(go tool goimports -l .)" ]; then echo "Go files are not properly formatted:"; go tool goimports -l .; exit 1; fi

lint-backend: ## Backend Linting (go vet + goimports)
	cd backend && go vet ./... && $(goimports-check)

lint-backend-full: ## Backend Linting mit golangci-lint (inkl. Integrationstest-Dateien)
	cd backend && golangci-lint run

lint-frontend: ## Frontend Linting (ESLint)
	cd frontend && pnpm lint

lint-e2e: ## E2E-Suite Linting (ESLint)
	cd e2e && pnpm lint

lint: lint-backend lint-frontend lint-e2e check-shell ## Backend-, Frontend-, E2E- und Shell-Linting

# Formatierung

fmt-backend: ## Backend Code formatieren (goimports)
	cd backend && go tool goimports -w .

fmt-frontend: ## Frontend Code formatieren (Prettier + ESLint --fix)
	cd frontend && pnpm format && pnpm lint:fix

PRETTIER_GLOB := "**/*.{ts,tsx,js,mjs,cjs,json,css,md}"

fmt-repo: ## Repo-weite Prettier-Formatierung schreiben (Gegenstück zu check-format)
	frontend/node_modules/.bin/prettier --write $(PRETTIER_GLOB)

fmt: fmt-backend fmt-frontend fmt-repo ## Backend, Frontend und Repo-Prettier formatieren

# Build

# Version-String fuer die Windows-Exes (per ldflags einkompiliert). Der
# Release-Workflow ruft die Targets mit VERSION=<tag> auf.
VERSION ?= dev

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
	# Migrationen liegen im jotti-migrate-Image (database/migrate/Dockerfile),
	# nicht im ZIP.
	# Das eingecheckte Compose bleibt Template; nur die gestagete Kopie wird gepinnt.
	sed -i 's|:RELEASE_VERSION|:$(VERSION)|g' "$(RELEASE_DIR)/docker-compose.release.yml"
	cd dist && zip -qr "$(RELEASE_NAME).zip" "$(RELEASE_NAME)"
	@echo "Release-ZIP erstellt: dist/$(RELEASE_NAME).zip"

build-frontend: ## Frontend kompilieren
	cd frontend && pnpm build

build: build-backend build-frontend ## Backend + Frontend kompilieren

# Code-Generierung

sqlc: ## sqlc Code generieren (aus SQL-Queries)
	cd backend && sqlc generate

# Produktion

prod-init: ## Ersteinrichtung Produktion (.env prüfen, Images ziehen, Caddy Auto-TLS, Stack)
	./scripts/prod-init.sh

prod-up: ## Stack starten/neustarten mit gepinnter Version (kein Update; fuer Updates: make prod-update)
	@v=$$(grep -E '^JOTTI_VERSION=' .env 2>/dev/null | tail -n1 | cut -d= -f2- | tr -d '[:space:]'); \
	if ! echo "$$v" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+([.+-].*)?$$'; then \
	  echo "FEHLER: JOTTI_VERSION in .env ist kein gepinntes Release-Tag (gefunden: '$${v:-<leer>}')."; \
	  echo "Setze JOTTI_VERSION=vX.Y.Z in .env. Zum Aktualisieren: make prod-update."; \
	  exit 1; \
	fi
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

prod-backup-verify: ## Backup probeweise in Wegwerf-Postgres einspielen (prüft Wiederherstellbarkeit)
	./scripts/prod-backup-verify.sh

prod-harden: ## Optionale Server-Härtung (ufw, fail2ban, unattended-upgrades, SSH nur per Schlüssel) — opt-in, idempotent
	./scripts/prod-harden.sh

# jotti.rocks Deployment

rocks-init: ## jotti.rocks Ersteinrichtung (Stack bauen und starten, Zertifikate holt Caddy)
	./scripts/rocks-init.sh

rocks-up: ## jotti.rocks Stack starten/aktualisieren (Landing + Demo App, inkl. Caddyfile)
	docker compose -f docker-compose.rocks.yml up -d --build --remove-orphans
	docker compose -f docker-compose.rocks.yml up -d --no-deps --force-recreate reverse-proxy

rocks-down: ## jotti.rocks Stack stoppen
	docker compose -f docker-compose.rocks.yml down

rocks-logs: ## jotti.rocks Stack Logs folgen
	docker compose -f docker-compose.rocks.yml logs -f

rocks-reset-db: ## jotti.rocks-DB zurücksetzen (Zertifikate bleiben erhalten) — nur Demo/Staging
	docker compose -f docker-compose.rocks.yml down
	docker volume rm jotti_postgres-data
	docker compose -f docker-compose.rocks.yml up -d --build

rocks-reset-and-seed: ## jotti.rocks-DB resetten + Seed einspielen (Zertifikate bleiben erhalten) — nur Demo/Staging
	./scripts/reset-and-seed.sh rocks --yes

rocks-backup: ## acme-dns-Datenbank vom VPS sichern, vom Laptop aus (DEST=<Verzeichnis>)
	./scripts/rocks-backup.sh "$(DEST)"

# Lokaler Betrieb (LAN, HTTPS via Caddy)

local-up: ## Lokalen LAN-Stack starten/aktualisieren (HTTPS via lokal.jotti.rocks + interner CA-Fallback) — siehe docs/leitfaden/installation.md
	@LAN_IP="$$(ip route get 1.1.1.1 2>/dev/null | awk '{for (i = 1; i <= NF; i++) if ($$i == "src") { print $$(i + 1); exit }}')"; \
	echo "Host-LAN-IP: $${LAN_IP:-<nicht erkannt>}"; \
	LAN_IP="$$LAN_IP" docker compose -f docker-compose.local.yml up -d --build; \
	LAN_IP="$$LAN_IP" docker compose -f docker-compose.local.yml up -d --no-deps --force-recreate reverse-proxy; \
	echo "Status & Zugangsadresse: http://localhost:8484"

local-down: ## Lokalen LAN-Stack stoppen
	docker compose -f docker-compose.local.yml down

local-logs: ## Lokalen LAN-Stack Logs folgen
	docker compose -f docker-compose.local.yml logs -f

# Datenbank

db-shell: ## psql-Shell im Dev-Postgres öffnen
	docker exec -it jotti-postgres-dev psql -U $${POSTGRES_USER:-admin} -d jotti

BACKEND_CONTAINER ?= jotti-backend-dev
seed: ## Demo-Daten per Seeder-Subkommando einspielen (Guard + Projektions-Rebuild inklusive)
	docker exec $(BACKEND_CONTAINER) go run ./main.go seed

rebuild-projections: ## tisch_sessions-Projektionen aus Events neu aufbauen
	docker exec $(BACKEND_CONTAINER) go run ./main.go rebuild-projections

# Aufräumen

clean: ## Dev-Stack stoppen und Volumes entfernen
	docker compose down -v

# Qualitätsprüfung (CI-nah)

check-tools: ## Prüfen, ob lokale Verify-Tools installiert sind
	@for tool in golangci-lint sqlc shellcheck actionlint pnpm; do \
		if ! command -v $$tool >/dev/null 2>&1; then \
			echo "Fehlendes Tool: $$tool"; \
			echo "Installiere es mit scripts/setup-dev-tools.sh."; \
			exit 1; \
		fi; \
	done

check-tools-integration: ## Prüfen, ob migrate und Docker für Integrationstests verfügbar sind
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Fehlendes Tool: migrate"; \
		echo "Installiere es mit scripts/setup-dev-tools.sh."; \
		exit 1; \
	fi
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Fehlendes Tool: docker"; \
		echo "Docker Engine manuell installieren (scripts/setup-dev-tools.sh installiert es nicht)."; \
		exit 1; \
	fi

check-backend: ## Backend komplett prüfen (Deps, Lint inkl. Integrationstest-Dateien, Format, Vet, Test, Build)
	cd backend && go mod tidy -diff && golangci-lint run && $(goimports-check) && go vet ./... && go test -count=1 -race ./... && go build ./...

check-sqlc: ## Prüfen, ob backend/sqlc/dbgen zu Queries und Migrationen passt (sqlc diff)
	cd backend && sqlc diff

check-relay: ## Print-Relay komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd windows/relay && go mod tidy -diff && golangci-lint run && $(goimports-check) && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-starter: ## Windows-Starter komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd windows/starter && go mod tidy -diff && golangci-lint run && $(goimports-check) && go vet ./... && go test -count=1 -race ./... && go build ./...

check-resolver: ## DNS-Resolver komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd resolver && go mod tidy -diff && golangci-lint run && $(goimports-check) && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-local-proxy: ## Lokales Proxy-Entrypoint-Binary komplett prüfen (Deps, Format, Lint, Vet, Test, Build)
	cd reverse-proxy && go mod tidy -diff && golangci-lint run && $(goimports-check) && go vet ./... && go test -count=1 -race ./... && go build -o /dev/null ./...

check-format: ## Repo-weite Prettier-Formatierung prüfen (ts, tsx, js, mjs, cjs, json, css, md)
	frontend/node_modules/.bin/prettier --check $(PRETTIER_GLOB)

check-frontend: ## Frontend komplett prüfen (Format, Typen, Lint, Test, Build)
	$(MAKE) check-format
	cd frontend && pnpm install --frozen-lockfile && pnpm typecheck && pnpm lint && pnpm test && pnpm build

check-e2e: ## E2E-Suite prüfen (tsc + ESLint, ohne Stack)
	cd e2e && pnpm install --frozen-lockfile && pnpm typecheck && pnpm lint

check-shell: ## Shell-Skripte mit shellcheck prüfen (wie CI)
	shellcheck -x scripts/*.sh

check-workflows: ## GitHub-Workflows mit actionlint prüfen (wie CI)
	actionlint

check-repo: ## Alle scripts/check-*.sh-Gates ausführen, Fehlschläge gesammelt am Ende
	@failed=""; \
	for script in scripts/check-*.sh; do \
		echo "→ $$script"; \
		bash "$$script" || failed="$$failed $$script"; \
	done; \
	if [ -n "$$failed" ]; then \
		echo "Fehlgeschlagene Gates:$$failed"; \
		exit 1; \
	fi

check-integration: check-tools-integration ## Integrationstests gegen echte Datenbank ausführen
	./scripts/test-integration.sh

check: check-tools check-backend check-sqlc check-relay check-starter check-resolver check-local-proxy check-frontend website-check check-e2e check-shell check-workflows check-repo ## Schnelle Komplettprüfung ohne DB-Integration

check-full: check check-integration ## Vollständige Prüfung inkl. Integrationstests

verify: check-full ## Alias für vollständige Repo-Prüfung

# Website (Astro + Starlight, website/)
# Die Abhängigkeiten installiert scripts/setup-dev-tools.sh.

website-dev: ## Astro Dev-Server starten (http://localhost:4321), liest docs/ live
	cd website && pnpm dev

website-build: ## Website bauen (Astro Build → website/dist)
	cd website && pnpm build

website-test: ## Website Unit-Tests (Vitest)
	cd website && pnpm test

website-check: ## Website prüfen (Vitest + astro check + Build)
	cd website && pnpm test && pnpm check

# Port des selbst gestarteten Screenshot-Stacks (parametrierbar für parallele Läufe).
E2E_SCREENSHOT_PORT ?= 8080

website-screenshots: ## App-Screenshots + OG-Bild reproduzierbar neu erzeugen (e2e-Stack)
	@# Erzeugt jedes Website-Motiv hell+dunkel (website/src/assets/screenshots/)
	@# und das OG-Bild (website/src/assets/og-startseite.png) via
	@# e2e/website/screenshots.mjs; ohne E2E_BASE_URL gegen einen eigenen
	@# docker-compose.e2e.yml-Stack (JOTTI_ENABLE_TEST_API=1), der danach abgeräumt wird.
	cd e2e && pnpm install --frozen-lockfile && pnpm exec playwright install chromium
	@if [ -n "$$E2E_BASE_URL" ]; then \
	  echo "Nutze laufenden Stack: $$E2E_BASE_URL"; \
	  node e2e/website/screenshots.mjs; \
	else \
	  set -e; \
	  trap 'docker compose -p jotti-screenshots -f docker-compose.e2e.yml down -v' EXIT; \
	  E2E_HTTP_PORT=$(E2E_SCREENSHOT_PORT) docker compose -p jotti-screenshots -f docker-compose.e2e.yml up -d --build; \
	  for i in $$(seq 1 60); do \
	    code=$$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$(E2E_SCREENSHOT_PORT)/ || true); \
	    if [ "$$code" = "200" ]; then echo "Stack ist bereit."; break; fi; \
	    echo "warte auf Stack ($$i/60), Status $${code:-none} ..."; sleep 2; \
	  done; \
	  E2E_BASE_URL=http://localhost:$(E2E_SCREENSHOT_PORT) node e2e/website/screenshots.mjs; \
	fi

# Hilfe

# Every documented target stands in exactly one CLASS_* list, and `make help` prints the
# lists as its sections. A target in none prints under UNCLASSIFIED.
CLASS_developer := init up up-attached down restart logs status \
	test test-frontend test-integration test-tse-live test-tse-live-setup test-all test-e2e fuzz \
	lint-backend lint-backend-full lint-frontend lint-e2e lint fmt-backend fmt-frontend fmt-repo fmt \
	build-backend build-relay build-resolver build-local-proxy build-starter-windows build-relay-windows \
	starter-syso release-windows build-frontend build sqlc \
	local-up local-down local-logs db-shell seed rebuild-projections clean \
	check-tools check-tools-integration check-backend check-sqlc check-relay check-starter check-resolver \
	check-local-proxy check-format check-frontend check-e2e check-shell check-workflows check-repo \
	check-integration check check-full verify \
	website-dev website-build website-test website-check website-screenshots help
CLASS_production := prod-init prod-up prod-update prod-down prod-logs prod-backup prod-restore \
	prod-backup-verify prod-harden \
	rocks-init rocks-up rocks-down rocks-logs rocks-reset-db rocks-reset-and-seed rocks-backup

# Column the help text wraps at.
HELP_WIDTH ?= 96

help: ## Alle Targets nach Klasse anzeigen (HELP_WIDTH=<n> verschiebt den Umbruch)
	@grep -hE '^[a-zA-Z0-9_-]+:.*##' $(firstword $(MAKEFILE_LIST)) \
	  | awk -v width=$(HELP_WIDTH) \
	        -v developer='$(CLASS_developer)' -v production='$(CLASS_production)' \
	        -f scripts/make-help.awk
