# Plan: Behebung der Findings aus dem Architektur-Review, Runde 2 (2026-07)

> Quelle: Architektur- und Compliance-Review der Gesamtimplementierung
> (Runde 2, nach Umsetzung von `plan-review-fixes.md`). Scope und
> Entscheidungen je Finding wurden im Review-Gespräch festgelegt.

## Goal

Die bestätigten Findings der zweiten Review-Runde werden behoben: ein
Geld-Bug beim Direktverkauf-Storno (Übererstattung durch doppelte
Positionsreferenzen), der im DSFinV-K-Export fehlende Tagesabschluss
(Verprobungslücke je Kassensitzung), das Race im Tagesabschluss (Saldo-
und Bestandsprüfung außerhalb jeder Barriere), fehlende Transparenz beim
Eröffnen ohne konfigurierte TSE sowie eine Reihe kleiner Export-,
Robustheits- und Frontend-Fixes. Parallel wird die Dokumentation an die
tatsächliche Implementierung angeglichen (TSE-Key-Speicherung,
Offline-Verhalten, DSFinV-K-Details), und die Offline-Anforderung Q-05
entfällt ersatzlos.

## Inventory (Findings → Code)

| #  | Finding                                                                   | Ort                                                                                              | Schwere |
| -- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ------- |
| F1 | Duplikat-Positionsrefs beim Storno passieren die Validierung → Übererstattung | `backend/api/direktverkauf/application/command.go:199-218`, `backend/api/table/application/command.go:240-259` | hoch |
| F5 | `tagesabschluss-erstellt:v1` fehlt in `belegeFromEvents` → keine `transactions_tse.csv`-Zeile trotz Signatur | `backend/domain/dsfinvk/mapper.go:197` (Switch ohne Case), Event: `backend/domain/kasse/kassensitzung_events.go:82-92` | hoch |
| F7 | Saldo-Sperre und Soll-Bestand werden vor dem Abschluss ohne Barriere geprüft; parallele Buchungen verfälschen den Z-Bon | `backend/api/kasse/application/command.go:260-367`, Status-Guard: `backend/repository/kassenjournal_repo/repo.go:283-292` | hoch |
| F6 | Eröffnen ohne konfigurierte TSE läuft ohne Hinweis durch                  | `frontend/src/admin/kasse/KassensitzungPage.tsx:100` (EroeffnenSection), TSE-Status: `backend/api/settings/http/query_handler.go:129-149` | mittel |
| F3 | Doku behauptet Verschlüsselung der TSE-Keys, Code speichert Klartext      | `docs/compliance.md:434`, `docs/leitfaden/tse-einrichten.md`                                        | Doku |
| F4 | Doku fordert Offline-UI-Blockade, die es nicht gibt; Q-05 widerspricht der Compliance-Einordnung | `docs/compliance.md:29`, `docs/anforderungen.md:22`                                          | Doku |
| B1 | Nachsignierte Vorgänge exportieren leere `TSE_TA_VORGANGSART`             | `backend/repository/kassenjournal_repo/repo.go:572-593`, `backend/domain/dsfinvk/mapper.go:1018`    | mittel |
| B2 | `TSE_VORGANGSDATEN` enthält den QR-Code-String statt der processData      | `backend/domain/dsfinvk/mapper.go:1020`                                                             | mittel |
| B3 | Zertifikat > 2000 Zeichen wird still weggelassen (Entscheidung ist dokumentiert, Warnung fehlt) | `backend/domain/dsfinvk/mapper.go:1253-1273`                                            | niedrig |
| B4 | `Z_BUCHUNGSTAG` immer gefüllt, mit UTC-Datum (kippt nach Mitternacht auf den Vortag) | `backend/domain/dsfinvk/mapper.go:548,557-561`                                              | niedrig |
| B5 | Deadlock (40P01) wird als 500 statt 409 gemappt                           | `backend/db/db.go:18` (nur 23505 gemappt)                                                           | niedrig |
| B6 | `parseCents` parst float-basiert; > 2 Nachkommastellen runden inkonsistent | `frontend/src/lib/utils.ts:40-43`, `frontend/src/components/common/EuroInput.tsx:22`               | niedrig |
| B7 | Deaktivierte Benutzer behalten ausgestellte JWTs bis 12 h                 | `backend/api/middleware/middleware.go` (kein Status-Check)                                          | mittel |
| B8 | Trinkgeld wird angezeigt, aber nirgends gebucht; kein Hinweis für die Servicekraft | `frontend/src/service/components/table/drawerUtils.ts:33-61`, `ZahlungDrawer.tsx:117-145`   | niedrig |
| B9 | Reporting rechnet USt auf Aggregatbasis, Beleg/TSE/Export auf Zeilenbasis → Cent-Differenzen | `backend/api/reporting/application/query.go:59-111`, SQL `GetUmsatzProSteuersatz`         | mittel |
| C  | Doku-Drift: `BON_STORNO`, Dezimaltrennzeichen, Z-Nr-Lückenlosigkeit, Einmalpasswort-Länge | `docs/compliance.md:291,352-353`, `docs/handbuch.md:253,333`, `database/migrations/01_initial.up.sql:122` (Kommentar) | Doku |

