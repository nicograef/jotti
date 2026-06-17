# Plan: DSFinV-K-Export-Findings beheben

> Source PRD: n/a (aus Code-Audit des DSFinV-K-Exports, 2026-06-17)

## Goal

Die fünf Findings des DSFinV-K-Audits beheben, damit der Export der DSFinV-K
v2.5 entspricht:

1. **Correctness 1 (fiskalisch):** Tisch-Bestellungen werden als
   `Forderungsentstehung` *und* die Zahlung als `Umsatz` verbucht — ohne je eine
   `Forderungsauflösung`. Das verdoppelt Umsatz und USt in `businesscases.csv`
   und hinterlässt nie aufgelöste Phantom-Forderungen. Umstellung auf das
   spec-konforme **Revenue-at-payment-Modell** (Tz 2.7.2 „Durchbedienen mit
   Bestell-Absicherung“).
2. **Correctness 2:** `TSE_ZEITFORMAT` deklariert `unixTime`, die Werte in
   `TSE_TA_START/ENDE` sind aber ISO-8601 — Widerspruch.
3. **Correctness 3:** `tse.csv` gibt nur 2 der Zertifikatsspalten aus;
   Zertifikate > 2000 Zeichen werden still abgeschnitten.
4. **Correctness 4:** Vorgänge im TSE-Ausfallfenster fehlen in
   `transactions_tse.csv`, statt eine `TSE_TA_FEHLER`-Zeile zu tragen.
5. **Minor:** `Z_BUCHUNGSTAG` leer; Frontend-Meldung für
   `kassensitzung_nicht_gefunden` fehlt; `UST_SCHLUESSEL` 5/6 gegen Anlage 2
   bestätigen.

## Recherche-Grundlage (Finding 1)

Belegt aus DSFinV-K v2.4 (struktur­identisch zu 2.5):

- **`Forderungsentstehung` ist der Revenue-at-delivery-Mechanismus** (S.53):
  setzt voraus, dass „die Warenbewegung bereits erfolgt ist“, bucht `Umsatz +X`
  **und** `Forderungsentstehung −X` und **muss** durch eine
  `Forderungsauflösung` mit Referenz auf den Ursprung aufgelöst werden (S.54,
  Anhang-Beispiel S.119). jotti tut weder das eine noch das andere.
- **`AVBestellung` ist kein Geschäftsvorfall** (S.43): „Lieferungen oder
  Leistungen werden im Rahmen des Bestellprozesses noch nicht ausgeführt.“
  `UMS_BRUTTO` = „Gesamtsumme über alle Geschäftsvorfalltypen eines Vorgangs“
  (S.79) → für eine reine Bestellung **0.00**.
- **Tz 2.7.2 erlaubt Revenue-at-payment für die Gastronomie:** „Die Transaktion
  für den prüfbaren ‚Kassenbeleg‘ wird dann erst bei Rechnungserstellung
  gestartet.“ Bestellungen werden nur TSE-gesichert; der Umsatz entsteht einmal,
  bei der Zahlung. Passt zur §-20-UStG-Istversteuerung typischer Vereine (die
  `Forderungs`-Maschinerie inkl. ID 7 existiert laut S.26 nur, um die USt
  *früher* als die Zahlung auszulösen — das Gegenteil des jotti-Ziels).

## Architectural decisions

Gelten über alle Phasen:

- **Revenue-at-payment (Model A):** `zahlung-kassiert` ist der **einzige**
  umsatzwirksame Vorgang (`BON_TYP=Beleg`, `GV_TYP=Umsatz`, `Zahlart=Bar`).
  `direktverkauf-getaetigt`/`-storniert` und die Bargeldbewegungen bleiben
  unverändert.
- **`bestellung-aufgenommen` und `stornierung-erteilt` werden geldneutrale
  `AVBestellung`-Vorgänge:** Sie erscheinen in `transactions.csv`
  (`BON_TYP=AVBestellung`, `UMS_BRUTTO=0.00`), `transactions_tse.csv`,
  `allocation_groups.csv` und — Entscheidung unten — `lines.csv` (informativ,
  mit Preisen). Sie tragen **nichts** zu `transactions_vat.csv`,
  `lines_vat.csv`, `datapayment.csv`, `businesscases.csv`, `payment.csv` bei.
  Der Storno trägt zusätzlich `BON_STORNO=1` und eine `references.csv`-Zeile auf
  die Ursprungsbestellung.
