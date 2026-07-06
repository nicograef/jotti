# Plan: v0.14.0 — Breaking-Rest vor der Erstinstallation

> Ersetzt die Phasen 4, 5, 6 und 10 aus `plan-v0.14.0-vorabversion.md` (dessen Phasen 1–3 sind committed: 51e4ef5, d84f563, 4eee745; die alte Datei ist gelöscht, siehe Git-Historie). Befund-Details stehen im [Audit-Bericht](audit-v1.0.0.md); der Umsetzungsstatus wird nur noch hier geführt, nicht mehr doppelt im Audit abgehakt. Arbeitsdokument, nach Abarbeitung aus `docs/plans/` entfernen.

## Warum dieser Plan

Am 2026-07-07 installiert der erste Verein jotti produktiv. Mit der ersten echten Installation existiert die erste dauerhafte Prod-Datenbank: ab dann liegt Event-JSON in einem zehn Jahre aufzubewahrenden Journal, und `01_initial.up.sql` kann faktisch nicht mehr editiert werden (Schema-Änderungen nur noch als additive Migration). Der wirksame Freeze-Zeitpunkt für Schema und Event-Contract ist damit die Installation, nicht erst das v1.0.0-Tag.

Dieser Plan enthält ausschließlich, was vor diesem Zeitpunkt passieren muss:

- Event-Contract-Änderungen: B1-Rename, B2-Idempotenzfelder (persistieren im Journal)
- Schema-Änderungen: B5-Feinschliff, C9-Index (letzter Edit an `01_initial.up.sql`)
- Update-Pfad-Härtung, die in der ausgelieferten Version stecken muss: C4, C5 (der Windows-Starter aktualisiert sich nicht selbst, siehe Phase 4)

Alles andere (API-Feinschliff, Robustheit, Frontend, CI, Doku, CHANGELOG) ist nach der Installation ohne Update-Aufwand für Betreiber nachziehbar — Backend, Frontend und Relay werden bei jedem Update komplett gemeinsam getauscht. Das steht in [plan-v1.0.0-nacharbeit.md](plan-v1.0.0-nacharbeit.md).

Voraussetzung: die laufende Session (Phase 3 des alten Plans, DSFinV-K-Fixes A4/C6/C11/D15) ist abgeschlossen und committed.

## Ausführungsmodell (autonome Multi-Agent-Session)

- Phasen strikt sequenziell (sie teilen sich `01_initial.up.sql`, Golden-Tests und Frontend-Schemas); keine parallelen Implementierer auf demselben Working Tree.
- Pro Phase: frischer Implementierer-Subagent, unabhängiges Review-Gate, eigener `make verify`-Lauf, eigener Commit (Conventional Commits auf Englisch, keine AI-Trailer).
- Direkt-Commits auf `main` sind für die Ausführungssession ausdrücklich autorisiert (Ausnahme von der No-Auto-Commit-Regel); die globale Regel gilt danach unverändert weiter.
- Session endet mit release-fertigem `main`; Tag `v0.14.0` setzt und pusht Nico selbst (vorbereiteter Befehl am Ende).

## Risiken

- Golden-Tests werden in Phase 1 und 3 angefasst; jede Phase hinterlässt sie grün.
- `kassenjournal` ist append-only mit Owner-festem Trigger; neue Indexe (Phase 2, 3) sind unkritisch, aber die REVOKE/GRANT-Struktur bleibt unangetastet.
- Die Hardware-Verifikation zu A5 (QR auf echtem 80-mm-Drucker) bleibt offen; sie steht im [QA-Guide](guide-manuelle-qa-v1.0.0.md), Block 5.

---

## Phase 1: Event-Contract festzurren (B1, B3, D10)

### Kontext

- `backend/domain/kasse/bestellung.go:45` — JSON-Key `einzelpreis` ohne Cents-Suffix
- `backend/domain/kasse/event_json_contract_test.go:40-57` — Round-Trip über dieselben Structs, `bestellung-korrigiert:v1` fehlt
- `docs/language.md` — verbindliche Namenskonvention (Cents-Suffix)
- SQL-Konsumenten der Event-Keys (u. a. `kj_extract_umsatz_pro_steuersatz`) und Frontend-Zod-Schemas

### Was zu bauen ist

Rename `einzelpreis` zu `einzelpreisCents` durchgängig: Event-Data-Structs, SQL-Extraktoren, Response-DTOs, Frontend-Zod-Schemas, `docs/language.md` (kein Dual-Read, v0-Politik). Der Contract-Guard wird auf fixe JSON-Literale umgestellt: je Event-Typ ein eingefrorenes JSON-Beispiel, das unmarshalt und feldweise assertet wird; `bestellung-korrigiert:v1` sowie Tagesabschluss/Kassensturz und alle `*Id`-Felder werden gepinnt; EventType-Konstanten werden gegen die SQL-Literale geprüft; ein Meta-Test enumeriert alle Event-Typen. Wertkonstanten `einlage`/`entnahme` kommen in den Contract-Test (D10).