Nicht in diesem Plan (bewusst): F2 (Default-Einmalpasswort des initialen
Admins, Entscheidung: vorerst ignorieren), verwaiste TSE-Transaktionen im
allgemeinen Sign-then-Persist-Pfad (nur die Duplikat-Ursache wird über F1
entschärft), Trinkgeld als eigener Geschäftsvorfall, Least-Privilege-DB-Role,
serverseitiger OCC-Retry, Idempotency-Keys.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Duplikate ablehnen statt kumulieren (F1).** Storno-Requests mit
  mehrfach vorkommender `PositionID` sind ungültig (HTTP 400,
  `ErrPositionNichtStornierbar`). Kein legitimer Client sendet Duplikate;
  Ablehnung ist die kleinste Änderung und hält die Validierung vor dem
  TSE-Call.
- **Zweiphasiger Tagesabschluss über Zwischenstatus (F7).** Die
  `kassensitzungen.status`-Zustandsmenge wird um `wird_abgeschlossen`
  erweitert. `KasseAbschliessen` setzt diesen Status als ersten Schritt;
  der bestehende Status-Guard im Event-Write (FOR SHARE + Prüfung auf
  `offen`) wird zur Barriere für alle Buchungs-Events. Nur die drei
  Abschluss-Events (Kassensturz, Differenz, Tagesabschluss) sind im
  Zwischenstatus zugelassen. Keine umschließende Transaktion, keine
  DB-Locks über TSE-Netzwerk-Calls hinweg; die Wiederholbarkeit bei
  Teilfehlern bleibt erhalten.
- **Keine TSE-Sperre (F6).** jotti funktioniert ohne konfigurierte TSE
  vollständig. Die Transparenz kommt aus einem Bestätigungsdialog beim
  Eröffnen (Frontend) und einem Log-Warning (Backend), nicht aus einer
  Sperre.
- **Doku folgt Code, nicht umgekehrt (F3, F4, C).** Wo die Dokumentation
  mehr verspricht als implementiert ist (Verschlüsselung, Offline-Blockade,
  BON_STORNO, Dezimalpunkt, lückenlose Z-Nr), wird die Doku auf die
  tatsächliche, rechtlich tragfähige Implementierung korrigiert. Für die
  Offline-Frage ist die maßgebliche Eigenschaft: Kein Vorgang kann ohne
  synchronen Backend-Request erfasst oder lokal zwischengespeichert werden
  (kein Service Worker, keine optimistischen Updates, kein Offline-Speicher).
- **USt-Rundung einheitlich auf Positionszeilenbasis (B9).** Referenz ist
  die bestehende Logik von Beleg, TSE-processData und DSFinV-K-Export
  (`steuer.Aufteilen` je Positionszeile, dann aggregieren). Das Reporting
  wird auf dieselbe Basis umgestellt.
