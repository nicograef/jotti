# Plan: Kassenbeleg — Steueraufschlüsselung & modulare TSE-Bereitschaft

> Source PRD: `docs/prds/prd-kassenbeleg.md`

## Goal

Den bestehenden Kassenbeleg um die gesetzlich vorgeschriebene Steueraufschlüsselung erweitern (Steuerkennzeichen je Position + Steuermatrix im Belegfuß) und einen modularen Einhängepunkt für die spätere TSE-Integration vorbereiten, sodass der Beleg heute vollständig ohne TSE funktioniert.

## Architectural decisions

- **Steuerkennzeichen**: `A` = 19 % (Regel), `B` = 7 % (ermäßigt), `C` = 0 % (befreit), `A/B` = kombi (70/30-Aufteilung, Verweis auf Matrix im Belegfuß).
- **Steuermatrix-Modul**: Neue Funktion `Steuermatrix()` im bestehenden Paket `domain/steuer/`. Eingabe: Liste von Positionen (Brutto + Steuersatz). Ausgabe: aggregierte Matrixzeilen (Netto, Steuer, Brutto je effektivem Satz). Nutzt intern `Aufteilen()`.
- **TSE-Abschnitt**: Optionaler Pointer `*TSEAbschnitt` auf `KassenbelegData`. `nil` = kein TSE-Block auf dem Beleg. Felder gemäß `docs/handbuch.md` §3.13.
- **Keine Schema- oder Event-Änderungen**: Der Steuersatz ist bereits in den Events persistiert (`positionEventData.steuersatz`). Es wird ausschließlich die Lese-/Decode-Seite (`zahlungPositionData`) korrigiert.

## Inventory

- `backend/api/bondruck/application/escpos/formatter.go:17-28` — `KassenbelegData` struct + `FormatKassenbeleg()` Funktion (Zeilen 166–217)
- `backend/api/bondruck/application/escpos/formatter_test.go:177-228` — Bestehende Kassenbeleg-Tests (Pflichtfelder + CutPaper)
- `backend/api/table/application/command.go:86-102` — `zahlungKassiertV1Data` + `zahlungPositionData` (Steuersatz fehlt hier)
- `backend/api/table/application/kassenbeleg_command.go:17-28` — `toKassePositionen()` Konvertierung
- `backend/domain/steuer/steuer.go:42-57` — `Aufteilen()` mit kombi 70/30 Split
- `backend/domain/steuer/steuer_test.go` — Bestehende Tests für `Aufteilen()`
- `backend/domain/kasse/bestellung.go:10-19` — `Position` struct (hat `Steuersatz` Feld)
- `backend/domain/kasse/bestellung.go:24-32` — `positionEventData` (hat `steuersatz` JSON-Tag)
- `docs/compliance.md:317-376` — §5.2 Pflichtangaben auf dem Beleg
- `docs/steuerrecht.md:114-140` — §6 Belegausweis und Pflichtangaben
- `docs/handbuch.md:429-520` — §3.13 TSE-Architektur (TSEData struct)
- `docs/anforderungen.md:224+236` — F-03 Status und Akzeptanzkriterien

## Resolved decisions

- Steuerkennzeichen-Buchstaben: A = 19 %, B = 7 %, C = 0 % (befreit).
- Kombi-Positionen zeigen auf der Positionszeile das kombinierte Kennzeichen `A/B`. Die exakte Aufteilung steht in der Steuermatrix.
- Steuermatrix-Modul lebt als neue exportierte Funktion im bestehenden `domain/steuer/`-Paket (kein neues Sub-Package).
- Steuermatrix führt nur effektive Sätze (19 %, 7 %, 0 %) — keine eigene „kombi"-Zeile.
- `KassenbelegData` bekommt ein neues Feld `Steuermatrix []steuer.Aufteilung` (vorab berechnet im Command) und `TSE *TSEAbschnitt` (optional).
- Kein Eventformat-Wechsel, keine Schemaänderung, keine Migration.

---

## Phase 1: Steuermatrix-Modul

**User stories**: 11, 12, 13

### Context

- `backend/domain/steuer/steuer.go:42-57` — `Aufteilen()` berechnet pro Einzelbetrag die Aufteilung nach Satz (inkl. kombi → zwei Teilzeilen). Wird vom neuen Modul pro Position aufgerufen.
- `backend/domain/steuer/steuer_test.go` — Testmuster (table-driven, `//go:build unit`).

### What to build

Eine neue exportierte Funktion im `domain/steuer/`-Paket, die aus einer Liste von Positionen (je Bruttobetrag = Einzelpreis × Menge und Steuersatz) eine nach effektivem Steuersatz aggregierte Steuermatrix berechnet. Kombi-Positionen werden durch `Aufteilen()` auf die 7-%- und 19-%-Zeilen verteilt. Die Ausgabe enthält je Satz: Brutto, Netto, Steuer. Die Summe aller Bruttowerte muss exakt dem Gesamtbetrag entsprechen (centgenau).

### Acceptance criteria

- [x] Neue Funktion akzeptiert eine Slice-artige Eingabe mit Brutto (Einzelpreis × Menge) und Steuersatz pro Position
- [x] Gibt eine nach effektivem Satz (19 %, 7 %, 0 %) aggregierte Liste von Matrixzeilen zurück (Netto, Steuer, Brutto)
- [x] Kombi-Positionen erzeugen über `Aufteilen()` anteilige Beiträge zu 7 %- und 19 %-Zeilen
- [x] Summe aller Matrix-Bruttowerte = Summe aller Eingabe-Bruttowerte (centgenau, keine Rundungsdifferenz)
- [x] Unit-Tests: einzelner Satz, gemischte Sätze, kombi mit korrekter 70/30-Aufteilung, befreit (0 %), leere Eingabe
- [x] `make test` grün

