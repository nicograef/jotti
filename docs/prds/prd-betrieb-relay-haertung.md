# PRD: Betriebs- und Relay-Härtung

> Umsetzungsplan: `docs/plans/plan-betrieb-relay-haertung.md`
> Referenz-Anforderung: Q-06 (`docs/anforderungen.md`) berührt die `/health`-Ausnahme
> Herkunft: extrahiert aus der ursprünglich gemischten Windows-PRD. Die
> klickbare Windows-EXE-Verpackung ist in `docs/prds/prd-windows-verpackung.md`
> abgetrennt; die lokale Transportverschlüsselung in
> `docs/prds/prd-lokale-tls-selbstsigniert.md` (Option 2) bzw.
> `docs/prds/prd-lokale-tls-vertrauenswuerdig.md` (Option 3).

## Problem Statement

Beim Entwurf des lokalen Windows-Betriebs traten mehrere **plattformunabhängige**
Mängel im Betrieb (dev, local, prod) zutage, die jotti unabhängig von jeder
Verpackung korrekter, sicherer und einfacher machen — ohne neue Domänenfeatures:

- **Der Bondruck (Relay) ist in keiner Umgebung funktionsfähig verdrahtet.**
  `RELAY_AUTH_TOKEN` fehlt in `.env.example` und in der Backend-Env aller drei
  Compose-Dateien. Das Relay erwartet seinen Token zudem heute als
  Pflicht-Kommandozeilen-Flag (`--token`), wodurch das Secret in der
  Prozessliste sichtbar wird.
- **Die Token-Prüfung ist unsicher.** Der Relay-Handler vergleicht den
  übergebenen Token direkt mit dem konfigurierten; ist der konfigurierte Token
  leer (heutiger Default), akzeptiert das Backend eine Anfrage mit
  `{"token":""}`.
- **Die Konfiguration ist inkonsistent.** `JWT_SECRET` ist ein Pflicht-Secret
  (fataler Abbruch beim Start, wenn nicht gesetzt), `RELAY_AUTH_TOKEN` hingegen
  still optional — zwei ungleich behandelte Secrets ohne sachlichen Grund.
- **Die Ersteinrichtung verlangt manuelle `.env`-Pflege.** Der Betreiber muss
  die Datei von Hand anlegen und Secrets selbst per `openssl` erzeugen.
- **Es gibt keinen nutzbaren Backend-Healthcheck.** `/health` hängt hinter der
  `PostMethodOnlyMiddleware` und ist damit für die Container-Orchestrierung
  (Docker-Healthcheck, `depends_on: service_healthy`) nicht per GET nutzbar.
- **Die drei Compose-Dateien duplizieren denselben Service-Block.**
  `postgres`, `migrate`, `backend` und `frontend` stehen in `docker-compose.yml`,
  `docker-compose.local.yml` und `docker-compose.prod.yml` mehrfach.

## Solution

jotti härtet Betrieb und Relay an genau diesen Punkten:

- **`RELAY_AUTH_TOKEN` wird ein vollwertiges Pflicht-Secret.** Das Backend lädt
  es über denselben Mechanismus wie `JWT_SECRET` (fataler Abbruch beim Start,
  wenn nicht gesetzt). Der Relay-Handler weist zusätzlich jede Anfrage ab, sobald
  der konfigurierte Token leer ist **oder** der übergebene Token nicht exakt
  passt (Defense-in-Depth).
- **Der Relay-Client wird vollständig über Umgebungsvariablen konfiguriert** und
  **ohne Kommandozeilen-Argumente** gestartet: `RELAY_AUTH_TOKEN` (Pflicht),
  `RELAY_BACKEND_URL` (Default `http://localhost/api`) und `RELAY_POLL_SECONDS`
  (Default `2`). Die Flags `--token`, `--backend`, `--poll` entfallen ersatzlos.
  Das deckt beide realen Szenarien mit **null Argumenten** ab.