- **Aktive Entwicklungsphase (AGENTS.md).** Schema-Änderungen direkt in
  `database/migrations/01_initial.up.sql`, keine neuen Migrationsdateien.
  Nach Query-Änderungen `make sqlc`.

## Resolved decisions

- F2 wird nicht behoben (Entscheidung im Review-Gespräch).
- F3: keine Verschlüsselung implementieren, nur Doku anpassen.
- F4: Q-05 ersatzlos aus der Roadmap streichen; compliance.md §2.2 von
  „muss sofort blockieren" auf die tatsächliche Architektur-Eigenschaft
  umformulieren; die Eigenschaft wird im Code verifiziert und das Ergebnis
  in der Doku festgehalten.
- F6: Dialog sinngemäß „Keine TSE konfiguriert. Trotzdem eröffnen?",
  keine Backend-Sperre.
- B2: `TSE_VORGANGSDATEN` bleibt leer (amtlich optional); die processData
  wird nicht rekonstruiert.
- B3: Verhalten bleibt (leer statt abgeschnitten), es kommt nur ein
  Log-Warning hinzu.

## Open questions / Risks

- **Wiederanlauf im Zwischenstatus (F7):** Bricht der Abschluss nach dem
  Statuswechsel ab, steht die Sitzung auf `wird_abgeschlossen`. Der erneute
  `KasseAbschliessen`-Aufruf muss diesen Status akzeptieren und fortsetzen;
  zusätzlich setzt der Fehlerpfad den Status best effort auf `offen`
  zurück. Beides testen, sonst kann sich die Kasse aussperren.
- **B7 kostet einen PK-Lookup pro Request.** Bei der Zielgröße (1 bis 30
  Nutzer) unkritisch; trotzdem im Blick behalten.
- **B9 verändert ausgewiesene Beträge.** Dashboard-Werte können um Cents
  von bisherigen Ständen abweichen; das ist gewollt (Konsistenz mit dem
  Export), sollte aber im Commit-Text erwähnt werden.
- **Golden-Files (F5, B1, B2, B4):** Die Export-Golden-Tests ändern sich in
  mehreren Phasen. Reihenfolge einhalten (Phase 2 vor Phase 3), damit die
  Fixtures nur zweimal angefasst werden.

---

## Phase 1: Storno lehnt Duplikat-Positionsrefs ab (F1)

### Context

- `backend/api/direktverkauf/application/command.go:199-218` —
  `validatePositionRefs` prüft jede Referenz einzeln gegen die verfügbare
  Menge; zwei Referenzen auf dieselbe Position summieren sich unbemerkt.
- `backend/api/direktverkauf/application/command.go:223-245` —
  `resolvePositionen` addiert die Duplikate ungeprüft in den Storno-Betrag.
- `backend/api/table/application/command.go:240-259` — gleichnamige
  Funktion für Tisch-Pfade; Aufrufer bei Zeile 589 (Storno), 685
  (Korrektur), 832 (Ausgabe). Dort verhindert die Projektion
  (`reduceByPositionStrict`) den Schaden, aber erst nach der
  TSE-Signierung, mit 500 statt 400 als Ergebnis.

### What to build

Beide `validatePositionRefs`-Varianten lehnen Requests ab, in denen eine
`PositionID` mehrfach vorkommt. Der Direktverkauf-Storno mit Duplikaten
liefert damit HTTP 400 statt einer Übererstattung; die Tisch-Pfade liefern
400 vor dem TSE-Call statt 500 danach (keine verwaiste TSE-Transaktion
für diesen Fall).

### Acceptance criteria

- [x] Direktverkauf-Storno mit `[{id, menge: 2}, {id, menge: 2}]` bei
      verfügbarer Menge 3 wird mit 400 abgelehnt; kein Event, keine
      TSE-Transaktion.