### Akzeptanzkriterien

- [ ] Kein JSON-Key `"einzelpreis"` mehr im Repo (nur `einzelpreisCents`), inklusive SQL und Frontend
- [ ] Contract-Test basiert auf fixen JSON-Literalen und würde einen Tag-Rename bemerken
- [ ] Alle Kassenjournal-Event-Typen sind abgedeckt, inklusive `bestellung-korrigiert:v1`; Meta-Test schlägt bei neuem, ungepinntem Event-Typ fehl
- [ ] `einlage`/`entnahme` als Wertkonstanten gepinnt
- [ ] `make verify` grün (inklusive Replay/rebuild-projections in den Integrationstests)

---

## Phase 2: Schema-Feinschliff, letzter Edit an 01_initial (B5, C9)

### Kontext

- `database/migrations/01_initial.up.sql` — letzter Edit vor dem De-facto-Freeze durch die Erstinstallation
- `database/migrations/README.md` — Migrations-Konvention
- `backend/sqlc/` — nach Typänderungen `make sqlc`
- `backend/sqlc/queries/tse_signaturauftraege.sql:78-85` — `GetTSESignaturQueueZustand` scannt die nie geleerte Tabelle voll (C9)

### Was zu bauen ist

CHECK-Constraints auf Geldspalten (`saldo_cents >= 0`, `gesamt_zahlungen_cents >= 0`, `preis_cents >= 0`); `kassenjournal.id` und `kassensitzungen.z_nr` auf GENERATED ALWAYS AS IDENTITY; `tse_signaturauftraege.transaktion_nummer` und `signatur_zaehler` auf BIGINT; partieller Index `(erledigt_am) WHERE status = 'erledigt'` fürs Queue-Monitoring (C9, vorgezogen aus dem Robustheits-Block, weil letzte Gelegenheit für `01_initial.up.sql`). Die ENUM-vs-TEXT+CHECK-Entscheidung wird in `database/migrations/README.md` begründet festgehalten (inklusive Zwei-Migrations-Muster für spätere ENUM-Erweiterungen).

### Akzeptanzkriterien

- [ ] Constraints, Typen und C9-Index wie oben in `01_initial.up.sql`; sqlc regeneriert, Go-Typen konsistent (int64 für BIGINT)
- [ ] Integrationstests auf frischer DB grün; `make rebuild-projections` läuft fehlerfrei
- [ ] ENUM-Entscheidung im Migrations-README dokumentiert
- [ ] `make verify` grün

---

## Phase 3: Idempotenz buchender Endpunkte (B2)

### Kontext

- `backend/api/kasse/direktverkauf/application/command.go:88` — frische Server-UUID je Request; die `verkaufId` ist bereits Pflichtfeld im Event-JSON und Teil des Subjects (`kassensitzung-{nr}/direktverkauf-{uuid}`), `UNIQUE (subject, version)` dedupliziert also schon, sobald die ID vom Client kommt
- `backend/api/kasse/tischgeschaeft/application/command.go:298` — Bestellung: reines Anhängen ohne Zustandsvalidierung, `GetMaxVersion` erst beim Write, kein natürlicher Schutz
- `backend/api/kasse/kassenfuehrung/application/command.go:206` — Geldtransit: dito
- `frontend/src/lib/Backend.ts:110` — zentrale Request-Schicht (Retry-/Doppel-Tap-Pfad)
- `database/migrations/01_initial.up.sql` — partielle UNIQUE-Indexe; Append-only-Trigger und REVOKE/GRANT unangetastet lassen
- Scope-Begründung: Zahlung, Storno (inklusive Warenrücknahme, die kein eigener Event-Typ ist, sondern `stornierung-erteilt:v1` mit `zahlungId`), Ausgabe und Umbuchung sind validierende Commands (Replay + Positions-Invariante + OCC gegen die gelesene Version); ein Retry scheitert dort sichtbar an der Invariante statt doppelt zu buchen. Der Kassensturz-/Tagesabschluss-Wiederanlauf (C8) ist nicht breaking und steht im Nacharbeit-Plan.

### Was zu bauen ist

