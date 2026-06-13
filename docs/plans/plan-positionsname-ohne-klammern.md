# Plan: Einheitlicher Positionsname „Produkt Variante" (ohne Klammern)

> Source PRD: docs/prds/prd-positionsname-ohne-klammern.md

## Goal

Genau **eine kanonische Regel pro Stack** für die Zusammensetzung des
Positionsnamens, überall angewandt:

> **Positionsname = Produktname + " " + Variantenname**, mit einem einzelnen
> Leerzeichen verbunden und an den Rändern getrimmt. Keine Klammern, kein Dedup,
> keine Sonderfälle.

Damit verschwinden die Klammern auf Bons (`Pommes (mit Ketchup)` → `Pommes mit
Ketchup`) und im Reporting, die Bon-Vorschau beim Bestellen zeigt den vollen
Namen statt nur der Variante, und alle bereits korrekten Drawer rufen künftig
dieselbe zentrale Funktion auf.

## Architectural decisions

Durable decisions, die für die gesamte Umsetzung gelten:

- **Backend-Regel**: Methode `Position.Bezeichnung() string` auf der
  `kasse.Position`-Struktur. Signatur `() string`. Implementierung:
  `strings.TrimSpace(p.ProduktName + " " + p.VarianteName)`. Tiefes, schmales
  Modul — die einzige Stelle im Backend, an der Produkt- und Variantenname
  verbunden werden.
- **Frontend-Regel**: Reine Helper-Funktion
  `formatPositionName(produktName: string, varianteName: string): string` in
  `frontend/src/lib/utils.ts` (neben `formatCents`). Implementierung:
  `` `${produktName} ${varianteName}`.trim() ``.
- **Kein Dedup, keine Klammern.** Verbatim verbinden; `"Cola" + "Cola"` →
  `"Cola Cola"`. Trim verhindert nur ein überflüssiges Leerzeichen bei leerem
  Variantennamen.
- **TSE-CSV-Schicht bleibt erhalten** und umschließt die kanonische Funktion:
  Anführungszeichen-Verdopplung und `"Unbekannt"`-Fallback bei komplett leerem
  Namen bleiben in `BuildBestellungProcessData`; nur die inline-Komposition
  (inkl. `EqualFold`-Dedup) wird durch `pos.Bezeichnung()` ersetzt.

## Inventory

**Backend (Go):**

- `backend/domain/kasse/bestellung.go:11-20` — `Position`-Struct
  (`ProduktName`, `VarianteName`); Zielort für `Bezeichnung()`.
- `backend/api/bondruck/application/escpos/formatter.go:77` — Einzel-Arbeitsbon,
  `"%dx %s (%s)\n"`.
- `backend/api/bondruck/application/escpos/formatter.go:135` — Sammelbon (auch
  von `FormatDirektverkaufAbholbon`, Zeile 168-175, genutzt), `"%dx %s (%s)"`.
- `backend/api/bondruck/application/escpos/formatter.go:216` — fiskalischer
  Kassenbeleg / Stornobeleg (`FormatKassenbeleg`), `"%dx %s (%s)"`. Die
  darunterliegende Preis-/Steuerzeile (Zeile 219) bleibt unverändert.
- `backend/api/tse/application/processdata.go:83-90` — inline-Komposition mit
  `EqualFold`-Dedup + `"Unbekannt"`-Fallback; CSV-Quoting in Zeile 92-93.

**Frontend (TypeScript):**

- `frontend/src/lib/utils.ts:12` — `formatCents`; Zielort für
  `formatPositionName`.
- `frontend/src/admin/reporting/LiveReportingSection.tsx:338` —
  `{pos.varianteName ? ` (${pos.varianteName})` : ''}` (Klammer-Darstellung).
- `frontend/src/admin/reporting/ReportingResults.tsx:351` — dito.
- `frontend/src/service/components/table/BestellungDrawer.tsx:122-150` —
  `toBestellungData`; `name: v.name` (nur Variante) in Zeile 132, fließt in
  `receiptItems` (Zeile 139-143). **Der eigentliche Anzeige-Bug.** Die
  `Receipt`-Komponente bleibt unverändert.
- `frontend/src/service/components/table/drawerUtils.ts:63-69` —
  `toReceiptItems`, inline `` `${p.produktName} ${p.varianteName}` ``.