- [x] Auch Duplikate, deren Summe die verfügbare Menge nicht übersteigt,
      werden abgelehnt (Duplikat ist per se ungültig).
- [x] Gleiches Verhalten für Tisch-Storno, Korrektur und
      Ausgabe-Bestätigung.
- [x] Bestehende Storno-Tests bleiben grün.

---

## Phase 2: Tagesabschluss im DSFinV-K-Export (F5)

### Context

- `backend/domain/dsfinvk/mapper.go:197` — `belegeFromEvents` hat keinen
  Case für `tagesabschluss-erstellt:v1`; die Signatur des Events erscheint
  nirgends im Export, obwohl `mapper.go:516-522` das Event bereits für
  `Z_ERSTELLUNG` liest.
- `backend/domain/kasse/kassensitzung_events.go:82-92` — Event-Payload
  inkl. `TSETxID`/`TSEData`.
- `backend/api/kasse/application/command.go:329-355` — der Tagesabschluss
  wird als `SonstigerVorgang` signiert.

### What to build

Der Tagesabschluss wird als eigener Bon exportiert: `BON_TYP =
AVSonstige`, ohne Positionen, geldneutral (0-Beträge), Bediener und
Zeitpunkt aus dem Event, mit `transactions_tse.csv`-Zeile aus den
TSE-Daten des Events (bei Ausfall die `TSE_TA_FEHLER`-Zeile, gleiche Logik
wie bei den übrigen Belegen). Der Abgleich fiskaly-TSE-Export gegen
`transactions_tse.csv` geht damit je Sitzung auf.

### Acceptance criteria

- [x] Export einer abgeschlossenen Sitzung enthält genau einen
      Tagesabschluss-Bon mit `transactions_tse.csv`-Zeile (Golden-Test,
      signierter Fall und Ausfall-Fall).
- [x] Die Aggregationsdateien (businesscases, payment, cashpointclosing,
      cash_per_currency) bleiben durch den 0-Bon unverändert. (Ausnahme laut
      Review-Entscheidung: `Z_ENDE_ID` zeigt nun auf den Abschluss-Bon als
      letzte BON_ID im Abschluss; die aggregierten Beträge bleiben unverändert.)
- [x] `BON_NR`-Vergabe der übrigen Belege verschiebt sich nicht (der
      Tagesabschluss ist das letzte Event der Sitzung).
- [x] Bestehende Invariante „jeder fiskalische Bonkopf hat eine TSE-Zeile"
      deckt den neuen Bon-Typ ab.

---

## Phase 3: DSFinV-K-Detailfixes (B1, B2, B3, B4)

### Context

- `backend/repository/kassenjournal_repo/repo.go:572-593` —
  `GetTSESignaturByTxID` baut `kasse.TSEData` ohne `ProcessType`
  (die Tabelle `tse_signaturen` führt keinen); in
  `backend/domain/dsfinvk/mapper.go:1018` landet dann ein Leerstring in
  `TSE_TA_VORGANGSART`.
- `backend/domain/dsfinvk/mapper.go:1020` — `TSE_VORGANGSDATEN` wird mit
  `b.tse.QRCodeData` befüllt (falsches Format für dieses Feld).
- `backend/domain/dsfinvk/mapper.go:1253-1273` — `certChunk` lässt
  Zertifikate über 2000 Zeichen bewusst leer; es gibt keine Warnung.
- `backend/domain/dsfinvk/mapper.go:548,557-561` — `Z_BUCHUNGSTAG` wird
  immer mit dem UTC-Datum von `Z_ERSTELLUNG` befüllt; amtlich ist das Feld
  nur für einen abweichenden Buchungstag vorgesehen.

### What to build

Vier gekapselte Export-Korrekturen: `TSE_TA_VORGANGSART` wird beim Export
aus dem Event-Typ abgeleitet (das Mapping Event-Typ → processType ist im
Signierpfad deterministisch festgelegt) und gilt damit auch für
nachsignierte Vorgänge. `TSE_VORGANGSDATEN` bleibt leer. Beim Weglassen
eines überlangen Zertifikats wird eine Log-Warnung ausgegeben.
`Z_BUCHUNGSTAG` bleibt leer.