Ein Mechanismus für alle drei Append-Endpunkte: clientseitig erzeugte Vorgangs-ID als Pflichtfeld im Request (UUID, validiert) und im Event — `verkaufId` beim Direktverkauf (Feld existiert, nur der Erzeuger wechselt vom Server zum Client), `bestellungId` und `geldtransitId` neu (Namen gegen `docs/language.md` prüfen). Deduplizierung je Event-Typ über einen partiellen UNIQUE-Index auf dem Event-JSON, zwingend type-gescoped (`(data->>'verkaufId') WHERE type = 'direktverkauf-getaetigt:v1'` usw.), weil `direktverkauf-storniert:v1` dieselbe `verkaufId` trägt; beim Direktverkauf greift daneben ohnehin `UNIQUE (subject, version)`, der explizite Index dokumentiert die Absicht und deckt zusätzlich den Retry über eine Kassensitzungsgrenze. Konfliktauflösung in allen drei Handlern gleich: bei einer Unique-Verletzung wird per Typ und Vorgangs-ID nachgeschlagen (Index-gedeckt); ein Treffer ergibt die idempotente Erfolgs-Antwort (gleiche ID gilt als derselbe Vorgang, der Payload wird nicht verglichen — als Vertrag im Handler dokumentieren), kein Treffer ist ein echter OCC-Konflikt und bleibt 409. Das Duplikat rollt die ganze Transaktion zurück, es entsteht also auch kein zweiter Signaturauftrag. Frontend erzeugt die IDs pro logischem Vorgang (nicht pro Retry); der Doppel-Submit-Schutz bleibt zusätzlich bestehen.

### Akzeptanzkriterien

- [ ] Zwei identische Direktverkauf-Requests erzeugen genau ein Event und einen Signaturauftrag; die zweite Antwort ist erfolgreich (Integrationstest)
- [ ] Gleiche Semantik für Bestellung und Geldtransit; ein echter OCC-Konflikt (andere Vorgangs-ID) antwortet weiterhin 409 (Integrationstests)
- [ ] Requests ohne oder mit ungültiger Vorgangs-ID werden als Validierungsfehler abgelehnt
- [ ] Frontend sendet die IDs pro logischem Vorgang; Doppel-Submit-Schutz bleibt zusätzlich bestehen
- [ ] Event-Contract-Test aus Phase 1 um `bestellungId`/`geldtransitId` ergänzt (`verkaufId` ist dort bereits gepinnt)
- [ ] `make verify` grün

---

## Phase 4: Update-Pfade entschärfen (C4, C5)

### Kontext

- `Makefile:156-158` — `prod-up` zieht Images und migriert ohne Backup/Downgrade-Guard
- `docker-compose.prod.yml` — `:-latest`-Fallback
- `windows/starter/update.go:27`, `windows/starter/main.go:100-112` — Starter startet ältere Exe klaglos gegen neuere Daten
- `scripts/prod-update.sh` — Referenz-`is_downgrade`-Logik
- Warum vor der Installation: der Starter aktualisiert sich nicht selbst (`update.go` gibt nur einen Hinweis aus, Update = manuell neues ZIP laden). Die Downgrade-Sperre schützt genau den Fall, dass nach einem späteren Update versehentlich die alte Exe gestartet wird — dafür muss sie in der Version stecken, die der Verein jetzt herunterlädt. Gleiches Argument für das Prod-Compose/Makefile auf dem Server des Vereins.

### Was zu bauen ist

`prod-up` wird als reiner Start-/Neustart-Weg positioniert: kein stilles Update mehr; ohne gesetzte Version bricht es ab statt `latest` zu ziehen; Make-Help und Doku nennen `prod-update` als einzigen Update-Weg. Der Windows-Starter spiegelt die `is_downgrade`-Sperre aus `prod-update.sh` und verweigert den Start einer älteren Version gegen neuere Daten.

### Akzeptanzkriterien

- [ ] `prod-up` ohne gepinnte Version schlägt mit klarer Meldung fehl; kein `:-latest` mehr im Prod-Compose
- [ ] Doku/Help: `prod-update` ist der einzige beworbene Update-Weg
- [ ] Starter-Downgrade-Test (ältere Version gegen neuere Datenversion wird verweigert)
- [ ] `make verify` grün

---

## Abschluss: Release-Schnitt v0.14.0

- [ ] Voller `make verify` und `make lint-backend-full` auf dem Release-Commit
- [ ] Checkboxen in diesem Plan abgeglichen (Audit bleibt unangetastet, es ist nur noch Register)
- [ ] Tag-Befehl ausgeben (`git tag -a v0.14.0 …` plus Push-Hinweis), nicht ausführen — Tag und Veröffentlichung macht Nico

Bewusst nicht Teil dieses Schnitts (nachziehbar, siehe Nacharbeit-Plan): CHANGELOG (C21), Beispiel-Versionen/Version-Kosmetik (C13), Doku-Konsistenz (C12, C16–C20, C23), API-Feinschliff (B4, B6 — muss vor dem v1.0.0-Tag, nicht vor der Installation).
