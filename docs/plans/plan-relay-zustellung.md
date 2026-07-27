# Plan: Zuverlässige Bon-Zustellung im Print-Relay

> Source PRD: n/a (aus Praxis-Fehlerbericht und Code-Analyse vom 2026-07-27)

## Goal

Arbeitsbons dürfen nicht mehr stillschweigend verschwinden. Ein Auftrag gilt erst
dann als `gedruckt`, wenn der Drucker die Daten nachweislich verarbeitet hat; im
Zweifel bleibt er `offen` und wird erneut zugestellt.

**Fehlerbild aus der Praxis:** Eine Bestellung mit 5–6 verschiedenen Getränken
erzeugt bei Bonmodus `pro_position` 5–6 Druckaufträge, von denen nur 1–2 gedruckt
werden. Die übrigen tauchen weder auf Papier noch in der Liste fehlgeschlagener
Aufträge auf.

**Ursachenanalyse (Code-belegt):**

1. `windows/relay/main.go — druckeAuftrag()` öffnet **zwei** TCP-Verbindungen pro
   Bon (`checkPrinter()` und `sendToPrinter()`). `verarbeiteGruppe()` läuft ohne
   Pause, und alle Bons einer Bestellung teilen sich dieselbe Ziel-IP, landen also
   in derselben Gruppe: 6 Bons = 12 Verbindungen in wenigen hundert Millisekunden,
   während der Drucker pro Bon ~1–2 s zum Drucken und Schneiden braucht. Bondrucker
   mit Ethernet-Modul akzeptieren auf Port 9100 typischerweise genau eine
   Verbindung gleichzeitig.
2. `sendToPrinter()` meldet Erfolg, sobald `conn.Write` zurückkehrt — also sobald
   die Bytes im Kernel-Sendepuffer liegen. Danach wird sofort geschlossen. Verwirft
   das Druckermodul die Daten, wird der Auftrag trotzdem auf `gedruckt` gesetzt
   (`backend/sqlc/queries/relay.sql — MarkDruckauftragGedruckt`). Das ist der
   einzige Pfad, auf dem ein Bon spurlos verschwinden kann.
3. `backend/repository/druckauftrag_repo/repo.go — ReportDruckergebnis()` setzt den
   Backoff nur auf den *gescheiterten* Auftrag. Seine Nachfolger bleiben sofort
   fällig und werden vorgezogen: die Bon-Reihenfolge zerfällt, und der blockierte
   Auftrag läuft wiederholt in denselben beschäftigten Drucker, bis er nach
   `MaxDruckversuche` auf `fehlgeschlagen` kippt.

## Architectural decisions

- **Eine TCP-Verbindung pro Ziel-IP und Poll-Zyklus.** Alle Bons einer Gruppe
  werden sequenziell über dieselbe offene Verbindung geschrieben. Damit wirkt
  TCP-Backpressure: ist der Empfangspuffer des Druckers voll, blockiert
  `conn.Write` bis zum Write-Timeout und liefert einen *echten* Fehler, statt dass
  Daten still verworfen werden.