### Acceptance criteria

- [x] Nachsignierter Vorgang exportiert `TSE_TA_VORGANGSART` mit dem
      korrekten processType (`Kassenbeleg-V1` bzw. `Bestellung-V1`).
- [x] `TSE_VORGANGSDATEN` ist in allen Zeilen leer; kein QR-String mehr im
      Export.
- [x] Zertifikat > 2000 Zeichen erzeugt eine Log-Warnung; Felder bleiben
      wie bisher leer.
- [x] `Z_BUCHUNGSTAG` ist leer; Golden-Files entsprechend aktualisiert.

---

## Phase 4: Tagesabschluss-Barriere über Zwischenstatus (F7)

### Context

- `backend/api/kasse/application/command.go:260-367` — Saldo-Sperre
  (Zeile 288-299) und Soll-Bestand (Zeile 280) werden ohne Barriere vor
  den Abschluss-Events geprüft; dazwischen liegen TSE-Roundtrips. Der
  OCC-Anker (Zeile 274) deckt nur den Kassensitzungs-Stream, nicht die
  Tisch-Streams und den Direktverkauf.
- `backend/repository/kassenjournal_repo/repo.go:283-292` — bestehender
  Status-Guard (`FOR SHARE`, `ErrKassensitzungNichtOffen`) in jeder
  Event-Transaktion; die Barriere muss nur den Status umfassen.
- `database/migrations/01_initial.up.sql:126` — CHECK-Constraint der
  Status-Werte (`offen`, `abgeschlossen`).

### What to build

Der Abschluss wird zweiphasig: `KasseAbschliessen` setzt die Sitzung als
ersten Schritt auf `wird_abgeschlossen`. Ab diesem Commit lehnt der
Status-Guard alle Buchungs-Events ab; erst danach laufen Saldo-Prüfung,
Soll-Berechnung, Reporting und TSE-Signierungen. Die drei Abschluss-Events
sind im Zwischenstatus zugelassen, alle anderen Event-Typen nicht. Bei
einem Fehler nach dem Statuswechsel wird der Status auf `offen`
zurückgesetzt; unabhängig davon setzt ein erneuter Abschluss-Aufruf im
Zwischenstatus fort. Das Frontend zeigt für abgelehnte Buchungen während
des Abschlusses eine verständliche Fehlermeldung.

### Acceptance criteria

- [x] Eine parallele Bestellung/Zahlung/Direktverkauf nach dem
      Statuswechsel wird mit 409 abgelehnt; der Z-Bon enthält
      ausschließlich Umsätze, die vor der Barriere committet waren.
- [x] Kassensturz-Differenz und Reporting-Summen desselben Abschlusses
      basieren auf demselben Datenstand (Test mit injizierter paralleler
      Zahlung).
- [x] Abbruch nach Statuswechsel: Sitzung steht wieder auf `offen` oder
      der Wiederholungs-Aufruf schließt sie erfolgreich ab; kein
      dauerhafter Sperrzustand (Test für beide Pfade).
- [x] Buchung in eine Sitzung im Zwischenstatus liefert im Frontend die
      Meldung, dass die Kasse gerade abgeschlossen wird.

---

## Phase 5: Bestätigungsdialog beim Eröffnen ohne TSE (F6)

### Context

- `frontend/src/admin/kasse/KassensitzungPage.tsx:100-190` —
  `EroeffnenSection` mit dem bestehenden Eröffnen-Flow (inkl. Vorbild für
  Vorbedingungs-Meldungen: Betreiber-Stammdaten).
- `backend/api/settings/http/query_handler.go:129-149` —
  TSE-Konfigurations-Query liefert `apiKeyGesetzt`;
  `frontend/src/admin/settings/hooks.ts` hat die zugehörigen Hooks.