---

## Phase 2: Steuersatz durchreichen + Beleg-Steuerausweis

**User stories**: 8, 9, 10, 11

### Context

- `backend/api/table/application/command.go:93-100` — `zahlungPositionData` ohne `Steuersatz`
- `backend/api/table/application/kassenbeleg_command.go:17-28` — `toKassePositionen()` ohne Steuersatz-Mapping
- `backend/api/bondruck/application/escpos/formatter.go:166-217` — `FormatKassenbeleg()` mit TODO-Markern für F-07/F-02
- `backend/api/bondruck/application/escpos/formatter.go:17-28` — `KassenbelegData` struct

### What to build

1. **Lese-Seite:** `Steuersatz`-Feld in `zahlungPositionData` ergänzen und in `toKassePositionen()` durchreichen, sodass `kasse.Position.Steuersatz` beim Belegdruck befüllt ist.

2. **Formatter:** Pro Position ein Steuerkennzeichen (A/B/C/A/B) hinter dem Preis drucken. Im Belegfuß nach GESAMT die Steuermatrix drucken: je Satz eine Zeile mit Kennzeichen, Netto, Steuer, Brutto. Die Matrix wird im Command vorab über die Funktion aus Phase 1 berechnet und als Feld auf `KassenbelegData` übergeben.

3. **Command:** Im `KassenbelegDrucken`-Command die Steuermatrix berechnen und dem `KassenbelegData`-Struct mitgeben.

4. **TODO-Marker** für F-07 entfernen (erledigt).

### Acceptance criteria

- [x] `zahlungPositionData` decodiert `steuersatz` aus dem Event-JSON
- [x] `toKassePositionen()` setzt `Position.Steuersatz`
- [x] Jede Positionszeile im gedruckten Beleg enthält ein Steuerkennzeichen (A, B, C, oder A/B)
- [x] Belegfuß enthält Steuermatrix mit Netto, Steuer und Brutto je Satz
- [x] Summe der Matrix-Bruttowerte = GESAMT-Betrag
- [x] Bestehende Tests (`ContainsPflichtfelder`, `EndsWithCutPaper`) bleiben grün
- [x] Neue Formatter-Tests prüfen Steuerkennzeichen-Präsenz und Matrixzeilen
- [x] End-to-End-Test (Command-Level): bekannte Positionen → korrekter Steuerausweis im Beleg
- [x] `make lint && make test` grün

---

## Phase 3: Optionaler TSE-Abschnitt

**User stories**: 14, 15, 16

### Context

- `backend/api/bondruck/application/escpos/formatter.go:207-208` — TODO(F-02) Marker
- `docs/handbuch.md:429-520` — §3.13 TSE-Architektur, TSEData-Felder
- `docs/compliance.md:329-337` — TSE-Pflichtdaten auf dem Beleg

### What to build

1. **Datenstruktur:** Neuer Struct-Typ `TSEAbschnitt` mit den TSE-Pflichtfeldern (Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkt Vorgangsbeginn, Zeitpunkt Vorgangsende, Signatur). Optionales Pointer-Feld `TSE *TSEAbschnitt` auf `KassenbelegData`.

2. **Formatter:** Wenn `TSE != nil`, drucke einen TSE-Block im Belegfuß (nach Steuermatrix, vor "Vielen Dank!"). Wenn `nil` (heutiger Zustand), drucke nichts — Beleg bleibt unverändert.

3. **TODO-Marker** für F-02 entfernen (Einhängepunkt ist nun implementiert).

### Acceptance criteria

- [x] `TSEAbschnitt`-Struct mit Feldern: TransaktionNr, Signaturzaehler, TSESeriennummer, ZeitpunktBeginn, ZeitpunktEnde, Signatur
- [x] `KassenbelegData.TSE` ist ein optionaler Pointer (`*TSEAbschnitt`)
- [x] Formatter druckt TSE-Block nur wenn `TSE != nil`
- [x] Test: `TSE = nil` → kein TSE-Block im Beleg (Regression: Beleg identisch zu vorher)
- [x] Test: `TSE` gesetzt → alle TSE-Pflichtfelder im Beleg enthalten
- [x] Beleg endet weiterhin mit "Vielen Dank!" + CutPaper (bestehender Test bleibt grün)
- [x] `make lint && make test` grün

---

## Phase 4: Dokumentation angleichen

**User stories**: 17

### Context

- `docs/anforderungen.md:222-265` — F-03 Statuszeile und Akzeptanzkriterien

### What to build

1. **Status-Zeile** von F-03 in der Übersichtstabelle aktualisieren (von "🔲 Offen" auf den realen Stand: Basis + Steuerausweis umgesetzt, TSE-Block vorbereitet/offen).

2. **Akzeptanzkriterien** abhaken, die nach Phase 1–3 erfüllt sind. Das Kriterium "Mit F-02 (TSE): Beleg enthält TSE-Pflichtfelder" bleibt offen (wird erst mit TSE-Anbindung erfüllt).

3. **Beschreibungstext** anpassen: "heute existiert nur der nicht-fiskalische Arbeitsbon" stimmt nicht mehr.

### Acceptance criteria

- [ ] F-03 Status in der Übersichtstabelle spiegelt die Realität wider
- [ ] Erfüllte Akzeptanzkriterien sind abgehakt
- [ ] Beschreibungstext korrigiert (Beleg existiert, nur TSE-Block fehlt noch)
- [ ] Keine inhaltlichen Widersprüche in `docs/anforderungen.md`
