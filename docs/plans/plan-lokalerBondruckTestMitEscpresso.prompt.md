# Plan: Lokaler E2E‑Bondruck‑Test mit escpresso

Die komplette Druck‑Pipeline (Bestellung/Verkauf → Outbox `druckauftraege` → Relay → Drucker) lokal gegen **einen** escpresso‑Emulator fahren. Kein Patch an escpresso. Phase 1 läuft per Loopback auf dem Dev‑Rechner, Phase 2 validiert den echten LAN‑Hop. Einzige Infrastruktur‑Änderung: eine kleine Ergänzung an `docker-compose.yml`.

**Kernidee Routing‑Nachweis:** escpresso bindet `0.0.0.0:9100` und antwortet damit auf *jeder* Loopback‑Adresse. Du gibst jedem Drucker eine eigene `127.0.0.x`‑IP → alles rendert in einem Fenster, aber `druckauftraege.ziel_ip` beweist die kategorienbasierte Zuordnung (Essen vs. Getränke). Das deckt „verschiedene Drucker" ohne Patch ab.

**So erkennst du die Bons im einen Fenster:** Arbeitsbon = `== Tisch N ==`, Artikel doppelhoch, keine Preise (Essen zusätzlich mit Beep); Abholbon = Kopf „Direktverkauf", keine Preise; Kassenbeleg = Vereinsname/Adresse, Beleg‑Nr., **Preise**, Gesamtbetrag, Zahlungsart.

## Phase 0 — Vorbereitung

1. **escpresso bauen/starten:** System‑Deps (`libgtk-3-dev libxcb-render0-dev libxcb-shape0-dev libxcb-xfixes0-dev libxkbcommon-dev libssl-dev`), dann `cd ~/r/escpresso && cargo run --release`. Es öffnet ein Fenster und lauscht auf TCP `:9100`. *(Optional `DEBUG=1` → schreibt `escpos_capture.raw` für Byte‑Diffs.)*
2. **`docker-compose.yml` → Service `backend-dev`:** in `environment:` ergänzen `RELAY_AUTH_TOKEN: devrelay` und einen Host‑Port `ports: ["3000:3000"]`. *(Dev‑Stack, risikoarm; alternativ override‑Datei oder Vite‑Proxy — siehe Optionen.)*
3. `make dev` (rekreiert `backend-dev`), dann `make seed`. App unter http://localhost, Login `thomas` / `jotti123`.
4. **Sanity‑Check Relay‑Endpoint:** `curl -X POST http://localhost:3000/relay/poll -H 'Content-Type: application/json' -d '{"token":"devrelay"}'` → erwartet `200` mit `{"auftraege":[]}`.

## Phase 1 — Loopback E2E (alle Szenarien)

5. **`/admin/einstellungen`:** Betreiber‑Stammdaten ausfüllen (*Pflicht für Kassenbeleg*); `kassenbeleg_drucker_ip=127.0.0.4`; `direktverkauf_modus=abholbon`; `abholbon_drucker_ip=127.0.0.5`.
6. **`/admin/druckstationen`:** `essen=127.0.0.1` (pro_position), `getraenk=127.0.0.2` (pro_position), `sonstiges=127.0.0.3`.
7. Kassensitzung ist durch den Seed bereits offen. **Relay starten:** `cd cmd/relay && go run . --backend=http://localhost:3000 --token=devrelay --poll=2`.
8. **A — Arbeitsbon:** als Service `/service/tische/<id>` → Artikel aus *essen* + *getraenk* bestellen → zwei Bons erscheinen (Essen mit Beep), DB‑`ziel_ip` zeigt `127.0.0.1` vs. `127.0.0.2`.
9. **B — Bonmodus:** `getraenk` → `pro_bestellung` umstellen, erneut bestellen → ein Sammelbon statt Einzelbons. *(hängt von Schritt 8 ab)*
10. **C — Kassenbeleg (Tisch):** `/service/zahlung-kassieren` → in der TischHistorie die Zahlung öffnen → „Beleg drucken" → Kassenbeleg mit Preisen, `ziel_ip=127.0.0.4`.
11. **D — Direktverkauf:** `/service/direktverkauf` → Verkauf abschließen → Abholbon (`127.0.0.5`); danach Drucker‑Icon in der DirektverkaufHistorie → Kassenbeleg (`127.0.0.4`). *(optional `direktverkauf_modus=an_stationen` zum Routing nach Kategorie)*
12. **Routing/Skip‑Test:** `getraenk`‑IP leeren → Getränke‑Positionen erzeugen keinen Job (in DB prüfen) → danach zurücksetzen.

