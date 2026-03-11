# DevOps, Deployment & Infrastruktur — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für DevOps-Praktiken, Containerisierung, CI/CD, Deployment-Strategien und Betriebskonzepte. Es erklärt Grundlagen von Docker über Orchestrierung und Reverse Proxies bis zu Monitoring, Backup und Self-Hosting.

---

## Inhaltsverzeichnis

1. [Warum DevOps?](#1-warum-devops)
2. [12-Factor App Methodology](#2-12-factor-app-methodology)
3. [Containerisierung mit Docker](#3-containerisierung-mit-docker)
4. [Orchestrierung](#4-orchestrierung)
5. [Reverse Proxy & Load Balancing](#5-reverse-proxy--load-balancing)
6. [CI/CD Pipeline Design](#6-cicd-pipeline-design)
7. [Deployment-Strategien](#7-deployment-strategien)
8. [Zero-Downtime Deployment](#8-zero-downtime-deployment)
9. [Monitoring & Observability](#9-monitoring--observability)
10. [Backup & Disaster Recovery](#10-backup--disaster-recovery)
11. [Self-Hosting Patterns](#11-self-hosting-patterns)
12. [Referenzen](#12-referenzen)

---

## 1. Warum DevOps?

DevOps ist keine Technologie — es ist eine **Kultur und Praxis**, die Entwicklung (Dev) und Betrieb (Ops) zusammenbringt, um Software schneller, zuverlässiger und sicherer bereitzustellen.

### 1.1 Das CALMS-Framework

Das CALMS-Framework beschreibt die fünf Säulen einer DevOps-Kultur:

| Buchstabe | Prinzip     | Bedeutung                                                           |
| --------- | ----------- | ------------------------------------------------------------------- |
| **C**     | Culture     | Gemeinsame Verantwortung für Entwicklung und Betrieb                |
| **A**     | Automation  | Manuelle Prozesse automatisieren (Tests, Builds, Deployments)       |
| **L**     | Lean        | Verschwendung eliminieren, kontinuierliche Verbesserung             |
| **M**     | Measurement | Metriken erheben, Hypothesen validieren, datengetrieben entscheiden |
| **S**     | Sharing     | Wissen, Verantwortung und Werkzeuge über Team-Grenzen hinweg teilen |

### 1.2 Three Ways of DevOps

Gene Kim beschreibt in _The Phoenix Project_ drei fundamentale Prinzipien:

**First Way — Flow (von Dev nach Ops):** Den Durchfluss von Arbeit von links nach rechts (Entwicklung → Betrieb → Kunde) beschleunigen. Kleine Batches, Automatisierung, Sichtbarkeit der gesamten Pipeline.

**Second Way — Feedback:** Schnelle und kontinuierliche Rückmeldung von rechts nach links. Fehler frühzeitig entdecken, aus Problemen lernen.

**Third Way — Continuous Experimentation & Learning:** Kultur des kontinuierlichen Experimentierens, Risiken eingehen, aus Fehlern lernen, Wiederholungen automatisieren.

### 1.3 CI/CD/CD — Begriffsabgrenzung

```
Continuous Integration (CI)
  └── Code wird bei jedem Commit automatisch gebaut und getestet

Continuous Delivery (CD)
  └── Jeder Build ist potenziell produktionsbereit (Deployment auf Knopfdruck)

Continuous Deployment
  └── Jeder erfolgreiche Build wird automatisch in Produktion deployed
```

**Martin Fowler:** _„Continuous Delivery ist eine Disziplin, bei der Software so gebaut wird, dass sie jederzeit in Produktion deployt werden kann."_

Der entscheidende Test: Ein Business-Sponsor könnte jederzeit verlangen, dass die aktuelle Entwicklungsversion in Produktion geht — und niemand würde in Panik geraten.

---

## 2. 12-Factor App Methodology

Die [12-Factor App](https://12factor.net/) ist ein Regelwerk für den Bau von cloud-nativen, wartbaren Applikationen. Sie ist besonders relevant für containerisierte Deployments.

### Die zwölf Faktoren

| #    | Faktor                | Prinzip                                                             |
| ---- | --------------------- | ------------------------------------------------------------------- |
| I    | **Codebase**          | Ein Codebase in Versionskontrolle, viele Deploys                    |
| II   | **Dependencies**      | Abhängigkeiten explizit deklarieren und isolieren                   |
| III  | **Config**            | Konfiguration in Umgebungsvariablen speichern (nie im Code)         |
| IV   | **Backing Services**  | Backing Services (DB, Cache) als austauschbare Ressourcen behandeln |
| V    | **Build/Release/Run** | Build- und Run-Stage strikt trennen                                 |
| VI   | **Processes**         | App als zustandslose Prozesse ausführen                             |
| VII  | **Port Binding**      | Services via Port-Binding exportieren                               |
| VIII | **Concurrency**       | Horizontal skalieren via Prozessmodell                              |
| IX   | **Disposability**     | Robustheit durch schnellen Start und Graceful Shutdown              |
| X    | **Dev/Prod Parity**   | Entwicklung, Staging und Produktion so ähnlich wie möglich halten   |
| XI   | **Logs**              | Logs als Event-Streams behandeln (nach stdout schreiben)            |
| XII  | **Admin Processes**   | Verwaltungsaufgaben als einmalige Prozesse ausführen                |

### Besonders wichtige Faktoren für self-hosted Apps

**Faktor III — Config:** Alle Konfiguration (Passwörter, Secrets, URLs) gehört in Umgebungsvariablen, nie in den Code. Bei Docker: `.env`-Dateien oder `environment:` in Compose-Dateien.

```yaml
# docker-compose.yml
environment:
  DATABASE_URL: postgres://${DB_USER}:${DB_PASS}@db:5432/myapp
  JWT_SECRET: ${JWT_SECRET}
```

**Faktor IX — Disposability:** Container sollten schnell starten (< 5 Sekunden) und bei `SIGTERM` sauber herunterfahren (Verbindungen schließen, laufende Requests abschließen).

**Faktor X — Dev/Prod Parity:** Entwicklung und Produktion sollten möglichst identisch sein. Docker Compose ermöglicht dieselbe Datenbankversion lokal wie in Produktion.

**Faktor XI — Logs:** Applikationen schreiben nach `stdout`/`stderr`. Die Infrastruktur (Docker, Kubernetes) leitet Logs weiter. Kein Log-Management in der App selbst.

---

## 3. Containerisierung mit Docker

Docker kapselt Anwendungen und ihre Abhängigkeiten in portable, isolierte Container.

### 3.1 Docker-Grundkonzepte

```
Image        = Unveränderlicher Snapshot (Dateisystem + Metadaten)
Container    = Laufende Instanz eines Images (isolierter Prozess)
Registry     = Speicher für Images (Docker Hub, GHCR, eigene Registry)
Dockerfile   = Bauanleitung für ein Image
```

**Image-Layer-Modell:** Jede Dockerfile-Instruktion erstellt einen neuen Layer. Layers werden gecacht — wenn ein Layer sich nicht ändert, wird der Cache wiederverwendet. Reihenfolge der Instruktionen ist daher entscheidend für Build-Geschwindigkeit.

### 3.2 Multi-Stage Builds

Multi-Stage Builds trennen Build-Umgebung von der Runtime-Umgebung:

```dockerfile
# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download              # Layer: wird gecacht, wenn go.mod unverändert
COPY . .
RUN go build -o /app/server ./main.go

# Stage 2: Runtime (minimal)
FROM alpine:3.21
COPY --from=builder /app/server /usr/local/bin/server
ENTRYPOINT ["server"]
```

**Vorteile:**

- Build-Tools (Compiler, Dev-Dependencies) landen nicht im finalen Image
- Finale Images sind klein (Alpine: ~5 MB statt ~300 MB mit Go-SDK)
- Angriffsfläche reduziert: weniger Binaries im Container

**Vergleich Image-Größen (Go-Backend):**

| Base Image           | Größe (typisch) | Anmerkung                                 |
| -------------------- | --------------- | ----------------------------------------- |
| `golang:1.24`        | ~800 MB         | Komplettes Go-SDK, für Build geeignet     |
| `golang:1.24-alpine` | ~250 MB         | Alpine-basiert, für Build geeignet        |
| `alpine:3.21`        | ~8 MB           | Minimales Runtime-Image                   |
| `scratch`            | 0 MB            | Leeres Image (nur für statische Binaries) |
| `distroless`         | ~5 MB           | Google, kein Shell — sehr sicher          |

### 3.3 Layer-Caching optimieren

Die Reihenfolge von Dockerfile-Instruktionen bestimmt Cache-Effizienz:

```dockerfile
# ❌ Schlecht: Quellcode-Copy vor Dependency-Installation
COPY . .
RUN npm install       # Cache-Miss bei jeder Code-Änderung

# ✓ Besser: Dependencies zuerst, dann Quellcode
COPY package.json package-lock.json ./
RUN npm ci            # Nur neu installieren, wenn package.json sich ändert
COPY . .
RUN npm run build
```

**Faustregel:** Was sich selten ändert (Dependencies, Config), kommt früher. Was sich oft ändert (Quellcode), kommt später.

### 3.4 .dockerignore

Analog zu `.gitignore` — schließt Dateien vom Build-Kontext aus:

```
node_modules/
dist/
.git/
*.md
.env
.env.local
coverage/
__tests__/
```

**Warum wichtig?** Docker schickt den gesamten Build-Kontext an den Docker-Daemon. Große oder sensible Dateien verlängern den Build und könnten ins Image gelangen.

### 3.5 Container Security

**Prinzip: Least Privilege**

```dockerfile
# Nicht als root laufen
RUN adduser -D -u 1001 appuser
USER appuser

# Read-only Filesystem (in Compose)
read_only: true
tmpfs:
  - /tmp
```

**Image-Hygiene:**

- **Basis-Images pinnen:** `FROM alpine:3.21.3` statt `FROM alpine:latest` (Reproduzierbarkeit)
- **Regelmäßig neu bauen:** Base Images erhalten Security Patches — Images müssen neu gebaut werden
- **Nur Offizielle/Verifizierte Images nutzen:** Docker Hub kennzeichnet geprüfte Images
- **Dependency Scanning:** Tools wie `trivy`, `snyk`, `docker scout` scannen Images auf CVEs

**Secrets nie ins Image:** Umgebungsvariablen, `.env`-Dateien oder Secret-Management-Tools (Vault, Docker Secrets) verwenden — niemals `RUN curl ... --header "Authorization: Bearer hardcoded-secret"`.

### 3.6 Docker Networking

Docker erstellt virtuelle Netzwerke zwischen Containern:

```yaml
networks:
  app-network: # Frontend ↔ Backend ↔ Reverse Proxy
    driver: bridge
  db-network: # Backend ↔ Datenbank (intern, kein externer Zugriff)
    driver: bridge
    internal: true # Kein Internetzugriff aus diesem Netzwerk
```

**Network Isolation:** Datenbankcontainer sollten in einem `internal`-Netzwerk liegen, das nur der Backend-Container erreichen kann. Der Reverse Proxy liegt im `app-network`.

---

## 4. Orchestrierung

Orchestrierung verwaltet, wie Container gestartet, gestoppt, skaliert und vernetzt werden.

### 4.1 Docker Compose

Docker Compose ist die einfachste Form der Orchestrierung — ideal für Self-Hosted und kleine Teams.

**Stärken:**

- Deklarativ (YAML): gesamter Stack in einer Datei beschrieben
- `depends_on` mit Health Checks: Startreihenfolge kontrollieren
- Volumes: persistente Daten überleben Container-Neustarts
- Einfaches Deployment: `docker compose up -d --build`

**Schwächen:**

- Kein automatisches Failover (Container neu starten auf _gleichem_ Host)
- Kein Load Balancing über mehrere Hosts
- Kein Rolling Update ohne Downtime (standardmäßig)

```yaml
# Health Check in Compose
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
  interval: 5s
  timeout: 3s
  retries: 10

# Startabhängigkeit mit Health Check
depends_on:
  postgres:
    condition: service_healthy
```

### 4.2 Kubernetes

Kubernetes (K8s) ist der De-facto-Standard für Container-Orchestrierung in der Cloud und bei hohen Anforderungen.

**Kernkonzepte:**

| Objekt         | Beschreibung                                                        |
| -------------- | ------------------------------------------------------------------- |
| **Pod**        | Kleinste Einheit: ein oder mehrere Container + geteiltes Netzwerk   |
| **Deployment** | Deklarativer Ist-Zustand für Pods (Rolling Updates, Replicas)       |
| **Service**    | Stabiler Endpunkt für eine Gruppe von Pods (Load Balancing)         |
| **Ingress**    | HTTP/HTTPS-Routing von außen in Services                            |
| **ConfigMap**  | Konfiguration (nicht-sensitiv)                                      |
| **Secret**     | Sensible Konfiguration (base64-kodiert, idealerweise verschlüsselt) |
| **PVC**        | Persistente Volumes für Stateful Workloads                          |

**Kubernetes-Architektur:**

```
Control Plane
  ├── API Server      — REST-API für alle K8s-Operationen
  ├── etcd            — Verteilter Key-Value-Store (Cluster-State)
  ├── Scheduler       — Entscheidet, auf welchem Node ein Pod läuft
  └── Controller Mgr  — Stellt Soll-Zustand her (Reconciliation Loop)

Worker Nodes
  ├── kubelet         — Agent auf jedem Node, führt Pods aus
  ├── kube-proxy      — Netzwerk-Routing für Services
  └── Container Runtime (containerd, CRI-O)
```

### 4.3 K3s — Leichtgewichtiges Kubernetes

[K3s](https://k3s.io/) ist eine zertifizierte Kubernetes-Distribution, optimiert für:

- **Edge Computing** und Single-Node-Setups
- **ARM-Prozessoren** (Raspberry Pi, ARM-Server)
- **Self-Hosted** mit minimalem Ressourcenbedarf (~512 MB RAM)

Unterschiede zu Standard-K8s: etcd durch SQLite ersetzt, viele optionale Komponenten entfernt, einzelnes Binary.

### 4.4 Entscheidungsmatrix: Orchestrierung

| Kriterium              | Docker Compose            | K3s               | Kubernetes                |
| ---------------------- | ------------------------- | ----------------- | ------------------------- |
| **Einstiegshürde**     | Sehr gering               | Mittel            | Hoch                      |
| **Ressourcenbedarf**   | Minimal                   | ~512 MB RAM       | ~2 GB RAM (Control Plane) |
| **Hochverfügbarkeit**  | ❌ Single Host            | ✓ (mit HA-Modus)  | ✓ (Multi-Node)            |
| **Auto-Healing**       | ✓ restart: unless-stopped | ✓                 | ✓                         |
| **Rolling Updates**    | Manuell                   | ✓                 | ✓                         |
| **Horizontal Scaling** | Manuell                   | ✓                 | ✓                         |
| **Zielgruppe**         | Solo/Klein-Teams, Dev     | Edge, Self-Hosted | Cloud, Enterprise         |

**Empfehlung für Self-Hosted Vereinssoftware:** Docker Compose. Einfach, wartbar, keine Kubernetes-Expertise nötig. Erst bei Multi-Node oder Hochverfügbarkeitsanforderungen wechseln.

---

## 5. Reverse Proxy & Load Balancing

Ein Reverse Proxy empfängt alle eingehenden Anfragen und leitet sie an die entsprechenden Backends weiter.

### 5.1 Aufgaben eines Reverse Proxys

```
Internet → Reverse Proxy → Backend-Services
                ├── TLS-Terminierung (HTTPS → HTTP intern)
                ├── Load Balancing (mehrere Backend-Instanzen)
                ├── Routing (nach Pfad, Host, Header)
                ├── Rate Limiting & Throttling
                ├── Security Headers (HSTS, CSP, X-Frame-Options)
                ├── Caching (statische Assets)
                ├── Compression (gzip, brotli)
                └── Access Logging
```

### 5.2 nginx

[nginx](https://nginx.org/) ist ein hochperformanter, weit verbreiteter Reverse Proxy und Webserver.

**Stärken:**

- Extrem performant (event-driven, non-blocking I/O)
- Ausgereift und gut dokumentiert
- Volle Kontrolle über Konfiguration
- Statische Dateien direkt ausliefern (sehr effizient)

**Schwächen:**

- Konfiguration muss manuell gepflegt werden (keine automatische Service-Discovery)
- Reload bei Konfigurationsänderungen nötig (`nginx -s reload`)
- TLS-Zertifikate müssen separat verwaltet werden (z.B. Certbot)

**Typische nginx-Konfiguration (Reverse Proxy + TLS):**

```nginx
# Syntax für nginx ≥ 1.25.1 (http2 als separate Direktive)
# Für ältere Versionen: listen 443 ssl http2;
server {
  listen 443 ssl;
  http2 on;
  server_name example.com;

  ssl_certificate     /etc/letsencrypt/live/example.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
  ssl_protocols TLSv1.2 TLSv1.3;

  # Security Headers
  add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
  add_header X-Content-Type-Options nosniff always;
  add_header X-Frame-Options DENY always;

  # API-Backend
  location /api/ {
    proxy_pass http://backend:3000/;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }

  # Frontend SPA
  location / {
    proxy_pass http://frontend:80;
  }
}
```

### 5.3 Traefik

[Traefik](https://traefik.io/) ist ein moderner Reverse Proxy mit eingebauter Service-Discovery, entwickelt für Container-Umgebungen.

**Stärken:**

- **Automatische Service-Discovery:** Liest Docker-Labels und konfiguriert sich selbst
- **Automatisches TLS:** Let's Encrypt-Integration ohne Certbot
- **Dashboard:** Eingebaute Web-UI für Routing-Übersicht
- **Middleware:** Plugin-System für Rate Limiting, Auth, etc.

**Schwächen:**

- Höhere Komplexität bei statischer Konfiguration
- Mehr Ressourcenbedarf als nginx
- Dashboard muss gesichert werden

**Traefik mit Docker-Labels:**

```yaml
# docker-compose.yml
services:
  backend:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`example.com`) && PathPrefix(`/api`)"
      - "traefik.http.routers.backend.tls.certresolver=letsencrypt"
      - "traefik.http.services.backend.loadbalancer.server.port=3000"
```

### 5.4 Caddy

[Caddy](https://caddyserver.com/) ist ein moderner Webserver mit automatischem HTTPS — Zertifikate werden vollautomatisch bezogen und erneuert.

**Stärken:**

- **Automatisches HTTPS:** Out-of-the-box (ACME via Let's Encrypt oder ZeroSSL)
- **Einfache Konfiguration:** Caddyfile ist deutlich lesbarer als nginx-Syntax
- **HTTP/2 und HTTP/3:** Standardmäßig aktiviert
- **Go-basiert:** Kein OpenSSL, kleinere Angriffsfläche

**Schwächen:**

- Weniger verbreitet → weniger Community-Ressourcen
- Weniger Feintuning-Optionen als nginx
- Ältere, bewährtere Systeme bevorzugen oft nginx aus Gewohnheit

**Caddyfile-Beispiel:**

```caddyfile
example.com {
  reverse_proxy /api/* backend:3000
  reverse_proxy * frontend:80
  header {
    Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
    X-Content-Type-Options nosniff
    X-Frame-Options DENY
  }
}
```

### 5.5 Entscheidungsmatrix: Reverse Proxy

| Kriterium                     | nginx               | Traefik              | Caddy                   |
| ----------------------------- | ------------------- | -------------------- | ----------------------- |
| **Einstiegshürde**            | Mittel              | Mittel-Hoch          | Gering                  |
| **Automatisches TLS**         | ❌ (Certbot nötig)  | ✓                    | ✓                       |
| **Service-Discovery**         | ❌                  | ✓ (Docker-Labels)    | ❌                      |
| **Performance**               | ★★★★★               | ★★★★                 | ★★★★                    |
| **Konfigurationskomplexität** | Mittel              | Hoch                 | Gering                  |
| **Community & Dokumentation** | Sehr groß           | Groß                 | Mittel                  |
| **Ressourcenverbrauch**       | Sehr gering         | Mittel               | Gering                  |
| **Zielgruppe**                | Erfahrene Ops-Teams | Container-Umgebungen | Entwickler, Self-Hosted |

---

## 6. CI/CD Pipeline Design

Eine CI/CD-Pipeline automatisiert den Weg vom Code-Commit bis zum deployt​en Artefakt.

### 6.1 Pipeline-Stufen

```
Developer Push
     │
     ▼
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Lint    │───►│  Test    │───►│  Build   │───►│ Publish  │───►│  Deploy  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘
  Format          Unit-Tests      Kompilieren     Image pushen    Neuen Build
  Code Style       Int.-Tests       Docker         Registry         aktivieren
  Security Scan    Coverage         Build
```

**Jede Stufe muss die nachfolgende Stufe verhindern, wenn sie fehlschlägt.** Kein Deployment mit fehlgeschlagenen Tests.

### 6.2 GitHub Actions

GitHub Actions ist eine Cloud-native CI/CD-Plattform, die direkt in GitHub integriert ist.

**Grundstruktur:**

```yaml
name: CI Pipeline

on: # Trigger
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs: # Parallele oder sequenzielle Jobs
  test:
    runs-on: ubuntu-latest # Runner-Umgebung
    steps:
      - uses: actions/checkout@v5 # Repo auschecken
      - uses: actions/setup-go@v6 # Go installieren
        with:
          go-version: 1.24.0
      - run: go test ./... # Befehl ausführen
```

**Key Concepts:**

| Konzept      | Beschreibung                                                           |
| ------------ | ---------------------------------------------------------------------- |
| **Workflow** | Automatisierter Prozess (YAML-Datei in `.github/workflows/`)           |
| **Event**    | Trigger: `push`, `pull_request`, `schedule`, `workflow_dispatch`, …    |
| **Job**      | Gruppe von Steps, läuft auf einem Runner (standardmäßig parallel)      |
| **Step**     | Einzelner Befehl oder Action                                           |
| **Action**   | Wiederverwendbare Komponente (`uses: actions/checkout@v5`)             |
| **Runner**   | Virtuelle Maschine, auf der Jobs laufen (`ubuntu-latest`, self-hosted) |
| **Artifact** | Dateien, die zwischen Jobs oder Workflows geteilt werden               |
| **Secret**   | Verschlüsselte Umgebungsvariable (Passwörter, Tokens)                  |

**Path-based Filtering (Build nur was sich geändert hat):**

```yaml
jobs:
  changes:
    outputs:
      backend: ${{ steps.filter.outputs.backend }}
    steps:
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            backend:
              - 'backend/**'

  backend-ci:
    needs: changes
    if: ${{ needs.changes.outputs.backend == 'true' }}
    # ...
```

### 6.3 Pipeline Best Practices

**Fail Fast:** Schnelle Prüfungen (Lint, Format) zuerst. Teure Operationen (Integration Tests, Builds) erst nach schnellen Checks.

**Caching:** Dependencies zwischen Runs cachen — spart Minuten:

```yaml
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: ${{ runner.os }}-go-
```

**Secrets sichern:**

- Secrets nie in Logs ausgeben
- Minimale Permissions für Tokens (Principle of Least Privilege)
- `GITHUB_TOKEN` für interne Operationen, keine langlebigen Personal Access Tokens

**Dependabot:** Automatische Dependency-Updates für Actions-Versionen:

```yaml
# .github/dependabot.yml
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

### 6.4 Deployment aus der Pipeline

**Container Registry → Deployment:**

```yaml
deploy:
  needs: [test, build]
  if: github.ref == 'refs/heads/main'
  steps:
    - name: Build and push Docker image
      uses: docker/build-push-action@v6
      with:
        push: true
        tags: ghcr.io/${{ github.repository }}:${{ github.sha }}

    - name: Deploy to server
      uses: appleboy/ssh-action@master
      with:
        host: ${{ secrets.SERVER_HOST }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        script: |
          docker compose pull
          docker compose up -d --no-deps backend
```

---

## 7. Deployment-Strategien

Die Wahl der Deployment-Strategie bestimmt Downtime, Risiko und Rollback-Fähigkeit.

### 7.1 Recreate (Stop → Start)

```
v1 ████████████  STOP
                  START  v2 ████████████
```

**Vorgehen:** Alle Instanzen der alten Version werden gestoppt, dann wird die neue Version gestartet.

**Downtime:** Ja — zwischen Stop und Start (Sekunden bis Minuten)

**Wann sinnvoll:**

- Kleine Teams, kein 24/7-Betrieb
- Nicht rückwärtskompatible Änderungen (DB-Schema-Break)
- Einfachste Variante, am einfachsten zu debuggen

### 7.2 Rolling Update

```
v1 ████  ████  ████  ████
         ↓           ↓
v2       ████        ████  ████  ████
```

**Vorgehen:** Instanzen werden schrittweise ersetzt (eine nach der anderen oder in Batches). Neue und alte Version laufen kurzzeitig parallel.

**Downtime:** Keine (bei korrekter Implementierung)

**Voraussetzungen:**

- v1 und v2 müssen gleichzeitig betreibbar sein (API-Kompatibilität)
- Load Balancer muss Traffic auf verbleibende gesunde Instanzen umleiten
- DB-Migrations müssen rückwärtskompatibel sein (Expand/Contract Pattern)

### 7.3 Blue/Green Deployment

```
        Router
       /      \
  Blue (v1)   Green (v2)
  ████████    ████████
  (aktiv)     (staging)
```

**Vorgehen:** Zwei identische Produktionsumgebungen (Blue und Green). Neue Version auf der inaktiven Umgebung deployen, testen, dann Traffic umleiten.

**Downtime:** Keine (Umleitung ist sofort)

**Stärken:**

- Sofortiges Rollback: Router zurück auf Blue umleiten
- Neue Version kann in Produktionsumgebung getestet werden, bevor Traffic umgeleitet wird

**Schwächen:**

- Doppelter Infrastruktur-Aufwand (zwei komplette Environments)
- DB-State muss synchronisiert werden oder geteilt werden

### 7.4 Canary Deployment

```
        Router
       /      \
  Main (v1)   Canary (v2)
  ████████     ██
   95% Traffic  5% Traffic
```

**Vorgehen:** Neue Version erhält zunächst nur einen kleinen Prozentsatz des Traffics. Bei guten Metriken wird der Anteil schrittweise erhöht.

**Stärken:**

- Risikominimierung: Fehler betreffen nur wenige Nutzer
- Real-World-Validierung mit echten Daten
- Schrittweises Rollout möglich

**Schwächen:**

- Erfordert Feature Flags oder Traffic-Splitting-Fähigkeit des Load Balancers
- Monitoring muss gut genug sein, um Probleme zu erkennen

### 7.5 Entscheidungsmatrix

| Strategie      | Downtime | Rollback         | Ressourcen | Komplexität | Wann empfohlen                       |
| -------------- | -------- | ---------------- | ---------- | ----------- | ------------------------------------ |
| **Recreate**   | Ja       | Einfach (Revert) | Niedrig    | Gering      | Kleine Systeme, akzeptable Downtime  |
| **Rolling**    | Nein     | Mittel           | Mittel     | Mittel      | Standard für die meisten Apps        |
| **Blue/Green** | Nein     | Sofort           | Hoch       | Mittel-Hoch | Kritische Systeme, Zero-Downtime     |
| **Canary**     | Nein     | Schrittweise     | Mittel     | Hoch        | Große Nutzerbasis, Risikominimierung |

---

## 8. Zero-Downtime Deployment

Zero-Downtime bedeutet: Nutzer merken nichts vom Deployment — keine Fehlermeldungen, keine Unterbrechungen.

### 8.1 Graceful Shutdown

Beim Empfang von `SIGTERM` soll die Anwendung:

1. Keine neuen Requests mehr akzeptieren
2. Laufende Requests abschließen
3. Datenbankverbindungen schließen
4. Sauber beenden (Exit Code 0)

**Go-Beispiel:**

```go
server := &http.Server{Addr: ":3000", Handler: router}

// Server in Goroutine starten
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal().Err(err).Msg("server error")
    }
}()

// Auf SIGTERM/SIGINT warten
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
<-quit

// Graceful Shutdown mit Timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := server.Shutdown(ctx); err != nil {
    log.Fatal().Err(err).Msg("shutdown error")
}
```

### 8.2 Health Checks

Health Checks ermöglichen Orchestratoren (Docker, Kubernetes) zu erkennen, wann ein Container bereit ist.

**Zwei Typen:**

| Typ           | Frage                                        | Aktion bei Fehler                    |
| ------------- | -------------------------------------------- | ------------------------------------ |
| **Liveness**  | Läuft der Prozess noch? (nicht blockiert)    | Container neu starten                |
| **Readiness** | Kann der Container Traffic empfangen?        | Aus Load-Balancer-Rotation entfernen |
| **Startup**   | Hat die App gestartet? (für langsame Starts) | Wartet, bis OK (vor Liveness)        |

**HTTP Health Check Endpoint:**

```go
// GET /healthz — Liveness
func healthz(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

// GET /readyz — Readiness (prüft DB-Verbindung)
func readyz(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(r.Context()); err != nil {
        http.Error(w, "db not ready", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}
```

**Docker Health Check:**

```dockerfile
HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --retries=3 \
  CMD wget -q --spider http://localhost:3000/healthz || exit 1
```

### 8.3 Zero-Downtime DB-Migrations

Datenbankmigrationen sind der kritischste Teil von Zero-Downtime Deployments. Die **Expand/Contract**-Strategie löst das Problem:

**Phase 1 — Expand (rückwärtskompatibel addieren):**

```sql
-- Migration: Neue Spalte hinzufügen (nullable oder mit Default)
ALTER TABLE users ADD COLUMN display_name TEXT;
-- v1 ignoriert die Spalte, v2 schreibt in sie
```

**Phase 2 — Migrate (Daten migrieren, ggf. Backfill):**

```sql
-- Bestehendes Daten befüllen
UPDATE users SET display_name = username WHERE display_name IS NULL;
-- Spalte als NOT NULL markieren (erst wenn alle Daten befüllt)
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
```

**Phase 3 — Contract (altes entfernen, nachdem v2 vollständig deployed):**

```sql
-- Jetzt ist es sicher, die alte Spalte zu löschen
ALTER TABLE users DROP COLUMN username;
```

**Wichtige Regeln:**

- Neue Spalten immer nullable oder mit Default-Wert hinzufügen
- Spalten nie in derselben Migration umbenennen und löschen
- Nie Datentypen inkompatibel ändern (INT → TEXT ist OK, TEXT → INT ist gefährlich)
- Migrationen vor dem Code-Deployment ausführen (Migration zuerst, dann Deploy)

---

## 9. Monitoring & Observability

**Monitoring** fragt: Ist das System gesund?  
**Observability** fragt: Warum verhält sich das System so?

Observability besteht aus drei Säulen: **Logs, Metriken und Traces**.

### 9.1 Die vier Goldenen Signale (Google SRE)

Das Google SRE-Buch definiert vier Signale, die jedes System überwachen sollte:

| Signal         | Beschreibung                         | Beispiel                         |
| -------------- | ------------------------------------ | -------------------------------- |
| **Latency**    | Wie lange dauern Requests?           | p50/p95/p99 Response Time        |
| **Traffic**    | Wie viel Last hat das System?        | Requests/Sekunde, DB-Queries/Min |
| **Errors**     | Wie viele Requests schlagen fehl?    | HTTP 5xx Rate, Exception Rate    |
| **Saturation** | Wie ausgelastet sind die Ressourcen? | CPU%, Memory%, Disk I/O%         |

### 9.2 RED und USE Method

**RED Method** (für Services/APIs, von Tom Wilkie):

- **R**ate — Wie viele Requests/Sekunde?
- **E**rrors — Wie viele Fehler/Sekunde?
- **D**uration — Wie lang dauern Requests?

**USE Method** (für Ressourcen/Infrastruktur, von Brendan Gregg):

- **U**tilization — Wie ausgelastet ist die Ressource? (CPU %)
- **S**aturation — Wie viel Wartezeit gibt es? (Run Queue)
- **E**rrors — Wie viele Fehler gibt es? (Disk Errors)

**Empfehlung:** RED für Application-Level-Monitoring, USE für Infrastructure-Monitoring.

### 9.3 Prometheus

[Prometheus](https://prometheus.io/) ist ein Pull-basiertes Monitoring-System mit eigenem Query-Language PromQL.

**Architektur:**

```
Prometheus Server
  ├── Scraper: Holt Metriken von /metrics Endpoints
  ├── TSDB: Time-Series Database (lokale Speicherung)
  ├── Alertmanager: Wertet Alert-Regeln aus, sendet Benachrichtigungen
  └── PromQL: Query-Sprache für Metriken-Aggregation
```

**Metrik-Typen:**

| Typ           | Beschreibung                      | Beispiel                        |
| ------------- | --------------------------------- | ------------------------------- |
| **Counter**   | Monoton wachsender Wert           | `http_requests_total`           |
| **Gauge**     | Wert kann steigen und fallen      | `memory_usage_bytes`            |
| **Histogram** | Verteilung von Werten (Buckets)   | `http_request_duration_seconds` |
| **Summary**   | Wie Histogram, aber mit Quantilen | `rpc_duration_seconds`          |

**PromQL-Beispiele:**

```promql
# Request Rate der letzten 5 Minuten
rate(http_requests_total[5m])

# 95. Perzentil der Response-Zeit
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error Rate (Anteil 5xx)
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])
```

### 9.4 Grafana

[Grafana](https://grafana.com/) ist das Standard-Visualisierungs-Tool für Prometheus-Metriken.

**Features:**

- Dashboards aus PromQL-Queries
- Alerting auf Basis von Grafana-Panels
- Unterstützt viele Datasources (Prometheus, Loki, PostgreSQL, Elasticsearch, …)
- Vorgefertigte Dashboards auf [Grafana Dashboards](https://grafana.com/grafana/dashboards/) (z.B. Node Exporter, PostgreSQL)

### 9.5 Loki — Log Aggregation

[Grafana Loki](https://grafana.com/oss/loki/) ist ein Log-Aggregations-System, das mit Grafana zusammenarbeitet. Ähnlich wie Prometheus, aber für Logs.

**Komponenten:**

```
Promtail (Log Collector)  →  Loki (Speicher)  →  Grafana (Visualisierung)
```

**Funktionsweise:**

- Promtail läuft auf jedem Server und sendet Logs an Loki
- Loki indiziert nur Labels (wie Prometheus), nicht den Log-Inhalt
- Sehr ressourcensparend im Vergleich zu Elasticsearch
- LogQL (ähnlich PromQL) für Queries

**Alternativen:**

- **ELK Stack** (Elasticsearch, Logstash, Kibana): Mächtiger, aber ressourcenintensiv
- **OpenSearch**: Open-Source Fork von Elasticsearch
- **Vector**: Hochperformanter Log-Kollektor (Rust), Alternative zu Promtail

### 9.6 Structured Logging

Für gute Observability müssen Logs maschinenlesbar sein:

```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "service": "backend",
  "method": "POST",
  "path": "/api/create-order",
  "status": 200,
  "duration_ms": 45,
  "user_id": "42",
  "order_id": "7"
}
```

**Vorteile:** Logs können nach beliebigen Feldern gefiltert und aggregiert werden. Kein String-Parsing nötig.

---

## 10. Backup & Disaster Recovery

Backup ist kein Selbstzweck — es geht darum, im Ernstfall **innerhalb eines definierten Zeitraums** wiederherstellen zu können.

### 10.1 RTO und RPO

**Recovery Time Objective (RTO):** Wie lange darf das System ausfallen?

- Beispiel: „Wir müssen innerhalb von 4 Stunden wieder online sein"

**Recovery Point Objective (RPO):** Wie viel Datenverlust ist akzeptabel?

- Beispiel: „Wir dürfen maximal 1 Stunde Daten verlieren"

**Tradeoff:** Niedrigeres RTO/RPO = höhere Kosten (häufigere Backups, Replikation, schnelleres Restore-Verfahren).

### 10.2 PostgreSQL-Backup

**pg_dump — Logische Backups:**

```bash
# Backup erstellen (komprimiert)
pg_dump -h localhost -U admin -d myapp --format=custom -f backup.dump

# Backup wiederherstellen
pg_restore -h localhost -U admin -d myapp --clean backup.dump

# Backup mit Timestamp
pg_dump ... -f "backup_$(date +%Y%m%d_%H%M%S).dump"
```

**pg_basebackup — Physische Backups (für PITR):**

```bash
# Gesamtes Cluster-Backup
pg_basebackup -h localhost -U admin -D /backups/base -Ft -z -P

# Point-in-Time Recovery (PITR) mit WAL-Archivierung
# Ermöglicht Wiederherstellung zu einem beliebigen Zeitpunkt
```

**Backup-Strategie:**

| Backup-Typ       | Frequenz       | Aufbewahrung | Zweck                         |
| ---------------- | -------------- | ------------ | ----------------------------- |
| Full Dump        | Täglich        | 7 Tage       | Einfachste Wiederherstellung  |
| Full Dump        | Wöchentlich    | 4 Wochen     | Langzeit-Archiv               |
| Full Dump        | Monatlich      | 12 Monate    | Compliance-Archiv             |
| WAL-Archivierung | Kontinuierlich | 7 Tage       | PITR (Point-in-Time Recovery) |

### 10.3 Restic — Deduplizierendes Backup

[Restic](https://restic.net/) ist ein modernes, verschlüsseltes Backup-Tool.

**Stärken:**

- **Deduplizierung:** Gleiche Daten werden nur einmal gespeichert
- **Verschlüsselung:** AES-256 (Backups sind immer verschlüsselt)
- **Viele Backend-Optionen:** Lokal, S3, Backblaze B2, SFTP, Azure, GCS
- **Snapshots:** Zeitpunkte mit Abhängigkeiten zwischen Snapshots

```bash
# Repository initialisieren
restic init --repo s3:s3.amazonaws.com/mybucket/restic

# Backup erstellen
restic backup /var/lib/postgresql/data \
  --repo s3:s3.amazonaws.com/mybucket/restic \
  --password-file /etc/restic/password

# Snapshots auflisten
restic snapshots

# Wiederherstellen
restic restore latest --target /tmp/restore
```

**3-2-1-Backup-Regel:**

- **3** Kopien der Daten
- auf **2** verschiedenen Medientypen
- davon **1** offsite (extern, Cloud)

### 10.4 Backup-Verifizierung

Ein Backup ist wertlos, wenn die Wiederherstellung nicht funktioniert.

**Regelmäßige Restore-Tests:**

```bash
# Automatisches Restore-Testing (z.B. wöchentlich in CI)
#!/bin/bash
pg_dump -f /tmp/test-backup.dump
pg_restore -d testdb /tmp/test-backup.dump
psql -d testdb -c "SELECT COUNT(*) FROM users;" | grep -q "[0-9]"
echo "Restore successful: $(date)"
```

---

## 11. Self-Hosting Patterns

Self-Hosting bedeutet, eine Anwendung auf eigener Infrastruktur zu betreiben — ohne Cloud-Provider.

### 11.1 VM-Provisionierung

**Minimale Produktionsanforderungen:**

| Ressource | Minimum    | Empfohlen |
| --------- | ---------- | --------- |
| vCPUs     | 1          | 2         |
| RAM       | 1 GB       | 2–4 GB    |
| Disk      | 20 GB SSD  | 50 GB SSD |
| Netzwerk  | 100 Mbit/s | 1 Gbit/s  |

**Anbieter-Vergleich (VPS):**

| Anbieter      | Einstiegspaket | Besonderheiten                   |
| ------------- | -------------- | -------------------------------- |
| Hetzner       | ~4 €/Monat     | Deutsch, DSGVO-konform, sehr gut |
| DigitalOcean  | ~6 $/Monat     | Gute Dokumentation, App Platform |
| Linode/Akamai | ~5 $/Monat     | Solide, globale Präsenz          |
| Contabo       | ~5 €/Monat     | Viel Ressourcen für wenig Geld   |

**OS-Wahl:** Ubuntu LTS (Long-Term Support) ist die Standard-Wahl — stabiler Kernel, gute Docker-Unterstützung, große Community.

### 11.2 Server-Grundabsicherung

```bash
# 1. System aktualisieren
apt update && apt upgrade -y

# 2. Nicht-root User anlegen
adduser deploy
usermod -aG sudo deploy

# 3. SSH: Nur Key-basierte Authentifizierung
# /etc/ssh/sshd_config:
# PasswordAuthentication no
# PermitRootLogin no
systemctl restart sshd

# 4. UFW Firewall
ufw allow 22   # SSH
ufw allow 80   # HTTP (für ACME-Challenge)
ufw allow 443  # HTTPS
ufw enable

# 5. Automatische Sicherheitsupdates
apt install unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades
```

### 11.3 DNS-Konfiguration

**Minimal-Konfiguration:**

```
A     example.com      → 1.2.3.4 (Server-IP)
A     www.example.com  → 1.2.3.4
AAAA  example.com      → 2001:db8::1 (IPv6, falls vorhanden)
```

**TTL:** Für initiales Setup niedrig halten (300s = 5 Minuten) für schnelle Propagation. Nach erfolgreichem Deployment erhöhen (3600s = 1 Stunde).

**DNS-Propagation prüfen:**

```bash
dig +short example.com
nslookup example.com 8.8.8.8
```

### 11.4 Automatische TLS-Zertifikate (Let's Encrypt)

Let's Encrypt stellt kostenlose TLS-Zertifikate via ACME-Protokoll aus. Zertifikate sind 90 Tage gültig und müssen regelmäßig erneuert werden.

**ACME-Challenges:**

| Challenge    | Anforderung         | Wann nutzen?                       |
| ------------ | ------------------- | ---------------------------------- |
| **HTTP-01**  | Port 80 erreichbar  | Standardfall (Webserver läuft)     |
| **DNS-01**   | DNS-API-Zugriff     | Wildcard-Zertifikate, kein Port 80 |
| **TLS-ALPN** | Port 443 erreichbar | Wenn kein HTTP möglich             |

**Certbot mit nginx:**

```bash
# Certbot installieren
apt install certbot python3-certbot-nginx

# Zertifikat beantragen
certbot --nginx -d example.com -d www.example.com \
  --email admin@example.com --agree-tos --no-eff-email

# Automatische Erneuerung (Cron oder Systemd Timer)
certbot renew --dry-run  # Test
# Cron: 0 12 * * * /usr/bin/certbot renew --quiet
```

**Certbot in Docker (Webroot-Methode):**

```yaml
# Docker Compose: nginx + certbot
services:
  nginx:
    volumes:
      - certbot-challenges:/var/www/certbot # ACME-Challenge-Files
      - letsencrypt:/etc/letsencrypt # Zertifikate

  certbot:
    image: certbot/certbot
    volumes:
      - certbot-challenges:/var/www/certbot
      - letsencrypt:/etc/letsencrypt
    # Loop: Alle 24h Erneuerung prüfen
    command: >
      sh -c "while true; do
        certbot renew --webroot -w /var/www/certbot --quiet;
        sleep 24h;
      done"
```

### 11.5 Deployment-Workflow Self-Hosted

**Typischer Deployment-Prozess:**

```
1. Entwickler pushed auf main
   │
2. CI/CD Pipeline läuft (GitHub Actions)
   ├── Tests, Lint, Build
   └── Docker Image pushen → GHCR/Docker Hub
   │
3. Deployment auf Server
   ├── SSH auf Server
   ├── docker compose pull (neues Image holen)
   ├── docker compose up -d (ohne --build: Image wurde bereits gebaut)
   └── Health Check abwarten
```

**Automatisiertes Deployment via SSH:**

```yaml
# GitHub Actions Deployment Step
- name: Deploy
  uses: appleboy/ssh-action@master
  with:
    host: ${{ secrets.SERVER_HOST }}
    username: deploy
    key: ${{ secrets.SSH_PRIVATE_KEY }}
    script: |
      cd /opt/myapp
      docker compose pull
      docker compose up -d
      docker image prune -f  # Alte Images aufräumen
```

---

## 12. Referenzen

### Docker & Container

- [Docker Build Best Practices](https://docs.docker.com/build/building/best-practices/) — Multi-Stage Builds, Layer Caching, Security
- [Docker Security Best Practices](https://docs.docker.com/engine/security/) — Rootless, Capabilities, Seccomp
- [Docker Compose Dokumentation](https://docs.docker.com/compose/) — Syntax-Referenz, Health Checks, Networks
- [Distroless Images (Google)](https://github.com/GoogleContainerTools/distroless) — Minimale Runtime-Images ohne Shell

### 12-Factor & Cloud-Native

- [The Twelve-Factor App](https://12factor.net/) — Methodologie für Cloud-native Apps
- [CNCF Cloud Native Landscape](https://landscape.cncf.io/) — Übersicht aller Cloud-Native Tools

### Orchestrierung

- [K3s Dokumentation](https://docs.k3s.io/) — Leichtgewichtiges Kubernetes
- [Kubernetes Dokumentation](https://kubernetes.io/docs/) — Offizielles Kubernetes-Handbuch

### Reverse Proxy

- [nginx Dokumentation](https://nginx.org/en/docs/) — Konfigurationsreferenz
- [Traefik Dokumentation](https://doc.traefik.io/traefik/) — Traefik Routing, Middleware
- [Caddy Dokumentation](https://caddyserver.com/docs/) — Caddyfile Syntax, automatisches TLS

### CI/CD

- [GitHub Actions Dokumentation](https://docs.github.com/en/actions) — Workflow-Syntax, Actions-Marketplace
- [Martin Fowler: Continuous Delivery](https://martinfowler.com/bliki/ContinuousDelivery.html) — CD-Prinzipien
- [Martin Fowler: Deployment Pipeline](https://martinfowler.com/bliki/DeploymentPipeline.html) — Pipeline-Konzept

### Deployment-Strategien

- [Martin Fowler: Blue Green Deployment](https://martinfowler.com/bliki/BlueGreenDeployment.html) — Blue/Green Pattern
- [Martin Fowler: Canary Release](https://martinfowler.com/bliki/CanaryRelease.html) — Canary Pattern
- [Expand/Contract Pattern](https://martinfowler.com/bliki/ParallelChange.html) — Zero-Downtime Migrations

### Monitoring & Observability

- [Google SRE Book — Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/) — 4 Golden Signals, SLOs, Alerting
- [Prometheus Dokumentation](https://prometheus.io/docs/) — PromQL, Metrik-Typen
- [Grafana Loki](https://grafana.com/oss/loki/) — Log Aggregation
- [Brendan Gregg: USE Method](https://www.brendangregg.com/usemethod.html) — Infrastructure Monitoring
- [Tom Wilkie: RED Method](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/) — Service Monitoring

### Backup

- [Restic Dokumentation](https://restic.readthedocs.io/) — Modernes Backup-Tool
- [PostgreSQL Backup Dokumentation](https://www.postgresql.org/docs/current/backup.html) — pg_dump, pg_basebackup, WAL

### Self-Hosting

- [Let's Encrypt Dokumentation](https://letsencrypt.org/docs/) — ACME-Protokoll, Certbot
- [Certbot Dokumentation](https://certbot.eff.org/docs/) — Automatische TLS-Zertifikate

### Bücher & Weiterführendes

- **Gene Kim, Kevin Behr, George Spafford** (2013): _The Phoenix Project_ — DevOps-Roman, grundlegende Konzepte
- **Jez Humble & David Farley** (2010): _Continuous Delivery_ — CD-Praktiken, Deployment Pipelines
- **Google SRE** (2016): _Site Reliability Engineering_ — Monitoring, Incident Response, SLOs
- **Brendan Gregg** (2020): _Systems Performance_ — USE Method, Performance-Analyse