- **`Forderungsentstehung`/`Forderungsauflösung` entfallen vollständig** aus dem
  Export (Konstanten `gvTypForderung`, `zahlartForderung` und ihre
  Sortier-Einträge werden entfernt). jotti hat keinen Kredit-/Rechnungs-Flow,
  der sie rechtfertigt (YAGNI).
- **`TSE_ZEITFORMAT` wird eine feste Konstante `"utcTime"`** (analog zu
  `tsePDEncoding = "UTF-8"`), passend zur RFC3339-Normalisierung der Log-Zeiten.
- **`tse.csv` deklariert fünf Zertifikatsspalten `TSE_ZERTIFIKAT_I…_V`**, per
  Schleife befüllt.
- **TSE-Ausfall:** Ein noch nicht nachsignierter Vorgang erscheint mit
  gesetztem `TSE_TA_FEHLER` statt zu fehlen.
- **Keine Daten-Migration nötig:** Der Export ist eine reine Funktion über die
  Events; das geänderte Mapping interpretiert auch bestehende, persistierte
  Events korrekt neu.

## Inventory

- `backend/domain/dsfinvk/mapper.go` — Kern-Mapper. Relevante Stellen:
  - `belegeFromEvents` `:182-359` — Event→Beleg; Bestellung `:201-217`,
    Zahlung `:225-241`, Stornierung `:243-267`.
  - GV-/Zahlart-Konstanten `:32-43`; `gvTypReihenfolge` `:893-900`;
    `zahlartReihenfolge` `:967-970`.
  - `buildTransactions` `:623-645` (UMS_BRUTTO), `buildTransactionsVat`
    `:683-703`, `buildDatapayment` `:711-729`, `buildLines` `:773-809`,
    `buildLinesVat` `:817-848`, `buildBusinesscases` `:913-959`, `buildPayment`
    `:974-1008`, `barbestand` `:1039-1048`.
  - `buildTSE` `:594-609`, `tseColumns` `:587-592`, `certChunk` `:1084-1094`.
  - `buildTransactionsTSE` `:858-881` (`TSE_TA_FEHLER` `:869`).
  - `buildCashpointclosing` `:476-495` (`Z_BUCHUNGSTAG` `:481`).
- `backend/domain/dsfinvk/dsfinvk.go` — `Version` `:20`, `tsePDEncoding` `:46`,
  `ustSchluessel`/`ustBeschreibung` `:104-131`.
- `backend/domain/dsfinvk/mapper_test.go` — Golden-File-Tests; betroffen u. a.
  `TestMapTischablaufTrennt` `:250`, `TestMapTischStornoNegativeWithReference`
  `:379`, `TestMapKassenabschlussGemischteSitzung` `:755`,
  `TestMapBarverkaufGoldenRows` `:86` (tse.csv-Zeile `:114-116`).
- `backend/api/table/application/command.go:680`,
  `backend/api/table/application/errors.go:47` — Storno nur auf *unbezahlte*
  Positionen (belegt die Geldneutralität des Tisch-Stornos).
- `backend/domain/kasse/tse_embedding.go:55` — `TSEAusfall`-Flag wird gesetzt;
  `backend/domain/kasse/tse_data.go` — `TSEData`.
- `backend/api/tse/application/signing.go:161-165` — `timeString` (RFC3339);
  `backend/repository/tse_repo/fiskaly_setup.go:193` — Quelle des
  `LogTimeFormat` (`unixTime`).
- `backend/domain/settings/tse_stammdaten.go:17` — `LogTimeFormat`-Feld; nach
  Phase 2 nur noch Schreib-Feld (einziger Leser war `mapper.go:598`).
- `frontend/src/admin/reporting/hooks.ts:48-53` — Fehler-Mapping des Exports.

## Resolved decisions