- **`/health` wird ein Ops-Endpunkt** und bewusst von der „POST-only"-Guardrail
  ausgenommen — per **GET** erreichbar. Alle übrigen Routen bleiben strikt
  POST-only. Backend-Service erhält in `local` und `prod` einen
  Docker-Healthcheck; der reverse-proxy wartet auf `service_healthy`.
- **`make init` erzeugt eine fehlende `.env`** mit kryptografisch sicheren
  Zufallswerten (`POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN`) und
  sinnvollen Defaults. Der Vorgang ist **idempotent** und überschreibt eine
  vorhandene `.env` **nie**.
- **Eine gemeinsame `docker-compose.base.yml`** trägt die geteilten Services;
  `local` und `prod` werden schlanke Overrides. Der Dev-Stack bleibt
  eigenständig (Live-Reload-Images).

Der Betreiber richtet jotti damit mit einem einzigen Kommando ein, der Bondruck
funktioniert in allen Umgebungen, das Relay-Secret verlässt nie die
Prozessumgebung, und die Orchestrierung kann sich auf einen echten
Backend-Health-Status verlassen.

## User Stories

### Betreiber / Admin — Einrichtung & Betrieb

1. Als Betreiber möchte ich mit `make init` aus einem frischen Checkout eine
   vollständige `.env` mit sicheren Zufalls-Secrets erhalten, damit ich keine
   Variablen kennen und kein `openssl` aufrufen muss.
2. Als Betreiber möchte ich, dass ein erneuter `make init`-Aufruf meine
   vorhandene `.env` nie überschreibt, damit Daten und Zugänge über mehrere
   Festtage stabil bleiben.
3. Als Betreiber möchte ich, dass der Bondruck in dev, local und prod
   funktioniert, weil `RELAY_AUTH_TOKEN` überall aus der `.env` an das Backend
   übergeben wird.
4. Als Betreiber möchte ich, dass das Backend mit einer klaren Fehlermeldung
   abbricht, wenn `RELAY_AUTH_TOKEN` fehlt, damit ich eine Fehlkonfiguration
   sofort bemerke (analog `JWT_SECRET`).

### Servicebetrieb — Bondruck

5. Als Betreiber möchte ich das Relay ohne jedes Kommandozeilen-Argument starten,
   damit das Secret nicht in der Prozessliste auftaucht und der Start einfach
   bleibt.
6. Als Betreiber möchte ich das Relay im Standardfall „alles auf einem Rechner"
   ohne `RELAY_BACKEND_URL` betreiben (Default `http://localhost/api`), und im
   Szenario „Backend in Cloud/VPS, Relay lokal am Drucker" nur diese eine
   Variable auf die öffentliche Adresse setzen.

### Orchestrierung / Plattform

7. Als Plattform möchte ich `GET /health` mit `200` (bzw. `503` bei DB-Problemen)
   abfragen können, damit Docker einen echten Health-Status kennt und der
   reverse-proxy erst startet, wenn das Backend `healthy` ist.
8. Als Plattform möchte ich, dass alle übrigen Routen GET weiterhin ablehnen,
   damit die POST-only-Guardrail außer der bewussten `/health`-Ausnahme intakt
   bleibt.

### Wartung / Codequalität

9. Als Wartender möchte ich die geteilten Services genau einmal in einer Basis
   pflegen, damit `local` und `prod` nur noch ihre Abweichungen tragen und keine
   Service-Definition mehrfach gepflegt werden muss.

## Implementation Decisions

### Relay-Authentifizierung & -Konfiguration

- `RELAY_AUTH_TOKEN` ist die **einzige** Quelle des Relay-Secrets. Das Backend
  lädt ihn als Pflichtwert über denselben Mechanismus wie `JWT_SECRET` (fatal
  beim Start, wenn nicht gesetzt). Der Handler lehnt leere **oder** abweichende
  Tokens ab (Defense-in-Depth).