- **Zustellquittung per `GS r 1` (`0x1D 0x72 0x01`).** `GS r` ist ein
  **gepuffertes** Kommando: der Drucker führt es erst aus, nachdem die davor
  empfangenen Druckdaten verarbeitet sind. Eine Antwort darauf ist deshalb ein
  Nachweis, dass alle vorangegangenen Bons konsumiert wurden. Das heute genutzte
  `DLE EOT` (`0x10 0x04 0x04`) ist ein Echtzeit-Kommando und beweist das
  ausdrücklich **nicht**. Quelle:
  [Epson ESC/POS Reference — GS r](https://download4.epson.biz/sec_pubs/pos/reference_en/escpos/gs_lr.html).
- **Im Zweifel doppelt drucken.** Bleibt unklar, ob ein Bon gedruckt wurde, bleibt
  der Auftrag `offen`. Arbeitsbons sind nicht-fiskalisch; ein doppelter Bon kostet
  Papier, ein fehlender kostet ein Getränk. Genau ein Fall ist davon ausgenommen:
  antwortet der Drucker auf `GS r 1` gar nicht (Timeout), gilt die Gruppe als
  zugestellt — sonst wäre jeder Drucker ohne `GS r`-Unterstützung dauerhaft
  unbenutzbar. Ein *Verbindungsabbruch* beim Warten auf die Quittung ist dagegen
  kein Timeout und bestätigt nichts.
- **Backoff pro Ziel-IP statt pro Auftrag.** Scheitert ein Auftrag, wartet die
  gesamte Warteschlange dieses Druckers. Das erhält die Bon-Reihenfolge und
  verhindert, dass der blockierte Auftrag von seinen eigenen Nachfolgern überholt
  wird und als Einziger die Fehlversuche aufbraucht.
- **Keine DB-Migration.** Der Backoff pro Ziel-IP kommt ohne Schema-Änderung aus:
  `druckauftraege.ziel_ip` existiert bereits, `naechster_versuch_ab` ebenfalls. Die
  Freeze-Disziplin bleibt unberührt.
- **Keine Änderung am Relay-API-Vertrag.** `POST /relay/poll` und
  `POST /relay/ergebnis` behalten Request- und Response-Format. Phase 1 (Relay) und
  Phase 2 (Backend) sind dadurch unabhängig voneinander ausrollbar.
- **ESC/POS-Transportkommandos gehören ins Relay.** `windows/relay` ist ein eigenes
  Go-Modul ohne Abhängigkeit zum Backend. Statusabfragen (`DLE EOT`, `GS r`) sind
  Transport, kein Bon-Inhalt, und werden im Relay als lokale Konstanten geführt —
  nicht in `backend/api/druck/bondruck/application/escpos/constants.go`.

## Inventory

**Relay (eigenes Modul `windows/relay`, geprüft mit `make check-relay`):**

- `windows/relay/main.go — verarbeiteZyklus()` — gruppiert nach Ziel-IP, Gruppen
  laufen parallel in Goroutines. Bleibt erhalten.
- `windows/relay/main.go — verarbeiteGruppe()` — Schleife über die Aufträge einer
  IP, Abbruch beim ersten Fehler. Geht in die neue Gruppen-Zustellung auf.
- `windows/relay/main.go — druckeAuftrag()`, `checkPrinter()`, `sendToPrinter()` —
  die drei Funktionen mit je eigener Verbindung. Werden ersetzt.
- `windows/relay/main.go — druckFunc` — Injektionspunkt pro Auftrag; wird zu einem
  Injektionspunkt pro Gruppe.
- `windows/relay/main.go — dialTimeout`, `readTimeout`, `writeTimeout` — bestehende
  Timeout-Konstanten.
- `windows/relay/main_test.go — newFakePrinter()` — bestehendes Test-Double auf
  Auftrags-Ebene (kein echtes TCP). Muss auf Gruppen-Ebene umgestellt werden.

**Backend:**

- `backend/sqlc/queries/relay.sql — IncrementDruckauftragFehlversuch`,
  `SetDruckauftragFaelligkeit`, `GetOffeneDruckauftraege` — die Backoff-Mechanik.
- `backend/repository/druckauftrag_repo/repo.go — ReportDruckergebnis()`,
  `backoffDauer()`, `MaxDruckversuche` — Fehlversuchszählung und Backoff-Staffel.
- `backend/repository/druckauftrag_repo/repo_test.go` — Integrationstests gegen
  echtes Postgres (Build-Tag `integration`).
- `backend/api/druck/relay/http/handler.go — ErgebnisHandler()` — bleibt unverändert.

**Dokumentation:**

- `README.md — §Print-Relay` — beschreibt heute „pro Zyklus genau einen kurzen
  Zustellversuch (TCP-Timeout 2 s)" und die Backoff-Staffel.
- `docs/handbuch.md — §4.6 Bondruck`, Absatz „Relay = Transport".

## Resolved decisions

- **Zustellbestätigung:** eine Verbindung pro Ziel-IP plus `GS r 1`-Quittung nach
  dem letzten Bon der Gruppe. Keine Quittung pro einzelnem Bon — der Gewinn an
  Auflösung rechtfertigt die Wartezeit je Bon nicht.
- **Zweifelsfall:** Auftrag bleibt `offen`, Doppeldruck wird in Kauf genommen.
- **Sichtbarkeit fehlgeschlagener Aufträge:** keine zusätzliche Anzeige. Die
  bestehende Admin-Sicht (Dashboard-Kachel plus Druckstationen-Seite mit Retry und
  Verwerfen) genügt. Kein Banner in der Service-Ansicht.
- **Papierstatus-Prüfung bleibt erhalten**, wandert aber auf die gemeinsame
  Verbindung und kostet damit keine zusätzliche Verbindung mehr.

## Open questions / Risks

- **`GS r`-Unterstützung der eingesetzten Hardware ist unbekannt.** Der
  ESC/POS-Formatter nennt als Referenzgerät den MUNBYN ITPP047P, ob das Gerät des
  Vereins `GS r` beantwortet, ist nicht verifiziert. Abgesichert durch den
  Timeout-Fallback: ein Drucker ohne `GS r` verhält sich wie heute, nur ohne den
  Verbindungssturm. Der Logeintrag pro Gruppe (Phase 1) macht nach dem nächsten
  Einsatz sichtbar, welcher Fall vorliegt.
- **Das Relay muss beim Verein neu ausgerollt werden.** Ein reines Server-Update
  behebt nichts — die Zustelllogik steckt in `jotti-relay.exe`
  (`make build-relay-windows VERSION=…`, Release-ZIP über `make release-windows`).
- **Doppeldrucke nehmen zu**, insbesondere bei instabilem WLAN zwischen Relay und
  Server. Bewusst akzeptiert.
- **Die Ursache ist erschlossen, nicht gemessen** — es lagen keine Relay-Logs und
  kein DB-Zugriff vor. Deshalb ist die Diagnosefähigkeit (Phase 1, Logging) Teil
  des Plans und nicht optional.

---

## Phase 1: Gruppen-Zustellung über eine Verbindung mit Quittung

### Context

- `windows/relay/main.go — verarbeiteZyklus()` — behält Gruppierung nach Ziel-IP
  und die parallele Verarbeitung je IP; nur der übergebene Callback wechselt von
  „ein Auftrag" auf „eine Gruppe".
- `windows/relay/main.go — verarbeiteGruppe()`, `druckeAuftrag()`,
  `checkPrinter()`, `sendToPrinter()` — die zu ersetzende Zustellkette.
- `windows/relay/main.go — zyklusErgebnis`, `fehlversuch` — Ergebnistypen; bleiben
  unverändert, damit `meldeErgebnis()` und der Backend-Vertrag gleich bleiben.
- `windows/relay/main_test.go — newFakePrinter()` — bisheriges Test-Double.

### What to build

Die Zustelleinheit wechselt vom einzelnen Auftrag zur Gruppe aller Aufträge einer
Ziel-IP. Eine neue Funktion `zustelleGruppe` übernimmt eine Gruppe und gibt die
bestätigt zugestellten Auftrags-IDs sowie höchstens einen Fehlversuch zurück:

1. **Einmal verbinden** (`dialTimeout`). Scheitert das, ist der erste Auftrag der
   Gruppe der Fehlversuch, nichts ist bestätigt.
2. **Papierstatus prüfen** auf derselben Verbindung: `DLE EOT 4` schreiben, eine
   Antwort mit kurzem Timeout lesen. Verhalten wie bisher — keine Antwort gilt als
   in Ordnung, gesetzte End-Sensor-Bits (`0x60`) sind ein Fehler, Near-End-Bits
   (`0x0C`) nur eine Warnung im Log.
3. **Bons sequenziell schreiben**, in ID-Reihenfolge, jeweils mit `writeTimeout`.
   Ein Base64-Dekodierfehler oder ein Schreibfehler bricht die Gruppe ab und macht
   genau diesen Auftrag zum Fehlversuch.
4. **Quittung anfordern:** nach dem letzten geschriebenen Bon `GS r 1` senden und
   eine Antwort lesen. Das Lese-Timeout skaliert mit der Gruppengröße, weil der
   Drucker erst antwortet, wenn er alle Bons verarbeitet hat.
5. **Ergebnis der Quittung auswerten:**
   - *Antwort erhalten* → alle geschriebenen Bons sind bestätigt.
   - *Timeout ohne Antwort* → Drucker unterstützt `GS r` nicht; alle geschriebenen
     Bons gelten als zugestellt (Fallback, nie schlechter als der Ist-Zustand).
   - *Verbindungsabbruch / EOF / Schreibfehler beim Quittungskommando* → nichts
     wird bestätigt; der erste Auftrag der Gruppe ist der Fehlversuch, der Rest
     bleibt offen.
6. **Bricht die Gruppe vor der Quittung ab** (Schritt 1–3), gilt nichts als
   zugestellt — auch nicht die Bons, die vorher erfolgreich geschrieben wurden.
   Sie werden im nächsten Zyklus erneut zugestellt (bewusster Doppeldruck).

Zusätzlich pro Gruppe eine Logzeile mit Ziel-IP, Anzahl Bons, Ausgang der Quittung
(bestätigt / unbeantwortet / abgebrochen) und Dauer, damit der nächste Vorfall
diagnostizierbar ist, ohne DB-Zugriff.

`checkPrinter`, `sendToPrinter`, `druckeAuftrag`, `verarbeiteGruppe` und
`druckFunc` entfallen ersatzlos — ihre Aufgaben liegen vollständig in
`zustelleGruppe`. Der Injektionspunkt für Tests wandert auf die Gruppen-Ebene.

### Acceptance criteria

- [ ] Eine Gruppe von 6 Aufträgen an dieselbe Ziel-IP öffnet **genau eine**
      TCP-Verbindung (heute: 12). Nachgewiesen mit einem echten TCP-Listener im
      Test, der Verbindungen zählt.
- [ ] Regressionstest gegen das gemeldete Fehlerbild: ein Test-Listener, der
      **nur eine** Verbindung gleichzeitig annimmt und weitere sofort ablehnt,
      empfängt alle 6 Bon-Payloads vollständig und in ID-Reihenfolge.
- [ ] Antwortet der Test-Listener auf `GS r 1`, werden alle Aufträge der Gruppe als
      `gedruckte IDs` gemeldet, kein Fehlversuch.
- [ ] Antwortet der Test-Listener **nicht** auf `GS r 1` (Timeout), werden alle
      geschriebenen Aufträge trotzdem als gedruckt gemeldet.
- [ ] Schließt der Test-Listener die Verbindung, bevor die Quittung kommt, wird
      **kein** Auftrag der Gruppe als gedruckt gemeldet, und genau ein Auftrag
      erscheint als Fehlversuch.
- [ ] Ein Schreibfehler beim dritten Bon macht Auftrag 3 zum Fehlversuch; die
      Aufträge 1, 2, 4, 5, 6 werden **nicht** als gedruckt gemeldet und bleiben
      damit offen.
- [ ] Meldet der Drucker Papier leer (End-Sensor-Bits gesetzt), ist der erste
      Auftrag der Gruppe der Fehlversuch und kein Bon wird gesendet.
- [ ] Ein nicht erreichbarer Drucker blockiert andere Ziel-IPs nicht — die
      bestehende Zusicherung aus `TestVerarbeiteZyklusSkipNachErstfehler` gilt
      unverändert auf Gruppen-Ebene weiter.
- [ ] Pro Gruppe erscheint eine Logzeile mit Ziel-IP, Anzahl Bons, Quittungs-Ausgang
      und Dauer.
- [ ] `make check-relay` läuft grün (inkl. `-race`, `golangci-lint`, `go vet`).
- [ ] `README.md — §Print-Relay` beschreibt die neue Zustellung: eine Verbindung je
      Drucker und Zyklus, Quittung per gepuffertem Statuskommando, im Zweifel
      erneute Zustellung.
- [ ] `docs/handbuch.md — §4.6`, Absatz „Relay = Transport", beschreibt die
      Gruppen-Zustellung und die Quittungssemantik.

---

## Phase 2: Backoff pro Ziel-IP statt pro Auftrag

### Context

- `backend/sqlc/queries/relay.sql — IncrementDruckauftragFehlversuch` — liefert
  heute `versuche, status` zurück; braucht zusätzlich `ziel_ip`, damit das
  Repository weiß, welche Warteschlange zu bremsen ist.
- `backend/sqlc/queries/relay.sql — SetDruckauftragFaelligkeit` — setzt
  `naechster_versuch_ab` für **eine** ID; wird durch die Ziel-IP-Variante ersetzt
  und entfällt.
- `backend/sqlc/queries/relay.sql — GetOffeneDruckauftraege` — filtert bereits auf
  `naechster_versuch_ab` und sortiert nach `id ASC`; bleibt unverändert und liefert
  dadurch automatisch die korrekte Reihenfolge.
- `backend/repository/druckauftrag_repo/repo.go — ReportDruckergebnis()` — die
  Transaktion, die Erfolge quittiert und Fehlversuche hochzählt.
- `backend/repository/druckauftrag_repo/repo.go — backoffDauer()`,
  `MaxDruckversuche` — Staffel und Obergrenze bleiben unverändert (5/15/30/60/180 s,
  Aufgabe nach 6 Fehlversuchen).

### What to build

Ein gemeldeter Fehlversuch bremst künftig die gesamte Warteschlange seines
Druckers, nicht nur den einen Auftrag.

`IncrementDruckauftragFehlversuch` gibt zusätzlich die `ziel_ip` des betroffenen
Auftrags zurück. Solange der Auftrag danach noch `offen` ist, setzt
`ReportDruckergebnis` die Fälligkeit über eine neue Query
`SetDruckauftragFaelligkeitFuerZielIP` auf **alle** offenen Aufträge dieser
Ziel-IP — die Wartezeit ergibt sich unverändert aus `backoffDauer(versuche)` des
gescheiterten Auftrags. Die Einzel-ID-Variante `SetDruckauftragFaelligkeit`
entfällt.

Kippt der Auftrag mit diesem Fehlversuch auf `fehlgeschlagen`, bekommt die
Warteschlange keinen Backoff: der Auftrag ist aus dem Rennen, der nächste darf
sofort versuchen.

Die Signatur von `ReportDruckergebnis` und der HTTP-Vertrag bleiben unverändert —
das Relay meldet weiterhin nur `id` und `fehler` je Fehlversuch, die Ziel-IP
ermittelt das Backend selbst.

Eine DB-Migration ist nicht nötig; `ziel_ip` und `naechster_versuch_ab` existieren
bereits. `make sqlc` muss nach der Query-Änderung laufen,
`backend/sqlc/dbgen/` wird nicht von Hand editiert.

### Acceptance criteria

- [ ] Integrationstest: fünf offene Aufträge an dieselbe Ziel-IP, ein gemeldeter
      Fehlversuch auf den ersten → **alle fünf** haben ein `naechster_versuch_ab`
      in der Zukunft, und `GetOffeneDruckauftraege` liefert unmittelbar danach
      keinen davon.
- [ ] Integrationstest: Aufträge an eine **andere** Ziel-IP bleiben von diesem
      Fehlversuch unberührt und sind weiterhin sofort fällig.
- [ ] Integrationstest: nach Ablauf der Wartezeit liefert
      `GetOffeneDruckauftraege` die Aufträge wieder in aufsteigender ID-Reihenfolge
      — der zuvor gescheiterte Auftrag zuerst.
- [ ] Integrationstest: kippt der Auftrag mit dem sechsten Fehlversuch auf
      `fehlgeschlagen`, bekommen die übrigen offenen Aufträge derselben Ziel-IP
      **keinen** neuen Backoff und sind sofort fällig.
- [ ] Ein Fehlversuch auf eine ID, die nicht (mehr) `offen` ist, bleibt ein
      No-Op ohne Backoff-Wirkung auf die Warteschlange — die bestehende Zusicherung
      aus `TestReportDruckergebnis_StaleFehlversuchIstNoOp` gilt weiter.
- [ ] `SetDruckauftragFaelligkeit` (Einzel-ID) existiert nicht mehr, und
      `grep -rn "SetDruckauftragFaelligkeit\b"` findet keine verwaisten Referenzen.
- [ ] `make sqlc` ausgeführt, `backend/sqlc/dbgen/` nur generiert, nicht editiert.
- [ ] Keine neue Datei unter `database/migrations/`.
- [ ] `make verify` läuft grün.
- [ ] `README.md — §Print-Relay` und `docs/handbuch.md — §4.6` beschreiben den
      Backoff als Wartezeit der gesamten Drucker-Warteschlange, nicht des einzelnen
      Auftrags.

---

## Bewusst nicht im Scope

- **Bonmodus „pro Stück"** (ein Abreiß-Bon je Getränk statt je Katalogzeile). Der
  Verein bestellt verschiedene Getränke, damit ist das nicht die Ursache dieses
  Vorfalls. Falls die Theke pro Getränk einen Bon will, ist das eine eigene
  Produktentscheidung.
- **`Beep` vor `Init`** in `backend/api/druck/bondruck/application/escpos/formatter.go
  — FormatPositionBon()`: der Signalton wird vor `ESC @` geschrieben, das den
  Drucker zurücksetzt. Betrifft nur die Essen-Station und ist bisher nicht als
  Problem gemeldet.
- **Nichtdeterministische Auftragsreihenfolge zwischen Kategorien** durch die
  Map-Iteration in `backend/api/druck/bondruck/application/arbeitsbon_policy.go —
  createStationsAuftraegeFromData()`. Kosmetisch, verschiedene Kategorien gehen an
  verschiedene Drucker.
- **Fragile Event-Deserialisierung** in `arbeitsbon_policy.go —
  unmarshalPositionenMitKommentar()`: das Event-JSON wird in `kasse.Position`
  (ohne json-Tags) gelesen und trifft nur über Gos case-insensitives
  Feld-Matching. Schlägt es fehl, entstehen null Druckaufträge — ohne Log und ohne
  Fehler für die Servicekraft. Heute durch einen Byte-Identisch-Test abgesichert,
  aber ein eigenständiges Härtungsthema.