- `backend/api/kasse/application/command.go` — `KassensitzungEroeffnen`
  für das Log-Warning.

### What to build

Vor dem Eröffnen prüft das Frontend den TSE-Status. Ist keine TSE
konfiguriert, erscheint ein Bestätigungsdialog (sinngemäß: keine TSE
konfiguriert, Vorgänge dieser Kassensitzung werden nicht nach § 146a AO
signiert, trotzdem eröffnen?) mit Abbrechen und Trotzdem-eröffnen. Das
Backend loggt beim Eröffnen ohne TSE-Konfiguration ein Warning. Keine
Sperre, kein sonstiger Verhaltensunterschied.

### Acceptance criteria

- [x] Ohne TSE-Konfiguration: Dialog erscheint, Abbrechen eröffnet nicht,
      Bestätigen eröffnet die Sitzung.
- [x] Mit konfigurierter TSE: kein Dialog, Flow unverändert.
- [x] Backend-Log enthält beim Eröffnen ohne TSE ein Warning mit z_nr.
- [x] Test in `KassensitzungPage.test.tsx` deckt beide Zweige ab.

---

## Phase 6: Kleine Härtungen (B5, B6, B7, B8)

### Context

- `backend/db/db.go:18` — nur `23505` wird auf einen typisierten Fehler
  gemappt; ein Deadlock (`40P01`, realistisch beim Abschluss parallel zu
  Buchungen) wird als 500 durchgereicht.
- `frontend/src/lib/utils.ts:40-43` — `parseCents` nutzt
  `parseFloat`; Eingaben mit mehr als zwei Nachkommastellen runden
  float-basiert und inkonsistent.
- `backend/api/middleware/middleware.go` — die JWT-Middleware prüft
  Signatur, Expiry und Rolle, aber nicht den aktuellen Benutzerstatus;
  deaktivierte Benutzer behalten Zugriff bis zum Token-Ablauf.
- `frontend/src/service/components/table/drawerUtils.ts:33-61` und
  `ZahlungDrawer.tsx:117-145` — Trinkgeld wird berechnet und angezeigt,
  aber nicht gebucht; ohne Hinweis landet es in der Kassenlade und
  erscheint beim Kassensturz als Differenz.

### What to build

Vier unabhängige Fixes: Deadlocks werden auf den bestehenden
Konflikt-Fehler (409) gemappt. `parseCents` parst string-basiert und
begrenzt auf zwei Nachkommastellen (ungültige bzw. übergenaue Eingaben
werden abgelehnt statt still gerundet). Die JWT-Middleware prüft
zusätzlich per PK-Lookup, dass der Benutzer `active` ist. Der
Zahlungs-Drawer bekommt einen kurzen Hinweis, dass das angezeigte
Trinkgeld nicht als Kasseneinnahme gebucht wird und nicht in die
Kassenlade gehört; der Leitfaden erwähnt das ebenfalls.

### Acceptance criteria

- [x] Simulierter 40P01-Fehler wird als Konflikt (409) beantwortet.
- [x] `parseCents`: `"12,50"` → 1250, `"12.50"` → 1250, `"12,505"` und
      `"1,2,3"` ungültig; Tests in `utils.test.ts` und `EuroInput.test.tsx`.
- [x] Deaktivierter Benutzer erhält mit gültigem Alt-Token sofort 401.
- [x] Trinkgeld-Hinweis erscheint nur, wenn Trinkgeld angezeigt wird;
      Leitfaden-Abschnitt ergänzt.

---

## Phase 7: Reporting-USt auf Zeilenbasis (B9)

### Context

- `backend/api/reporting/application/query.go:59-111` —
  `berechneUmsatzProSteuersatz` aggregiert erst Brutto je Steuersatz (SQL
  `GetUmsatzProSteuersatz`) und ruft dann `steuer.Aufteilen` auf dem
  Aggregat auf; Beleg, TSE-processData und DSFinV-K-Export teilen dagegen
  je Positionszeile auf und aggregieren danach. Bei Kombi-Positionen und
  der Netto-Herausrechnung entstehen Cent-Differenzen zwischen Dashboard
  und `businesscases.csv`.