- `frontend/src/service/components/table/Zahlung.tsx:142` — inline
  `{position.produktName} {position.varianteName}`.
- `frontend/src/service/components/table/Ausgabe.tsx:123` — dito.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx:152` — dito.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:120` — dito.
- `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx:124`
  — Anzeige; `:136` + `:153` — aria-labels (`… verringern` / `… hinzufügen`).

**Tests (Prior Art):**

- `backend/domain/kasse/tisch_session_test.go` — Domänen-Tests im `kasse`-Paket.
- `backend/api/bondruck/application/escpos/formatter_test.go` — Golden-/String-
  Assertions, aktuell mit Klammern: `38`, `119`, `122`, `167`, `170`, `198`.
- `backend/api/tse/application/processdata_test.go:102-115` —
  `TestBuildBestellungProcessData_CSVFormat` (Fälle „Maß Bier" + leer und
  „Weißwurst" + „normal" bleiben mit der neuen Regel grün); `:117-131`
  Quoting-Test.
- `frontend/src/lib/utils.test.ts` — Tests neben `formatCents`.
- `frontend/src/service/components/table/drawerUtils.test.ts` — Prior Art für
  Mapping-/Helper-Tests.

## Resolved decisions

- **Granularität**: Eine einzige Phase über beide Stacks (Backend + Frontend),
  vom Nutzer gewählt.
- **Backend-Helper** als Methode `Position.Bezeichnung()` auf `kasse.Position`
  (nahe der Struct in `bestellung.go`).
- **Frontend-Helper** `formatPositionName(produktName, varianteName)` in
  `lib/utils.ts` neben `formatCents`.
- **Trim nur an den Rändern** des zusammengesetzten Strings (nicht je Teil) —
  entspricht der kanonischen Regel; in der Praxis sind Namen validiert.
- **Kein Dedup** mehr in der TSE-Komposition; `"Unbekannt"`-Fallback und
  CSV-Quoting bleiben als TSE-Schicht erhalten.
- **Out of Scope** (unverändert): ProductList-Auswahl-UI, Produktanlage/
  -bearbeitung, Validierungsregeln, bereits gespeicherte/signierte Events,
  Preis-/Steuerzeilen des Kassenbelegs.

## Open questions / Risks

- **Fiskale Konsequenz (bewusst akzeptiert)**: Im praktisch nicht vorgesehenen
  Sonderfall `variante == produkt` ändert sich die künftig signierte
  Bezeichnung von z. B. „Cola" zu „Cola Cola". Variantennamen sind Pflicht
  (min. 3 Zeichen); ein leerer Variantenname kommt im `kasse`-Pfad nicht vor.
  Bereits signierte (immutable) Events bleiben unverändert.

---

## Phase 1: Kanonische Positionsbezeichnung über beide Stacks

**User stories**: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18

### Context

Backend:

- `backend/domain/kasse/bestellung.go:11-20` — `Position`-Struct; hier neue
  Methode.
- `backend/api/bondruck/application/escpos/formatter.go:77,135,216` — drei
  Klammer-Formatierungen → `"%dx %s"` mit `pos.Bezeichnung()`.
- `backend/api/tse/application/processdata.go:83-90` — inline-Komposition →
  `pos.Bezeichnung()`; `"Unbekannt"`-Fallback (Zeile 88-90) und CSV-Quoting
  (Zeile 92-93) bleiben.

Frontend:

- `frontend/src/lib/utils.ts:12` — neben `formatCents` neuer Helper.
- `frontend/src/admin/reporting/LiveReportingSection.tsx:338`,
  `frontend/src/admin/reporting/ReportingResults.tsx:351` — Klammer-Darstellung
  → Helper.
- `frontend/src/service/components/table/BestellungDrawer.tsx:132` —
  `name: v.name` → `name: formatPositionName(p.name, v.name)`.
- `frontend/src/service/components/table/drawerUtils.ts:65`,
  `Zahlung.tsx:142`, `Ausgabe.tsx:123`, `HistorieUmbuchungDrawer.tsx:152`,
  `HistorieStornierungDrawer.tsx:120`,
  `DirektverkaufStornoDrawer.tsx:124,136,153` — inline-Kompositionen → Helper
  (visuell unverändert; aria-labels eingeschlossen).

