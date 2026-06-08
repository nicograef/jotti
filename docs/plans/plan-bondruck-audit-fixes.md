# Plan: Bondruck Audit-Fixes

> Source PRD: n/a (abgeleitet aus dem Code-Audit „Konzept & Implementierung von Bondruck")

## Goal

Alle acht im Bondruck-Audit gefundenen Befunde beheben — plus die Beobachtung
zum fehlenden Frontend-Fehler-Mapping. Jeder Befund wird in einer eigenen,
isoliert verifizierbaren Phase bearbeitet. Reihenfolge: nach Befund-Nummer
(#1–#8), die Beobachtung als Phase 9.

Jede Phase ist abgeschlossen, sobald die Akzeptanzkriterien erfüllt sind und
`make check` grün ist. Phasen werden **einzeln und iterativ** umgesetzt; eine
Phase wird nicht begonnen, bevor die vorherige grün ist.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Outbox-Transaktionalität (#1):** Das `bestellung-aufgenommen:v1`-Event und
  die daraus erzeugten Arbeitsbon-Druckaufträge werden in **genau einer**
  DB-Transaktion geschrieben (echter transaktionaler Outbox). Die Aufträge
  werden aus dem Event **mit bereits generierter ID** gebildet (die `referenz`
  hängt von der Event-ID ab). Die Transaktionsverwaltung darf nicht in die
  Domain-Schicht lecken.
- **Codepage (#2):** Ziel-Codepage ist **CP858** (deckt ä/ö/ü/Ä/Ö/Ü/ß + €).
  Transkodierung UTF-8 → CP858 über `golang.org/x/text/encoding/charmap`
  (bereits indirekt im Module-Graph → wird zur direkten Dependency, kein neuer
  externer Download). Die Drucker-Codepage wird per `ESC t n` gesetzt; nur
  sichtbarer Text wird transkodiert, **niemals** ESC/POS-Steuerbytes.
- **IP-Validierung (#3):** Beide Drucker-IP-Felder (Arbeitsbon-Druckstation und
  Kassenbeleg-Drucker) werden über die **nativen IPv4-Validatoren der
  Validierungsbibliotheken** geprüft (Backend zog `.IPv4()`, Frontend Zod
  `z.ipv4()`, beide `net.ParseIP`-basiert) statt über einen eigenen
  `netip`-Validator; leere IP bleibt erlaubt (= „kein Drucker"). Der bestehende
  `netip`-Backstop in `NewBondruckEinstellungen` bleibt als Domain-Invariante.
- **Scope:** Reine Korrektheits-, Konsistenz- und Aufräum-Fixes am bestehenden
  Bondruck. **Keine** neuen Features (kein Direktverkauf-Bondruck, kein KDS,
  keine TSE-Felder).

## Inventory

Relevante bestehende Dateien (Pfad:Zeilen):

- `backend/api/table/application/command.go:139-155` — Helper
  `enqueueArbeitsbonDruckauftraege` (separate Tx nach Event-Commit)
- `backend/api/table/application/command.go:392-407` — `BestellungAufnehmen`:
  `writeEvent` → `evt.ID = eventID` → Enqueue
- `backend/repository/kassenjournal_repo/repo.go:31-81` — `WriteEvent` besitzt
  aktuell die Tx-Grenze (Event + Projektion in einer Tx, committet intern)
- `backend/repository/druckauftrag_repo/repo.go:31-59` —
  `EnqueueDruckauftraege` (eigene Tx)
- `backend/api/bondruck/application/escpos/constants.go:1-28` — ESC/POS-Codes,
  kein Codepage-Set; ungenutzte Konstanten
- `backend/api/bondruck/application/escpos/formatter.go:12` — `lineWidth = 48`;
  `truncate` (`:198-204`) und `wrapLine` (`:206-230`) rechnen mit `len()` (Bytes)
- `backend/api/bondruck/application/arbeitsbon_policy.go:105-119` —
  `parseTischName` mit totem Legacy-Zweig (`tisch:{id}`)
- `backend/domain/kasse/subject.go:13-16` — erzeugt **nur** das Format
  `kassensitzung-{nr}/tisch-{id}`
- `backend/api/druckstation/http/handler.go:70-82` — loses IPv4-Regex (akzeptiert
  `999.999.999.999`)
- `backend/domain/settings/bondruck_einstellungen.go:13-35` — strikte
  `netip`-Validierung + `NewBondruckEinstellungen` (untested)
- `backend/api/settings/http/command_handler.go:43-47` — loses IPv4-Regex +
  Domain-Backstop
- `backend/repository/druckstation_repo/repo.go:42-43` — irreführender Kommentar
  („Wird vom Relay-Service verwendet")
- `backend/api/druckstation/application/query.go:11-29` — `GetAlleDruckstationen`
  gibt Repository-Typ `[]druckstation_repo.Druckstation` nach außen
- `cmd/relay/main.go:159-167` — `checkPrinter` hardcodet `0x10,0x04,0x04`
  (separates Go-Modul, kann `escpos`-Konstanten nicht importieren)
- `frontend/src/service/components/table/TischHistorie.tsx:104-114` —
  `belegDrucken`-`byCode` ohne `kasse_nicht_geoeffnet`
- `backend/api/table/http/command_handler.go:380-403` — `KassenbelegDruckenHandler`
  kann `kasse_nicht_geoeffnet` liefern

## Resolved decisions

- **Outbox (#1):** Transaktionaler Outbox — Event + Druckauftrag in einer Tx
  (nicht: Log-only-Mitigation).
- **Codepage (#2):** CP858.
- **Transkodierung (#2):** `golang.org/x/text/encoding/charmap` (nicht: handgeschriebene Mapping-Tabelle).
- **Beobachtung:** Als eigene Phase 9 aufnehmen.
- **Reihenfolge:** Nach Befund-Nummer.

## Open questions / Risks

- **ESC `t n`-Index für CP858 (#2):** Der exakte Codepage-Index ist
  druckerabhängig und lässt sich in dieser Session **nicht** auf der echten
  Hardware (MUNBYN ITPP047P-UE) verifizieren. Index als benannte Konstante
  führen und am Gerät gegenprüfen. Phase 2 wird „blind implementiert, später am
  Drucker verifiziert".
- **Event-ID-Abhängigkeit (#1):** Die `referenz` (`bestellung-aufgenommen:{id}`)
  braucht die beim INSERT generierte Event-ID. Der Tx-Mechanismus muss die
  Aufträge **nach** dem Event-INSERT, aber **vor** dem Commit erzeugen (z. B.
  über einen Callback, der das Event-mit-ID erhält). Ein simples
  „Aufträge in `WriteEvent` hineinreichen" funktioniert wegen dieser
  Abhängigkeit nicht direkt.
- **Dependency-Promotion (#2):** `golang.org/x/text` wandert von `// indirect`
  zu direkter Dependency. Kein neuer externer Download, da bereits im Graph.

---

## Phase 1: Transaktionaler Arbeitsbon-Outbox (#1)

### Context

- `backend/api/table/application/command.go:139-155` — Helper läuft in separater Tx
- `backend/api/table/application/command.go:392-407` — Event-Write, dann Enqueue
- `backend/repository/kassenjournal_repo/repo.go:31-81` — Event-Repo besitzt die Tx-Grenze
- `backend/repository/druckauftrag_repo/repo.go:31-59` — Enqueue mit eigener Tx

### What to build

Das `bestellung-aufgenommen:v1`-Event und die daraus erzeugten
Arbeitsbon-Druckaufträge in **einer** Transaktion schreiben. Die Aufträge
werden aus dem Event mit generierter ID gebildet und im selben Tx-Kontext
in `druckauftraege` eingefügt. Schlägt der Auftrags-Insert fehl, wird das Event
zurückgerollt — keine Bestellung ohne Bon, kein Bon ohne Bestellung. Damit
entfällt das Risiko, dass ein Retry nach committetem Event per OCC eine zweite
Bestellung erzeugt.

Mechanismus ist Sache der Umsetzung (z. B. ein outbox-fähiger Write-Pfad im
Event-Repo, der einen Callback `event.Event → []NeuerDruckauftrag` erhält);
die Tx-Verwaltung bleibt in der Repository-Schicht, die Domain bleibt
HTTP-/DB-frei.

### Acceptance criteria

- [x] Event-INSERT und Druckauftrag-INSERT teilen dieselbe Transaktion.
- [x] Fehler beim Auftrags-INSERT rollt das Event zurück (verifiziert durch Test).
- [x] `referenz` nutzt weiterhin die generierte Event-ID.
- [x] Kein zweiter Tx-Roundtrip mehr nach dem Event-Commit (der separate
      `enqueueArbeitsbonDruckauftraege`-Pfad ist aufgelöst).
- [x] Bestehende Tisch-/Bestellungs-Tests bleiben grün; neuer Test deckt den
      Rollback-Fall ab.
- [x] `make check` grün.

---

## Phase 2: CP858-Transkodierung + runenbasierte Breite (#2)

### Context

- `backend/api/bondruck/application/escpos/constants.go:4` — `Init`, kein Codepage-Set
- `backend/api/bondruck/application/escpos/formatter.go:12,198-230` —
  `lineWidth`, `truncate`, `wrapLine` rechnen mit Bytes
- `cmd/relay/main.go:185-196` — Relay sendet Payload-Bytes unverändert

### What to build

Sichtbaren Text vor dem Senden von UTF-8 nach CP858 transkodieren (über
`golang.org/x/text/encoding/charmap.CodePage858`) und die Drucker-Codepage per
`ESC t n` setzen. Nur Text wird transkodiert — ESC/POS-Steuerbytes bleiben
unangetastet. `truncate`/`wrapLine` zählen **Runen** statt Bytes, sodass
Umlaute die Spaltenbreite korrekt belegen und `truncate` nie mitten in einer
Rune schneidet. Gilt für alle drei Bon-Typen (`FormatPositionBon`,
`FormatSammelBon`, `FormatKassenbeleg`).

### Acceptance criteria

- [x] Drucker-Codepage wird per `ESC t n` (CP858) gesetzt; Index als benannte Konstante.
- [x] Umlaute (ä/ö/ü/Ä/Ö/Ü/ß) und € werden als CP858-Einzelbytes ausgegeben.
- [x] ESC/POS-Steuerbytes werden **nicht** transkodiert.
- [x] `truncate`/`wrapLine` arbeiten runenbasiert; `truncate` erzeugt nie ungültiges Encoding.
- [x] Unit-Test: Text mit Umlauten + € → erwartete CP858-Bytes; `truncate` auf
      Umlaut-Grenze bleibt gültig.
- [x] Risiko dokumentiert: exakter `ESC t`-Index am MUNBYN-Gerät zu verifizieren.
- [x] `make check` grün.

---

## Phase 3: Einheitliche strikte IP-Validierung (#3)

### Context

- `backend/api/druckstation/http/handler.go:70-82` — loses Regex
- `backend/domain/settings/bondruck_einstellungen.go:13-24` — strikte `netip`-Prüfung
- `backend/api/settings/http/command_handler.go:43-47` — loses Regex + Backstop

### What to build

Die Arbeitsbon-Druckstation-IP strikt validieren wie die Kassenbeleg-Drucker-IP.
Statt eines eigenen `netip`-Validators werden die **nativen IPv4-Validatoren der
bereits genutzten Validierungsbibliotheken** verwendet (Entscheidung des Nutzers):
Backend zog `z.String().IPv4(...)`, Frontend Zod `z.ipv4(...)`. Beide lehnen
`999.999.999.999` & Co. ab (`net.ParseIP`-basiert); leere IP bleibt erlaubt
(`.Optional()` bzw. `.or(z.literal(''))`). Der bestehende `netip`-Backstop in
`NewBondruckEinstellungen` bleibt als Domain-Invariante erhalten. Backend bleibt
Single Source of Truth.

### Acceptance criteria

- [x] Druckstation-IP wird strikt (zog `.IPv4()`) validiert; ungültige Oktette werden abgelehnt.
- [x] Leere IP (= kein Drucker) bleibt zulässig.
- [x] Beide Drucker-IP-Felder teilen denselben Validierungspfad (keine Duplikat-Logik; loses Regex entfernt).
- [x] `make check` grün.

---

## Phase 4: Kommentar in `druckstation_repo` korrigieren (#4)

### Context

- `backend/repository/druckstation_repo/repo.go:42-43` — „Wird vom Relay-Service verwendet"

### What to build

Den irreführenden Kommentar korrigieren: `GetKonfigurierteDruckstationen` wird
von der **Arbeitsbon-Policy** (Table-Command) genutzt, nicht vom Relay (das
Relay kennt laut Handbuch keine Kategorien).

### Acceptance criteria

- [x] Kommentar beschreibt den tatsächlichen Aufrufer (Arbeitsbon-Policy).
- [x] `make check` grün.

---

## Phase 5: Toten Legacy-Zweig in `parseTischName` entfernen (#5)

### Context

- `backend/api/bondruck/application/arbeitsbon_policy.go:105-119` — Legacy-Fallback `tisch:{id}`
- `backend/domain/kasse/subject.go:13-16` — erzeugt nur `kassensitzung-{nr}/tisch-{id}`

### What to build

Den toten `tisch:{id}`-Fallback entfernen; nur das aktuelle Subject-Format
parsen. Ein defensiver Rückgabewert (Subject unverändert) bleibt, falls
`/tisch-` fehlt. Konsistent zur Aktive-Phase-Regel „keine Legacy-Migration".

### Acceptance criteria

- [x] `parseTischName` behandelt nur noch `kassensitzung-{nr}/tisch-{id}`.
- [x] Kein Aufrufer ist auf das Legacy-Format angewiesen (verifiziert).
- [x] Bestehende Arbeitsbon-Tests bleiben grün.
- [x] `make check` grün.

---

## Phase 6: Ungenutzte ESC/POS-Konstanten entfernen (#6)

### Context

- `backend/api/bondruck/application/escpos/constants.go:9,18,25-28` —
  `AlignRight`, `TextDoubleWidth`, `StatusPaper` ungenutzt; irreführender
  „wird im Relay verwendet"-Kommentar
- `cmd/relay/main.go:159-167` — Relay hardcodet die Status-Bytes selbst

### What to build

Die ungenutzten Konstanten entfernen und den irreführenden Kommentar bereinigen.
(Hinweis: Falls Phase 2 eine neue Codepage-Konstante einführt, bleibt diese
erhalten — entfernt wird nur tatsächlich Ungenutztes.)

### Acceptance criteria

- [x] `AlignRight`, `TextDoubleWidth`, `StatusPaper` entfernt (sofern nach Phase 2 weiterhin ungenutzt).
- [x] Kein toter „wird im Relay verwendet"-Kommentar mehr.
- [x] `make check` (inkl. Lint) grün.

---

## Phase 7: Repository-Typ aus der Application-Signatur entfernen (#7)

### Context

- `backend/api/druckstation/application/query.go:11-29` — gibt
  `[]druckstation_repo.Druckstation` zurück
- `backend/api/druckstation/http/handler.go:14-50` — mappt bereits auf DTO

### What to build

`GetAlleDruckstationen` einen Application-lokalen (oder Domain-)Typ zurückgeben
lassen statt das Repository-Struct nach außen zu reichen. Die HTTP-Schicht mappt
weiterhin auf das Response-DTO; das Verhalten bleibt unverändert. Bringt
Druckstation auf dieselbe Schichtentrennung wie die übrigen Module.

### Acceptance criteria

- [x] Die exportierte Query-Signatur referenziert keinen `*_repo`-Typ mehr.
- [x] HTTP-Response unverändert (DTO-Form identisch).
- [x] `make check` grün.

---

## Phase 8: Test für `NewBondruckEinstellungen` ergänzen (#8)

### Context

- `backend/domain/settings/bondruck_einstellungen.go:13-35` — Validierung untested
- Keine Tests in `settings/application`, `settings/http`, `druckstation/application`

### What to build

Unit-Test (`//go:build unit`) für `NewBondruckEinstellungen`/`Validate`: gültige
IPv4 wird akzeptiert, ungültige (inkl. Oktett > 255, IPv6) abgelehnt, leere IP
erlaubt. Optional, falls schlank: Happy-Path-Tests für die ungetesteten
Settings-/Druckstation-Application-Services.

### Acceptance criteria

- [ ] `NewBondruckEinstellungen` durch Unit-Test abgedeckt (gültig/ungültig/leer).
- [ ] Test folgt dem `//go:build unit`-Muster.
- [ ] `make check` grün.

---

## Phase 9: Frontend-Mapping für `kasse_nicht_geoeffnet` (Beobachtung)

### Context

- `frontend/src/service/components/table/TischHistorie.tsx:104-114` —
  `belegDrucken`-`byCode` ohne `kasse_nicht_geoeffnet`
- `backend/api/table/http/command_handler.go:386-401` — Handler kann den 409 liefern

### What to build

Den Fehlercode `kasse_nicht_geoeffnet` in die `byCode`-Map des
`belegDrucken`-`useActionSubmit` aufnehmen, mit deutscher Meldung (z. B.
„Keine Kassensitzung geöffnet."). Bestehende Codes bleiben unverändert.

### Acceptance criteria

- [ ] `belegDrucken` zeigt für `kasse_nicht_geoeffnet` eine spezifische deutsche Meldung.
- [ ] Übrige Fehlercodes unverändert.
- [ ] `make check` grün.