- **Finding 1 = Model A (Revenue-at-payment).** Spec-konform per Tz 2.7.2,
  passend zur Istversteuerung; bestätigt durch Recherche oben.
- **AVBestellung behält `lines.csv`** (Positionen mit Preisen, geldneutral) —
  für „inhaltlich reproduzierbar“ (S.14). Akzeptiertes Restrisiko: ein strikter
  Validator könnte `Σ(Bonpos × Preis)` gegen `UMS_BRUTTO=0` monieren; vertretbar,
  da `AVBestellung` per Spec geldneutral ist.
- **Findings 2–4 und Minor-Notes** werden gemäß den Audit-Empfehlungen
  umgesetzt.
- **`LogTimeFormat`-Feld bleibt** in den Stammdaten (treuer Mitschnitt dessen,
  was fiskaly meldet), wird vom Export aber nicht mehr gelesen.

## Open questions / Risks

- **Validator-Reconciliation** der geldneutralen `AVBestellung` mit `lines.csv`
  (s. o.) — beim ersten echten IDEA-/fiskaly-Validatorlauf gegenprüfen.
- **`GV_TYP` der `AVBestellung`-Positionen in `lines.csv`:** Wert in Phase 1
  festlegen (Spec-AV-Beispiele nutzen `"Umsatz"` mit 0-Beträgen) und per Golden
  Test fixieren.
- **`UST_SCHLUESSEL` 5/6:** Anlage 2 lag dem extrahierten PDF nicht bei; gegen
  die DFKA-Taxonomie-Anlage final bestätigen (Erwartung: keine Code-Änderung).

---

## Phase 1: Revenue-at-payment im Mapper (Finding 1)

### Context

- `backend/domain/dsfinvk/mapper.go:201-267` — Bestellung/Zahlung/Stornierung
  → Beleg.
- `backend/domain/dsfinvk/mapper.go:32-43` — GV-/Zahlart-Konstanten;
  `:893-900`, `:967-970` — Sortier-Maps.
- `backend/domain/dsfinvk/mapper.go:623-1008` — Tabellen-Builder, die das
  Vorzeichen/den Geldfluss aus dem Beleg ableiten.
- `backend/domain/dsfinvk/mapper_test.go:250,379,755` — umzuschreibende Golden
  Tests.
- `backend/api/table/application/command.go:680` — belegt: Tisch-Storno nur auf
  unbezahlte Positionen.

### What to build

Das Tisch-Mapping auf Revenue-at-payment umstellen. `bestellung-aufgenommen`
und `stornierung-erteilt` werden geldneutrale `AVBestellung`-Vorgänge: sie
liefern weiterhin `transactions.csv` (jetzt `UMS_BRUTTO=0.00`),
`transactions_tse.csv`, `allocation_groups.csv`, `lines.csv` (Positionen mit
Preisen, informativ) und — beim Storno — `references.csv` + `BON_STORNO=1`,
tragen aber **nichts** zu `transactions_vat`, `lines_vat`, `datapayment`,
`businesscases` und `payment` bei. `zahlung-kassiert` bleibt der einzige
`Umsatz`/`Bar`-Vorgang. Die Konstanten und Sortier-Einträge für
`Forderungsentstehung`/`Forderungsauflösung` werden entfernt. Umsetzung z. B.
über ein `beleg`-Flag (analog zu `nichtSteuerbar`), das die Geld-Builder
überspringen lässt. `barbestand` und der Kassenabschluss bleiben dadurch
korrekt (keine Forderung mehr im Bestand).

### Acceptance criteria

- [x] `bestellung-aufgenommen` erzeugt `transactions.csv` mit
      `BON_TYP=AVBestellung`, `UMS_BRUTTO=0.00`, eine `transactions_tse.csv`- und
      `allocation_groups.csv`-Zeile sowie `lines.csv`-Positionen mit Preisen.
- [x] `bestellung-aufgenommen` erzeugt **keine** Zeile in
      `transactions_vat.csv`, `lines_vat.csv`, `datapayment.csv`,
      `businesscases.csv`, `payment.csv`.
- [x] `stornierung-erteilt` ist geldneutral, trägt `BON_STORNO=1` und eine
      `references.csv`-Zeile auf die Ursprungsbestellung.