- Der Relay-Client wird vollständig über Umgebungsvariablen konfiguriert; alle
  Flags entfallen. Die Konfigurations-Auflösung wird in eine kleine, reine
  Funktion gezogen, sodass das `jotti-relay`-Modul seinen ersten Unit-Test
  erhält.
- `RELAY_AUTH_TOKEN` wird in der Backend-Env von `docker-compose.yml`,
  `docker-compose.local.yml` und `docker-compose.prod.yml` aus der `.env`
  übergeben und in `.env.example` dokumentiert.

### `/health` als Ops-Endpunkt

- `/health` wird **bewusst** von der „POST-only"-Guardrail ausgenommen und per
  GET erreichbar gemacht; alle übrigen Routen bleiben strikt POST-only. Diese
  Ausnahme wird im Handbuch als bewusste Ops-Entscheidung dokumentiert, damit
  sie nicht als Regelbruch missverstanden wird.
- Backend-Service erhält in `local` und `prod` einen Docker-Healthcheck (GET via
  BusyBox-`wget`); der reverse-proxy hängt an `condition: service_healthy`.

### `.env`-Erzeugung

- `make init` ruft `scripts/init-env.sh` (bash + `openssl`), erzeugt eine
  fehlende `.env` mit sicheren Zufallswerten und überschreibt eine vorhandene
  Datei nie. Die Init-Skripte und die Doku verweisen auf `make init` statt auf
  die manuelle `openssl`-Prozedur.

### Compose-Struktur

- Die für `local` und `prod` gemeinsamen Services wandern in eine neue
  `docker-compose.base.yml`; `local`/`prod` werden Overrides (`-f base -f …`).
  Alle Makefile-Targets und beide Init-Skripte werden umgestellt. Der Dev-Stack
  bleibt eigenständig.

## Testing Decisions

- **Relay-Konfiguration** (Unit-Tests, `jotti-relay`-Modul): Token gesetzt/fehlend,
  `RELAY_BACKEND_URL` Default/überschrieben.
- **Relay-Handler** (Unit-Tests): Token vorhanden → akzeptiert; leerer/falscher
  Token → abgelehnt.
- **`/health`** (Test): `GET /health` erlaubt, `GET` auf eine andere Route
  abgelehnt.
- **Nicht unit-getestet:** die fatalen Pflicht-Pfade beim Start (konsistent mit
  `JWT_SECRET`) und die Compose-Auflösung — letztere wird per
  `docker compose … config` gegen den Vor-Zustand abgeglichen.

## Out of Scope

- **Klickbare Windows-EXE-Verpackung** (`jotti-start.exe`, `jotti-relay.exe`,
  Release-ZIP) → `docs/prds/prd-windows-verpackung.md`.
- **Lokale Transportverschlüsselung (TLS)** → `docs/prds/prd-lokale-tls-selbstsigniert.md`
  (Option 2) und `docs/prds/prd-lokale-tls-vertrauenswuerdig.md` (Option 3).
- Neue Domänenfeatures; Änderungen an POST-only (außer der bewussten
  `/health`-GET-Ausnahme), Event-Sourcing oder Datenmodell.

## Further Notes

- **Zwei reale Relay-Szenarien (kein zweites Gerät als Sonderfall):** (1) alles
  auf einem Rechner (`RELAY_BACKEND_URL` lokal) und (2) Backend in Cloud/VPS +
  Relay lokal am Drucker (`RELAY_BACKEND_URL` zeigt auf die öffentliche Adresse).
  Beide nutzen denselben env-basierten Mechanismus ohne Flags; einen
  „Relay auf zweiter Station"-Override gibt es nicht.
- **Guardrail-Ausnahme:** Die GET-Freigabe für `/health` ist die einzige bewusste
  Aufweichung der „POST-only"-Regel und ausschließlich als Ops-Endpunkt gedacht.
- Die phasenweise Umsetzung (Relay-Token, Relay-Env, `make init`, `/health`,
  Compose-Entdoppelung) ist in `docs/plans/plan-betrieb-relay-haertung.md`
  detailliert.
