# Plan: Lokaler E2E-Bondruck-Test mit escpresso

> Source PRD: n/a (manueller Testplan)
> Voraussetzung: `docs/plans/plan-betrieb-relay-haertung.md` Phase 1+2 ist
> abgeschlossen (Relay env-basiert, Token Pflicht).

## Goal

Die komplette Druck-Pipeline (Bestellung/Verkauf → Outbox `druckauftraege` →
Relay → Drucker) lokal gegen den escpresso-Emulator E2E validieren. Nachweis
aller Bon-Typen, des kategorienbasierten Routings und des Lifecycle
`offen → gedruckt` — erst per Loopback, dann über einen echten LAN-Hop.

## Resolved decisions

- **Kein docker-compose-Edit.** Der lokale Stack (`make local-up`) wird
  unverändert genutzt. Das Relay läuft als nativer Prozess auf dem Host und
  verbindet sich über den nginx-Proxy (`http://localhost/api`, der Default von
  `RELAY_BACKEND_URL` nach Relay-Härtung).
- **Kein escpresso-Patch.** Ein Fenster, alle Bons. Routing-Beweis über
  `druckauftraege.ziel_ip` in der DB, nicht über separate Fenster.
- **Umlaute + `€` werden in escpresso falsch dargestellt** (CP858 → Windows-1252-
  Artefakt des Emulators). Auf dem echten MUNBYN korrekt, da jotti echte
  CP858-Bytes sendet. Layout/Format/Cut/Beep/Routing sind originalgetreu.
- **TLS ist für diesen Test irrelevant** (alles HTTP im lokalen Stack).

## Inventory

- `docker-compose.local.yml` — Produktions-naher Stack mit nginx (Port 80),
  Backend, Frontend, Postgres.
- `reverse-proxy/nginx.local.conf:31-39` — `/api/*` → `backend:3000/`.
- `cmd/relay/main.go` — env-basiert nach Härtung: `RELAY_AUTH_TOKEN` (Pflicht),
  `RELAY_BACKEND_URL` (Default `http://localhost/api`), `RELAY_POLL_SECONDS`
  (Default `2`).
- `backend/api/relay/http/handler.go:41-50` — `isRelayTokenValid()`:
  Defense-in-Depth (leerer + falscher Token → Ablehnung).
- `backend/api/bondruck/application/arbeitsbon_policy.go:97-120` — Routing nach
  Kategorie, Skip bei leerer IP, Beep nur `essen`.
- `backend/api/bondruck/application/escpos/formatter.go` + `constants.go` —
  ESC/POS-Rendering, CP858, Paper-Cut, Buzzer.
- `backend/api/settings/http/command_handler.go:37-55` — Bondruck-Einstellungen
  (Kassenbeleg-IP, Direktverkauf-Modus, Abholbon-IP).
- `backend/api/druckstation/http/handler.go:70-107` — Druckstationen CRUD
  (Kategorie, IP, Bonmodus).
- `database/migrations/01_initial.up.sql:219-237` — `druckstationen` (Seeds:
  alle leer, `pro_position`).
- `database/migrations/01_initial.up.sql:388-408` — `bondruck_einstellungen`
  (Seed: leer, `kein_bon`).
- `database/seed.sql:167-171` — Kassensitzung z_nr=3, Status `offen`.
- `.env.example` — alle benötigten Env-Vars inkl. `RELAY_AUTH_TOKEN`.

## Bon-Erkennung im escpresso-Fenster

| Bon-Typ     | Erkennungsmerkmal                                                                        |
| ----------- | ---------------------------------------------------------------------------------------- |
| Arbeitsbon  | `== Tisch N ==`, Artikel doppelhoch, **keine Preise**. Essen zusätzlich mit Buzzer-Beep. |
| Abholbon    | Kopf „Direktverkauf", **keine Preise**.                                                  |
| Kassenbeleg | Vereinsname/Adresse, Beleg-Nr., **Preise**, Gesamtbetrag, Zahlungsart, MwSt-Aufstellung. |

---

## Phase 1: Vorbereitung

### What to do

Den lokalen Stack mit Seed-Daten starten, escpresso als Emulator öffnen,
`RELAY_AUTH_TOKEN` in `.env` setzen, Relay-Endpoint per curl verifizieren.

### Steps

1. **escpresso starten:** `escpresso` in einem separaten Terminal. Erwartung:
   `TCP Server listening on 0.0.0.0:9100`.