- [x] `businesscases.csv` enthält keinen `GV_TYP=Forderungsentstehung` mehr;
      `payment.csv` keine `Zahlart=Forderungsentstehung`.
- [x] Für eine Bestellung + zugehörige Zahlung (gleiche Ware) weist
      `businesscases.csv` genau einmal `Umsatz` aus (keine Verdopplung).
- [x] Invariante erhalten: `Σ transactions.UMS_BRUTTO == Σ businesscases.Z_UMS_BRUTTO == Σ payment.Z_ZAHLART_BETRAG`.
- [x] `Forderungsentstehung`/`Forderungsauflösung` kommen nirgends mehr im
      `dsfinvk`-Paket vor.
- [x] Golden-Tests (`TestMapTischablaufTrennt`,
      `TestMapTischStornoNegativeWithReference`,
      `TestMapKassenabschlussGemischteSitzung`) aktualisiert und grün; der
      irreführende Kommentar „und Forderungsauflösung“ entfernt/korrigiert.

---

## Phase 2: TSE_ZEITFORMAT als feste Konstante (Finding 2)

### Context

- `backend/domain/dsfinvk/mapper.go:598` — derzeit
  `s.TSEStammdaten.LogTimeFormat` (`unixTime`).
- `backend/domain/dsfinvk/dsfinvk.go:46` — Vorbild `tsePDEncoding = "UTF-8"`.
- `backend/api/tse/application/signing.go:161-165` — Log-Zeiten sind RFC3339.
- `backend/domain/dsfinvk/mapper_test.go:115` — tse.csv-Golden-Zeile.

### What to build

Eine Paket-Konstante `tseZeitformat = "utcTime"` einführen und in `buildTSE`
statt `s.TSEStammdaten.LogTimeFormat` verwenden, sodass `TSE_ZEITFORMAT` zu den
tatsächlich ausgegebenen ISO-8601-Werten in `TSE_TA_START/ENDE` passt. Das
`LogTimeFormat`-Stammdatenfeld bleibt erhalten (Mitschnitt der
fiskaly-Meldung), wird aber nicht mehr gelesen.

### Acceptance criteria

- [x] `tse.csv` führt `TSE_ZEITFORMAT=utcTime`.
- [x] Genauer Wert gegen die DSFinV-K-Spec bestätigt (`utcTime` vs. exakte
      Schreibweise). DSFinV-K Stamm_TSE lässt für `TSE_ZEITFORMAT` genau
      `unixTime`/`utcTime`/`unknown` zu; RFC3339-UTC entspricht `utcTime`.
- [x] Golden-Test angepasst; `go test ./domain/dsfinvk/...` grün.

---

## Phase 3: tse.csv-Zertifikatsspalten I–V (Finding 3)

### Context

- `backend/domain/dsfinvk/mapper.go:587-592` — `tseColumns` (aktuell `_I`, `_II`).
- `backend/domain/dsfinvk/mapper.go:599` — zwei hartcodierte `certChunk`-Aufrufe.
- `backend/domain/dsfinvk/mapper.go:1084-1094` — `certChunk`.
- `backend/domain/dsfinvk/mapper_test.go:114-116` — tse.csv-Golden-Zeile.

### What to build

`tseColumns` um `TSE_ZERTIFIKAT_III/_IV/_V` ergänzen und die Zertifikats-Chunks
in `buildTSE` per Schleife (`certChunk(cert, 0..4)`) ausgeben, statt zweier
Aufrufe. Damit deckt die Ausgabe die fünf Spalten der Spec ab und schneidet
längere Zertifikate nicht mehr still ab.

### Acceptance criteria

- [x] `tse.csv` deklariert und befüllt fünf `TSE_ZERTIFIKAT_I…_V`-Spalten.
- [x] Spaltenzahl gegen DSFinV-K-Anhang (Stamm_TSE) bestätigt. DSFinV-K v2.4
      §3.2.7 definiert `TSE_ZERTIFIKAT_I/_II` (je 1000 Zeichen base64) und erlaubt
      bei Zertifikaten > 2000 Zeichen die Felder `_III/_IV/_V`; die `index.xml`
      wird aus `tseColumns` abgeleitet und deckt die Ergänzung automatisch ab.