### What to build

Eine einzige, durchgängige Umstellung auf die kanonische Regel pro Stack:

**Backend** — `Position.Bezeichnung()` einführen und an allen drei
Bondruck-Stellen sowie in der TSE-`processData`-Komposition verwenden. Die
Bons (Einzel-Arbeitsbon, Sammelbon, Direktverkauf-Abholbon, Kassenbeleg,
Stornobeleg) drucken danach `{Produkt} {Variante}` ohne Klammern. Die
TSE-signierte Bezeichnung verwendet dieselbe Funktion; CSV-Quoting und
`"Unbekannt"`-Fallback bleiben als umschließende Schicht erhalten.

**Frontend** — `formatPositionName` einführen und alle Anzeigepfade darauf
umstellen: Reporting (Klammern entfallen), Bon-Vorschau beim Bestellen (zeigt
nun den vollständigen Namen statt nur der Variante), sowie die bereits
korrekten Drawer (Zahlung, Ausgabe, Historie-Stornierung, Historie-Umbuchung,
Direktverkauf-Storno inkl. aria-labels) und `toReceiptItems` — letztere
visuell unverändert, Ziel ist zentrale Wartbarkeit.

Tests prüfen externes Verhalten (zurückgegebene/gerenderte Strings bzw.
Bon-Ausgabe), nicht, ob eine bestimmte Hilfsfunktion aufgerufen wurde.

### Acceptance criteria

**Backend:**

- [ ] `Position.Bezeichnung()` existiert auf `kasse.Position` und liefert
      `TrimSpace(ProduktName + " " + VarianteName)`.
- [ ] Unit-Tests für `Bezeichnung()`: Normalfall („Pommes" + „mit Ketchup" →
      „Pommes mit Ketchup"), gleichlautend („Cola" + „Cola" → „Cola Cola",
      kein Dedup), leerer Variantenname („Maß Bier" + „" → „Maß Bier", kein
      Trailing-Space).
- [ ] Alle drei Formatter-Stellen verwenden `"%dx %s"` mit `pos.Bezeichnung()`;
      Bon-Ausgabe enthält keine Klammern mehr um die Variante.
- [ ] `formatter_test.go`-Assertions auf das klammerlose Format aktualisiert
      (Einzel-Arbeitsbon, Sammelbon, Kassenbeleg, Stornobeleg) — z. B.
      „3x Pommes gross", „1x Bratwurst mit Brot".
- [ ] `BuildBestellungProcessData` nutzt `pos.Bezeichnung()`; der
      `EqualFold`-Dedup ist entfernt; `"Unbekannt"`-Fallback und
      Anführungszeichen-Verdopplung bleiben funktional.
- [ ] Bestehende `processdata_test.go`-Fälle (leere/abweichende Variante)
      bleiben grün; Quoting-Test unverändert.

**Frontend:**

- [ ] `formatPositionName(produktName, varianteName)` existiert in
      `lib/utils.ts` und liefert `` `${produktName} ${varianteName}`.trim() ``.
- [ ] Unit-Tests für `formatPositionName`: Normalfall, leerer Variantenname
      (nur Produktname, kein Trailing-Space), gleichlautende Namen.
- [ ] Live-Reporting und Reporting-Ergebnisse zeigen `{Produkt} {Variante}`
      ohne Klammern (Helper statt bedingter Klammer-Darstellung).
- [ ] `toBestellungData` setzt das `name`-Feld der Bon-Vorschau auf den
      zusammengesetzten Namen; Test belegt „{Produkt} {Variante}" statt nur der
      Variante. `Receipt`-Komponente unverändert.
- [ ] `toReceiptItems`, Zahlung, Ausgabe, Historie-Stornierung,
      Historie-Umbuchung und Direktverkauf-Storno (inkl. beider aria-labels)
      rufen `formatPositionName` auf; Darstellung visuell unverändert.

**Querschnitt:**

- [ ] `make check` (Backend + Frontend Lint/Build/Unit-Tests) ist grün.
- [ ] Die Zusammensetzungsregel existiert pro Stack an genau einer Stelle; eine
      künftige Format-Änderung ist eine Ein-Zeilen-Änderung pro Stack.