2. **`.env` prüfen/ergänzen:** Die `.env` muss `RELAY_AUTH_TOKEN` enthalten
   (z. B. `RELAY_AUTH_TOKEN=devrelay`). Falls `make init` genutzt wurde, steht
   dort bereits ein zufälliger Wert — dann diesen für das Relay notieren.

3. **Lokalen Stack starten + Seed:**

   ```bash
   make local-up
   PG_CONTAINER=jotti-postgres-local make seed
   docker exec jotti-backend-local jotti rebuild-projections
   ```

   App unter http://localhost. Login: `thomas` / `jotti123`.

4. **Sanity-Check Relay-Endpoint:**
   ```bash
   curl -s -X POST http://localhost/api/relay/poll \
     -H 'Content-Type: application/json' \
     -d '{"token":"<RELAY_AUTH_TOKEN aus .env>"}'
   ```
   Erwartung: `200` mit `{"auftraege":[]}`.

### Acceptance criteria

- [ ] escpresso lauscht auf `:9100`.
- [ ] `make local-up` läuft ohne Fehler; App erreichbar auf Port 80.
- [ ] Seed-Kassensitzung (z_nr=3) ist offen.
- [ ] Relay-Endpoint antwortet `200` mit leerem Array.

---

## Phase 2: Loopback E2E — alle Bon-Szenarien

### Context

escpresso bindet `0.0.0.0:9100` und antwortet damit auf **jeder**
Loopback-Adresse (`127.0.0.x`). Jeder Drucker-Kategorie wird eine eigene
Loopback-IP zugewiesen → alles rendert in einem Fenster, aber
`druckauftraege.ziel_ip` beweist die kategorienbasierte Zuordnung.

### Steps

5. **`/admin/einstellungen` → Bondruck-Einstellungen:**
   - Betreiber-Stammdaten ausfüllen (_Pflicht für Kassenbeleg_).
   - `kassenbeleg_drucker_ip` = `127.0.0.4`
   - `direktverkauf_modus` = `abholbon`
   - `abholbon_drucker_ip` = `127.0.0.5`

6. **`/admin/druckstationen`:**
   - `essen` = `127.0.0.1`, Bonmodus `pro_position`
   - `getraenk` = `127.0.0.2`, Bonmodus `pro_position`
   - `sonstiges` = `127.0.0.3`, Bonmodus `pro_position`

7. **Relay starten:** In neuem Terminal:

   ```bash
   cd cmd/relay && \
    RELAY_AUTH_TOKEN=<Token aus .env> \
    RELAY_BACKEND_URL=http://localhost/api \
    RELAY_POLL_SECONDS=2 \
    go run .
   ```

   Erwartung: `jotti Print-Relay gestartet | Backend: http://localhost/api | Poll: 2s`

8. **Szenario A — Arbeitsbon (pro_position):**
   Als Service unter `/service/tische/<id>` Artikel aus Kategorie _essen_ +
   _getraenk_ bestellen.
   - Erwartung: zwei Arbeitsbons im escpresso-Fenster (Essen mit Beep).
   - DB: `ziel_ip` = `127.0.0.1` (essen) vs. `127.0.0.2` (getraenk).

9. **Szenario B — Bonmodus pro_bestellung:**
   Druckstation `getraenk` auf `pro_bestellung` umstellen, erneut bestellen.
   - Erwartung: ein Sammelbon für alle Getränke-Positionen (statt je einer).

10. **Szenario C — Kassenbeleg (Tisch-Zahlung):**
    Bestellung ausgeben → kassieren (`/service/zahlung-kassieren`). Danach in
    der Tisch-Historie die Zahlung öffnen → „Beleg drucken".
    - Erwartung: Kassenbeleg mit Preisen, Vereinsname, `ziel_ip=127.0.0.4`.

11. **Szenario D — Direktverkauf (Abholbon + Kassenbeleg):**
    `/service/direktverkauf` → Verkauf abschließen.
    - Erwartung: Abholbon auf `127.0.0.5`.
      Danach Drucker-Icon in der Direktverkauf-Historie.
    - Erwartung: Kassenbeleg auf `127.0.0.4`.

12. **Szenario E — Routing/Skip bei leerer IP:**
    Druckstation `getraenk`: IP leeren (leer lassen, speichern).
    Bestellung mit Getränk-Positionen aufgeben.
    - Erwartung: **kein** Druckauftrag für Getränke in der DB.
      Danach IP zurücksetzen.