### What to build

Die USt-Aufschlüsselung des Reportings rechnet auf derselben Basis wie der
Export: Aufteilung je Positionszeile, danach Aggregation. Ob die Zeilen in
SQL entfaltet oder in Go aus den Events gerechnet werden, entscheidet sich
an der bestehenden Query-Struktur; maßgeblich ist die Gleichheit mit den
Export-Summen.

### Acceptance criteria

- [x] Für eine Sitzung mit Kombi-Positionen, Warenrücknahmen und
      Teilzahlungen gilt: Reporting-Aufschlüsselung ==
      `businesscases.csv`-Summen (gemeinsamer Testfall).
- [x] Σ(Brutto je Steuersatz) == Gesamtumsatz; Netto + Steuer == Brutto je
      Satz.
- [x] Bestehende Reporting-Tests angepasst und grün.

---

## Phase 8: Doku-Abgleich (F3, F4, C)

### Context

- `docs/compliance.md:29` — §2.2 fordert eine sofortige UI-Blockade bei
  Verbindungsverlust, die nicht existiert.
- `docs/compliance.md:434` und `docs/leitfaden/tse-einrichten.md` —
  behaupten verschlüsselte Speicherung der TSE-Keys.
- `docs/anforderungen.md:22` — Roadmap-Eintrag Q-05 (Offline-Fähigkeit).
- `docs/compliance.md:291` — §6.2 nennt Punkt als Dezimaltrennzeichen; die
  Implementierung und die amtliche `index.xml` nutzen Komma.
- `docs/compliance.md:352-353` — §6.6 beschreibt `BON_STORNO = 1`; der
  Export nutzt die zulässige Negativ-Darstellung mit `BON_STORNO = 0`.
- `docs/handbuch.md:253,333` — sechsstelliges Einmalpasswort; der Code
  erzeugt 8 Zeichen.
- `database/migrations/01_initial.up.sql:122` — Kommentar nennt die z_nr
  „sequential"; die Identity-Sequenz kann bei Fehlversuchen Lücken lassen.

### What to build

Die Dokumentation wird auf den Ist-Zustand korrigiert. §2.2 beschreibt die
tatsächliche Eigenschaft (jeder Vorgang ist ein synchroner
Backend-Request, es gibt keinen Service Worker, keine optimistischen
Updates und keinen Offline-Speicher; Offline-Erfassung ist damit technisch
ausgeschlossen), nachdem genau das im Frontend-Code verifiziert wurde;
Q-05 entfällt ersatzlos. Die TSE-Key-Passagen beschreiben die Speicherung
ehrlich (Datenbank, kein Auslesen über die API, Schutz über Server- und
Backup-Zugriffsschutz als Betreiberpflicht). §6.6 beschreibt die
implementierte Negativ-Darstellung, §6.2 das Komma, das Handbuch die
8 Zeichen, und Z-Nr-Formulierungen sagen „fortlaufend" mit Hinweis auf
mögliche technische Lücken. Verwandte Stellen (Verfahrensdokumentation,
Leitfaden) werden per Suche über alle Docs mitgezogen.

### Acceptance criteria

- [x] Verifikation dokumentiert: kein Service-Worker-Register, keine
      optimistischen Mutationen, kein Offline-Cache im Frontend (Suche
      über `frontend/src/`).
- [x] Q-05 kommt in `docs/anforderungen.md` nicht mehr vor.
- [x] Kein Dokument behauptet mehr Verschlüsselung der TSE-Keys,
      UI-Blockade, Dezimalpunkt im Export, `BON_STORNO = 1` oder
      sechsstellige Einmalpasswörter (Wortstrom-Suche über `docs/`).
- [x] Formulierungen bleiben im nüchternen Doku-Stil des Repos.