- [x] Ein > 2000 Zeichen langes Test-Zertifikat wird vollständig (über mehrere
      Chunks) ausgegeben, nicht abgeschnitten.
- [x] Golden-Test (`CERTBASE64` in `_I`, Rest leer) angepasst; Tests grün.

---

## Phase 4: TSE-Ausfall als TSE_TA_FEHLER-Zeile (Finding 4)

### Context

- `backend/domain/dsfinvk/mapper.go:862-869` — überspringt `tse == nil`,
  `TSE_TA_FEHLER` stets `""`.
- `backend/domain/dsfinvk/mapper.go:67-88` — `beleg`-Struct (Feld ergänzen).
- `backend/domain/dsfinvk/mapper.go:182-359` — `belegeFromEvents` liest aktuell
  `TSEAusfall` nicht.
- `backend/domain/kasse/tse_embedding.go:55` — Quelle des `TSEAusfall`-Flags.

### What to build

Das `TSEAusfall`-Flag der Event-Daten in den `beleg` durchreichen.
`buildTransactionsTSE` gibt für einen Beleg mit `tse == nil`, der ein TSE-Ausfall
war, eine Zeile mit gesetztem `TSE_TA_FEHLER` (z. B. `"TSE-Ausfall"`) und leerer
Signatur aus — so hat jeder Bonkopf-Vorgang eine korrespondierende
TSE-Zeile. Noch nicht nachsignierte Ausfall-Vorgänge sind damit kenntlich statt
unsichtbar.

### Acceptance criteria

- [x] Ein unsigniert persistierter Ausfall-Vorgang (kein Backfill) erzeugt eine
      `transactions_tse.csv`-Zeile mit `TSE_TA_FEHLER` gesetzt und leerer
      `TSE_TA_SIG`.
- [x] Nach Nachsignierung erscheint weiterhin die vollständige signierte Zeile
      (bestehendes Verhalten, `TestMapNachsigniertVorgang` bleibt grün).
- [x] Jeder Bonkopf-Vorgang hat genau eine `transactions_tse.csv`-Zeile.
- [x] Neuer Test deckt den Ausfall-ohne-Backfill-Fall ab.

---

## Phase 5: Minor-Notes (Finding 5)

### Context

- `backend/domain/dsfinvk/mapper.go:481` — `Z_BUCHUNGSTAG` aktuell `""`.
- `frontend/src/admin/reporting/hooks.ts:48-53` — Export-Fehler-Mapping.
- `backend/domain/dsfinvk/dsfinvk.go:104-131` — `UST_SCHLUESSEL`-Mapping.

### What to build

Drei kleine, unabhängige Verbesserungen: `Z_BUCHUNGSTAG` aus dem
Erstellungsdatum (`YYYY-MM-DD`) befüllen; im Frontend eine eigene Meldung für
den Fehlercode `kassensitzung_nicht_gefunden` ergänzen (defensiv, neben der
bestehenden `leere_kassensitzung`-Behandlung); `UST_SCHLUESSEL` 5/6 gegen
Anlage 2 verifizieren und das Ergebnis dokumentieren (voraussichtlich keine
Code-Änderung).

### Acceptance criteria

- [ ] `cashpointclosing.csv` führt `Z_BUCHUNGSTAG` im Format `YYYY-MM-DD`;
      Golden-Test angepasst.
- [ ] Frontend zeigt für `kassensitzung_nicht_gefunden` eine spezifische
      Meldung statt der generischen.
- [ ] `UST_SCHLUESSEL` 5 (nicht steuerbar) / 6 (umsatzsteuerfrei) gegen Anlage 2
      bestätigt; Befund in `docs/compliance.md` oder als Code-Kommentar notiert.

---

## Verifikation (alle Phasen)

- [ ] `make build` ✓
- [ ] `make lint-backend` ✓ (`go vet`, `goimports`)
- [ ] `make test` ✓ (`go test -tags=unit -race ./...`)
- [ ] `make test-frontend` ✓ (für Phase 5)