### Acceptance criteria

- [ ] Arbeitsbons (pro_position): je ein Bon pro Position, korrekte `ziel_ip`.
- [ ] Arbeitsbons (pro_bestellung): ein Sammelbon pro Bestellung/Kategorie.
- [ ] Arbeitsbon Essen: Buzzer-Beep hörbar/erkennbar.
- [ ] Kassenbeleg: Preise, Betreiber-Stammdaten, Beleg-Nr. vorhanden.
- [ ] Abholbon: korrekt auf Abholbon-Drucker-IP geroutet.
- [ ] Skip: leere IP erzeugt keinen Druckauftrag.
- [ ] DB-Lifecycle: alle Jobs `offen → gedruckt` mit `gedruckt_am` gesetzt.
- [ ] Relay-Logs: „Auftrag N erfolgreich gedruckt auf <ip>" / „M Auftraege quittiert".

---

## Phase 3: LAN-Hop-Validierung

### Context

Dieser Schritt validiert, dass das Relay über das Netzwerk drucken kann —
der reale Anwendungsfall, wenn Drucker an einem anderen Gerät hängen.

### Steps

13. escpresso auf einem zweiten Gerät (VM, zweiter Rechner, WSL-Host) starten.
    LAN-IP notieren (z. B. `192.168.1.50`), Firewall für TCP `9100` inbound
    freigeben.

14. In der Admin-UI **alle** Drucker-IPs auf die LAN-IP des zweiten Geräts
    setzen.

15. Relay bleibt auf dem Host mit `RELAY_BACKEND_URL=http://localhost/api`
    (Default, keine Änderung nötig). Das Relay verbindet sich zum lokalen
    Backend und sendet Druckdaten über das Netz an die LAN-IP.

16. Szenarien A + C wiederholen.
    - Erwartung: Bons erscheinen auf dem LAN-Emulator, Jobs → `gedruckt`.

### Acceptance criteria

- [ ] Druckdaten erreichen den Emulator auf der LAN-IP (TCP 9100).
- [ ] Jobs wechseln auf `gedruckt` in der DB.
- [ ] Relay-Logs bestätigen erfolgreichen Druck auf die LAN-IP.

---

## Verifikation (zusammengefasst)

| Kanal      | Prüfung                                                                                                                           |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Visuell    | escpresso-Fenster: Struktur, Ausrichtung, Bold, Doppelgröße, Paper-Cut.                                                           |
| DB         | `SELECT id, ziel_ip, bon_art, status, referenz, erstellt_am, gedruckt_am FROM druckauftraege ORDER BY id;` — Lifecycle + Routing. |
| Relay-Logs | „Auftrag N erfolgreich gedruckt auf \<ip\>" / „M Auftraege quittiert".                                                            |
| Negativ    | Leere Kategorie-IP → kein Job für diese Kategorie.                                                                                |

DB-Shell für lokalen Stack: `docker exec -it jotti-postgres-local psql -U admin -d jotti`

## Relevante Dateien

- `cmd/relay/main.go` — env-basiert: `RELAY_AUTH_TOKEN`, `RELAY_BACKEND_URL`,
  `RELAY_POLL_SECONDS`. TCP-Verbindung `ip:9100`.
- `backend/api/bondruck/application/escpos/formatter.go` + `constants.go` —
  ESC/POS-Rendering, CP858, Paper-Cut, Buzzer.
- `backend/api/bondruck/application/arbeitsbon_policy.go` — Routing je Kategorie,
  Skip bei leerer IP, Beep nur `essen`.
- `database/migrations/01_initial.up.sql` — Tabellen `druckauftraege`,
  `druckstationen`, `bondruck_einstellungen`.

## Optionale Erweiterungen (nicht blockierend)

1. **Umlaut-Treue:** Falls die korrekte Glyph-Darstellung im Emulator wichtig
   wird → ~15-Zeilen-Patch an escpresso (Code page 19 via `oem_cp` als CP858
   dekodieren). Empfehlung: nur bei Bedarf.
2. **Mehrere Fenster:** `ESCPRESSO_LISTEN`-env für getrennte Ports/Fenster pro
   Kategorie — visualisiert das Routing physisch. Nur bei Bedarf.
3. **Byte-genaue Asserts:** Der bestehende `TestRelayPollQuittierenFlow`
   (Integrationstest) ist die automatisierte Ergänzung zum visuellen GUI-Check.