## Phase 2 — LAN‑Hop‑Validierung

13. escpresso auf einem zweiten WLAN‑Gerät/VM starten, LAN‑IP notieren (z. B. `192.168.1.50`), Firewall für TCP `9100` inbound freigeben.
14. In der Admin‑UI **alle** Drucker‑IPs auf `192.168.1.50` setzen.
15. Relay erreicht das Backend: auf der Dev‑Box `--backend=http://localhost:3000`; auf einem anderen Gerät `--backend=http://<devbox-LAN-IP>:3000` (Port 3000 ist freigegeben, Backend bindet `0.0.0.0:3000`).
16. Szenarien A + C wiederholen → Bons erscheinen auf dem LAN‑Emulator, Jobs wechseln auf `gedruckt`. Damit ist „Relay im lokalen Netzwerk" nachgewiesen.

## Verifikation

1. **Visuell:** escpresso‑Fenster zeigt je Bon Struktur, Ausrichtung, Bold, Doppelgröße, Paper‑Cut.
2. **DB:** `make db-shell` → `SELECT id, ziel_ip, bon_art, status, referenz, erstellt_am, gedruckt_am FROM druckauftraege ORDER BY id;` → Lifecycle `offen → gedruckt`, `ziel_ip` = konfigurierte Kategorie‑IP (Routing‑Beweis).
3. **Relay‑Logs:** „Auftrag N erfolgreich gedruckt auf <ip>" / „M Auftraege quittiert".
4. **Negativ:** leere Kategorie‑IP → kein Job für diese Kategorie.

## Relevante Dateien

- `docker-compose.yml` — einzige Änderung (`backend-dev`: env + ports).
- `cmd/relay/main.go` — Run‑Target, Flags `--backend/--token/--poll`, `checkPrinter`/`sendToPrinter` (TCP `ip:9100`).
- `backend/api/bondruck/application/escpos/formatter.go` + `constants.go` — Aussehen/CP858/Cut/Beep.
- `backend/api/bondruck/application/arbeitsbon_policy.go` — Routing je Kategorie, Skip bei leerer IP, Beep nur `essen`.
- `database/migrations/01_initial.up.sql` — Tabellen `druckauftraege`/`druckstationen`/`bondruck_einstellungen`.

## Entscheidungen

- **Kein escpresso‑Patch:** ein Fenster. **Umlaute + `€` werden in escpresso falsch dargestellt** (CP858→Windows‑1252‑Artefakt des Emulators; auf dem echten MUNBYN korrekt, da jotti echte CP858‑Bytes sendet). Layout/Format/Cut/Beep/Routing sind originalgetreu.
- **Backend‑Erreichbarkeit:** direkter Edit der Dev‑`docker-compose.yml` (Token + Port 3000), Relay → `http://localhost:3000`.
- **Umfang:** erst Loopback (Phase 1), dann LAN‑Hop (Phase 2).

## Weitere Überlegungen (optional, nicht blockierend)

1. **Umlaut‑Treue später:** Falls die korrekte Glyph‑Darstellung doch wichtig wird → ~15‑Zeilen‑Patch an escpresso (Code page 19 via `oem_cp` CP858 dekodieren + `ESCPRESSO_LISTEN`‑env für mehrere Fenster). Empfehlung: nur bei Bedarf.
2. **Port 3000 nicht freigeben:** Alternative ohne Host‑Port → Relay `--backend=http://localhost/api` (Vite entfernt `/api`). Empfehlung: Port freigeben (vereinfacht auch Phase 2).
3. **Byte‑genaue Asserts:** Der bestehende `TestRelayPollQuittierenFlow` (Integrationstest) ist die automatisierte Ergänzung zum visuellen GUI‑Check.
