# Plan: Kommentar je Position

> Source PRD: docs/prds/prd-kommentar-je-position.md

## Goal

Die Servicekraft hängt im Bestell- und im Direktverkauf-Abschluss an jede Position einen eigenen Kommentar. Er druckt nur auf dem Arbeits- oder Abholbon der Station, die die Position zubereitet. Die Historie zeigt ihn unter seiner Position. Der Bestellkommentar bleibt, wie er ist. Kassenbeleg, TSE-Daten, DSFinV-K-Export, Projektion, Reporting und Schema bleiben unverändert. Ohne Positionskommentar ist jedes neue Event byte-identisch zum Bestandsformat.

Nico hat das PRD mit seinen Empfehlungen freigegeben: additives Feld als Freeze-Ausnahme (Ansatz A), kein DSFinV-K-Export, Kommentar auf jedem Stückbon.

## Architectural decisions

**Voraussetzung „Zeichen gegen Bytes".** Fakt: Das Backend zählt seit `e999f531` Zeichen (`backend/internal/zeichen`). Das Frontend zählt seit `755f6b0d` UTF-8-Bytes (`frontend/src/lib/utf8.ts`). Beide Commits landeten am selben Tag und haben sich gekreuzt. Ein Kommentar aus 60 Umlauten ist im Frontend zu lang, im Backend gültig. Phase 1 stellt das Frontend auf Zeichen (Unicode-Codepoints, wie `utf8.RuneCountInString`) um; erst dann zählen beide Seiten dieselbe Einheit.

**Event-Format (D22, Freeze-Ausnahme).**

| Event                        | Positionsform (Go)                      | JSON je Position                           |
| ---------------------------- | --------------------------------------- | ------------------------------------------ |
| `bestellung-aufgenommen:v1`  | `kasse.KommentiertePositionEventData`   | Basisform plus optional `kommentar`        |
| `direktverkauf-getaetigt:v2` | `kasse.KommentiertePositionV2EventData` | Form mit `pfand` plus optional `kommentar` |

- `KommentiertePositionEventData` bettet `PositionEventData` ein, `KommentiertePositionV2EventData` bettet `PositionV2EventData` ein. Beide tragen `Kommentar string` mit `json:"kommentar,omitempty"`. Der Schlüssel steht zuletzt in der Position und fehlt, wenn er leer ist.
- `BestellungAufgenommenV1Data.Positionen` und `DirektverkaufGetaetigtV2Data.Positionen` wechseln auf diese Typen. `PositionEventData`, `PositionV2EventData` und `DirektverkaufGetaetigtV1Data` bleiben unberührt.
- Damit erreicht der Kommentar keine Zahlung, Stornierung, Korrektur, Umbuchung und keine Pfandrückgabe: Deren Events nutzen weiter die Basisformen.
- Schema: `kasse.PositionsKommentarSchema = z.String().TestFunc(zeichen.Max(100))`, exportiert und damit in `TestSchemaGrenzen` gepinnt. Event-Schema (Schreiben und Lesen) und Request-Schema nutzen dieselbe Variable.
- Ein Alt-Event liest sich mit leerem Positionskommentar. Nichts wird migriert oder umgedeutet.
- Rollback-sicher: Ein älteres Backend ignoriert den unbekannten Schlüssel beim Dekodieren und Validieren (zog prüft nur die Shape-Felder).
- `docs/decisions.md` bekommt **D22** (Wortlaut in Phase 2). Er nennt das Konzept, keinen Pfad.

**Wer liest den Kommentar.**

- Mit Kommentar: Tisch-Historie (`buildBestellungFromEvent()`), Direktverkauf-Historie (`BuildDirektverkaufHistorieEintrag()`) und die Arbeitsbon-Policy.
- Ohne Kommentar, über eine Umwandlung in die Basisform: Tisch-Projektion (`ApplyEvent()`), Storno-Aufteilung, `FiskalischeProjektion()` (TSE-processData) und der DSFinV-K-Export (`belege.go`).
- Leser, die `DirektverkaufGetaetigtV1Data` dekodieren (Tagesabschluss, Kassenbeleg, Pfand-Verrechnung, Seed), ignorieren den Schlüssel und ändern sich nicht.
- Die Domänen-`kasse.Position` bekommt `Kommentar string`. Gesetzt wird er nur beim Lesen der zwei Erfassungs-Events mit Kommentar und beim Anreichern einer Eingabe.
- `BON_NOTIZ` trägt weiter genau den Bestellkommentar.

**API** (alle POST, Endpunkte unverändert).

| Endpunkt                             | Änderung                                                                            |
| ------------------------------------ | ----------------------------------------------------------------------------------- |
| `service/bestellung-aufnehmen`       | `positionen[]` je optional `kommentar` (≤ 100 Zeichen)                              |
| `service/direktverkauf-taetigen`     | `positionen[]` je optional `kommentar` (≤ 100 Zeichen); `pfand.positionen` ohne     |
| `service/get-tisch-historie`         | `positionen[]` eines Eintrags `art: "bestellung"` tragen `kommentar`, leer als `""` |
| `service/get-direktverkauf-historie` | `positionen[]` tragen `kommentar`, leer als `""`                                    |

- Backend: `eingabe.KommentiertePositionInput` (bettet `eingabe.PositionInput` ein, plus `Kommentar`) mit `eingabe.KommentiertePositionInputSchema`. `eingabe.PositionInput` bleibt für Pfand-Positionen ohne Kommentar.
- `enrichment.PositionInput` bekommt `Kommentar`; `positionAus()` übernimmt ihn in `kasse.Position`. Die Command-Signaturen ändern sich nicht.
- `stornierbarePositionen`, `umbuchbarePositionen`, `offenePositionen`, Tisch-State und Zahlungsauswahl tragen keinen Kommentar.
- Die Pfandrückgabe-Historie teilt das DTO `vorgang.HistorieEintrag` mit dem Direktverkauf. Ihre Positionen tragen deshalb `kommentar: ""`. Ein eigenes DTO wäre mehr Code für keinen Nutzer.
- Das Frontend schickt `kommentar` nur, wenn er nach `trim()` nicht leer ist. Ohne Kommentar bleibt der Request byte-gleich zum heutigen.

**Arbeitsbon und Abholbon.**

- Die Policy dekodiert die Positionen beider Erfassungs-Events als `[]kasse.KommentiertePositionEventData` (der `pfand`-Schlüssel wird ignoriert). Der Kommentar reist in `kasse.Position.Kommentar` mit seiner Position.
- Layout: unter der Positionszeile, normale Schrifthöhe, fett, `wrapCP437` auf `lineWidth − 2`; jede umbrochene Zeile ist zwei Leerzeichen eingerückt. Der Bestellkommentar bleibt am Bonende.
- Ohne Positionskommentar entstehen keine zusätzlichen Steuerbytes: Die Umschaltung auf normale Höhe und zurück steht nur um einen nicht leeren Kommentar. Bons ohne Positionskommentar bleiben byte-gleich, `TestFormatSammelBon_WithKommentar` und `TestCreateArbeitsbonAuftraege_ByteIdentischZumFormatter_ProPosition` bleiben unverändert grün.
- Sammelbon: jede Position mit ihrem Kommentar. Einzelbon (`pro_position`): Position, ihr Kommentar, dann der Bestellkommentar.
- Abholbon `pro_stueck` und `pro_position`: Die kopierte Einheit behält `Kommentar`, also trägt jeder Stückbon den Kommentar seiner Position.
- `FormatSammelBon()`, `FormatPositionBon()` und `FormatDirektverkaufAbholbon()` behalten ihre Signaturen.

**Frontend.**

- Zustand in `useVorgang()` (Befund #221): `positionsKommentare: Partial<Record<K, string>>` und `setPositionsKommentar(key, text)`. Er übersteht Tab-Wechsel und Remount wie der Bestellkommentar.
- Fällt die Menge einer Variante auf null, entfernt der Hook ihren Schlüssel aus dem State, im selben Render-Muster wie beim Wechsel von `leer`. Ein bloßes Filtern beim Lesen reicht nicht: Der Text käme beim erneuten Hinzufügen zurück (User Story 5). Wird die Auswahl leer, leert er alle, zusammen mit Bestellkommentar und `vorgangId`.
- Die Positionen mit Kommentar gehören zum Inhalt von `schluesselFuer()`. Ein Retry mit geändertem Kommentar zieht damit einen frischen Schlüssel.
- Der Warenkorb bleibt je Variante geschlüsselt; `toBestellungData()` legt den Kommentar an die Eingabe- und Belegzeile seiner Variante.
- `Receipt` zeigt einen vorhandenen Kommentar als „…" unter der Zeile. Mit `kommentarBearbeiten` zeigt jede Zeile eine Schaltfläche „Kommentar". Ein Tipp öffnet darunter ein einzeiliges Feld; eine Zeile mit Kommentar zeigt ihr Feld offen.
- Das Feld `PositionsKommentarFeld` liegt in `CommentField.tsx` und teilt die Zeichen-Kappung des `KommentarField`.
- Nur Bestell- und Direktverkauf-Abschluss reichen `kommentarBearbeiten` durch; die Tisch-Zahlung im `KassierAbschluss` nicht.

**Keine Migration.** SQL-Funktionen, Idempotenz-Indexe und Reporting lesen Schlüssel einzeln aus `positionen` und bleiben gleich. Migration 18 geht an Auto-Pfand.

## Inventory

- `backend/domain/kasse/bestellung.go — Position`, `PositionEventData`, `PositionV2EventData`, `positionSchema`, `toPositionenEventData()`, `PositionenFromEventData()` — Positionsformen und Abbildungen.
- `backend/domain/kasse/tisch_session_events.go — BestellungAufgenommenV1Data`, `bestellungAufgenommenV1DataSchema`, `NewBestellungAufgenommenEvent()`, `buildBestellungFromEvent()` — Tisch-Erfassungs-Event.
- `backend/domain/kasse/direktverkauf_events.go — DirektverkaufGetaetigtV2Data`, `v1()`, `NewDirektverkaufGetaetigtEvent()` — Direktverkauf-Erfassungs-Event.
- `backend/domain/kasse/tisch_session.go — ApplyEvent()`; `storno_aufteilung.go — stornoBestand.apply()`; `fiskalische_projektion.go — FiskalischeProjektion()` — Leser der Basisform.
- `backend/domain/kasse/direktverkauf.go — replayVorgang()`, `vorgangKern`, `BuildDirektverkaufHistorieEintrag()`; `pfandrueckgabe.go` — teilt `replayVorgang()`.
- `backend/domain/kasse/event_json_contract_test.go`, `kommentar_test.go — optionalKommentarTests`, `historie_test.go`, `direktverkauf_test.go` — Testvorbilder.
- `backend/api/schema_grenzen_test.go — lengthCases` — pinnt jedes exportierte Domain-Schema.
- `backend/api/event_decode_contract_test.go — TestDecodeTargetsUseEventDataTypes()` — JSON-Ziele unter `backend/api` müssen auf `EventData`/`V<n>Data` enden.
- `backend/api/kasse/eingabe/eingabe.go — PositionInput`, `PositionInputSchema`, `ToPositionInputs()` — Request-Positionen.
- `backend/api/kasse/enrichment/enrichment.go — PositionInput`, `positionAus()` — Anreicherung.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go`, `query_handler.go — position`, `bestellung`, `toBestellung()`.
- `backend/api/kasse/direktverkauf/http/command_handler.go`; `backend/api/kasse/vorgang/vorgang.go — Position`, `HistorieEintrag`, `ToHistorieEintrag()`.
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go — positionenMitKommentarEventData`, `unmarshalPositionenMitKommentar()`, `createAbholbonAuftraege()`, `createStationsAuftraegeFromData()`.
- `backend/api/druck/bondruck/application/escpos/formatter.go — FormatSammelBon()`, `FormatPositionBon()`, `wrapCP437()`.
- `backend/api/fiskal/dsfinvk/belege.go — belegAbleitung.bestellung()`, `belegAbleitung.direktverkauf()` — DSFinV-K-Leser.
- `backend/repository/kassenjournal_repo/tisch_session.go — upsertTischSessionState()` — serialisiert `kasse.Position` ohne DTO.
- `frontend/src/lib/utf8.ts`, `frontend/src/service/CommentField.tsx`, `frontend/src/service/schemas.ts`, `frontend/src/admin/kasse/Kassensitzung.ts` — Byte-Zählung.
- `frontend/src/hooks/use-vorgang.ts — useVorgang()` — Vorgangszustand der Seite.
- `frontend/src/service/drawerUtils.ts — toBestellungData()`, `toReceiptItems()`; `frontend/src/service/Receipt.tsx — Receipt`.
- `frontend/src/service/table/BestellungAbschluss.tsx`, `BestellungTab.tsx`, `Bestellung.ts`, `TischHistorie.tsx`.
- `frontend/src/service/kassieren/KassierAbschluss.tsx`, `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx`, `DirektverkaufTab.tsx`, `Direktverkauf.ts`, `frontend/src/service/VorgangHistorie.tsx`.
- `docs/handbuch.md` §3.6 und §4.5, `docs/language.md`, `docs/decisions.md`, `docs/backlog.md`, `AGENTS.md` „Freeze discipline".

## Open questions / Risks

Entschieden von Nico am 7. November: Der AGENTS.md-Freeze-Satz wird wie in Frage 1 umformuliert. Für alle übrigen Fragen unten gilt die jeweilige Empfehlung.

1. **AGENTS.md-Freeze-Satz.** „A change is a new event version (`:vN`), never an edit in place" wird mit D22 falsch; die Umbuchung mit `benutzerKommentar` war schon ein Präzedenzfall. Phase 2 formuliert: „A change is a new event version (`:vN`). The only edit in place is an additive optional field omitted when empty, approved as a line in `docs/decisions.md`." Empfehlung: so übernehmen. Lehnt Nico ab, entfällt der AGENTS.md-Punkt aus Phase 2, D22 bleibt.
2. **Phase 1 hier oder vorher.** Die Zeichen-Umstellung ist ein eigenständiger Fehler (Frontend strenger als Backend), kein Feature. Empfehlung: als Phase 1 dieses Plans lassen; `fix/q-cross` hat einen festen Befund-Umfang. Will der Lead sie früher auf `main`, landet Phase 1 allein, und die Kriterien bleiben gleich.
3. **Kreuzung mit `fix/q-cross`.** Die Lane ändert Storno- und Umbuchungs-Handler, ihre Tests und Schemas, und druckt TSE-Start und -Ende auf den Kassenbeleg. Gemeinsame Dateien mit diesem Plan:
   - `backend/api/kasse/tischgeschaeft/http/command_handler.go` und `command_handler_test.go` (`mockCommand`, `TestCommandHandler_Abweisungen`),
   - `backend/api/kasse/direktverkauf/http/command_handler.go` und `command_handler_test.go`,
   - `frontend/src/service/direktverkauf/Direktverkauf.ts` (`DirektverkaufStornierenSchema` neben `DirektverkaufTaetigenSchema`),
   - `backend/api/druck/bondruck/application/escpos/formatter.go` und `formatter_test.go` (Kassenbeleg neben Arbeitsbon).
     Die Feature-Spur startet laut Programme erst nach dem Abschluss; der Run-Branch zweigt dann vom gelandeten Stand ab. Läuft `fix/q-cross` noch, rebased die Lane nach dessen Landung. Dieser Plan fasst `DirektverkaufStornieren()`, `StornoDrawer` und die Umbuchung nicht an.
4. **Geschwister-Features.** Auto-Pfand (Abholbon lässt Pfand-Positionen weg, Warenkorb legt Pfand dazu) und Stationsmodell/Zusatzdrucker ändern `arbeitsbon_policy.go`, `toBestellungData()` und `useVorgang()`. Dieser Plan landet zuerst; beide rebasen darauf. Ein Kommentar an einer Pfand-Position verschwindet mit Auto-Pfand vom Abholbon; das ist hinnehmbar. D22 kann mit einer Entscheidungszeile der Geschwister kollidieren; der Lead vergibt Nummern beim Landen wie bei Migrationen. Dieser Plan braucht keine Migration.
5. **Projektion speichert das neue Feld.** Fakt: `upsertTischSessionState()` serialisiert `kasse.Position` ohne DTO. Jede Position in `tisch_sessions.unbezahlte_positionen` bekommt so den Schlüssel `"Kommentar": ""`, wie zuvor `Pfand` und `BestellerName`. Die Projektion ist neu aufbaubar und liest den Kommentar nie. Empfehlung: hinnehmen, kein eigenes Projektions-DTO in diesem Plan.
6. **Abweichungen vom PRD.** Zustand in `useVorgang()` statt im Abschluss (Befund #221). Ein gemeinsames `KommentiertePositionInputSchema` statt zweier Schemas und kein Frontend-Grenztest dafür: Es verdrahtet nur `KommentarSchema`, dessen Grenze `zeichen.test.ts` pinnt; das Feld kappt die Eingabe ohnehin. Die Sätze „Heute zählt Zod Code-Units, zog Bytes" und „Voraussetzung … Zeichen gegen Bytes" im PRD sind veraltet. Empfehlung: Der Lead löscht PRD und Plan beim Landen und schreibt die Changelog-Zeile, wie bei der Pfandrückgabe.
7. **Druckbild.** Die Formatter-Tests pinnen Bytes, nicht das Druckbild. Empfehlung: Positionskommentar auf Sammel-, Einzel- und Stückbon in die manuelle QA von `plan-release-v1.0.0.md` aufnehmen (echter Epson-Drucker).
8. **#145.** Der Melder bietet an zu testen. Antwort und Schließen des Issues sind Außenkommunikation und bleiben bei Nico. Empfehlung: Nico bittet ihn nach dem Release um einen Test am Fest und schließt das Issue danach.

## Phase 1: Kommentare zählen Zeichen im Frontend

**Depends on**: none

### Context

- `frontend/src/lib/utf8.ts — utf8Bytes()`, `kuerzeAufUtf8Bytes()`, `maxUtf8Bytes()` — zählen UTF-8-Bytes.
- `frontend/src/service/schemas.ts — KOMMENTAR_MAX_BYTES`, `KommentarSchema`, `PflichtKommentarSchema`.
- `frontend/src/service/CommentField.tsx — begrenzeEinfuegung()`, `KommentarField` — kappt Einfügungen auf das Byte-Budget.
- `frontend/src/admin/kasse/Kassensitzung.ts — BezeichnungSchema`, `KommentarSchema` — 200 als Bytes gezählt; Backend `kasse/kassensitzung_events.go` zählt `zeichen.Max(200)`.
- `backend/internal/zeichen/zeichen.go — Min()`, `Max()` — Maßstab: `utf8.RuneCountInString`.
- `frontend/src/service/CommentField.test.tsx` — `it.each` „kürzt die Einfügung auf 100 Bytes …".
- `docs/backlog.md` — Punkt „Zeichen gegen Bytes"; der Punkt „Kommentar je Position (#145)" nennt ihn als Voraussetzung.

### What to build

- `frontend/src/lib/zeichen.ts` ersetzt `frontend/src/lib/utf8.ts`: `zeichenAnzahl(text)` zählt Codepoints (`[...text].length`). `kuerzeAufZeichen(text, max)` kürzt an einer Codepoint-Grenze. `maxZeichen(schema, max, message?)` ist die Zod-Verfeinerung.
- `schemas.ts`: `KOMMENTAR_MAX_ZEICHEN = 100`; beide Kommentar-Schemas zählen Zeichen. Der Kommentar nennt die eingefrorenen Event-Schemas als Quelle der 100 Zeichen.
- `CommentField.tsx`: `begrenzeEinfuegung()` rechnet in Zeichen; der Cursor bleibt in UTF-16-Einheiten (`setSelectionRange`). Der Kommentar über `maxLength` nennt Zeichen statt Bytes.
- `Kassensitzung.ts`: beide Schemas über `maxZeichen()`; der Kommentar „Das Backend zählt … Bytes" wird wahr umformuliert.
- Die übrigen Zod-`.max()` (Namen über `createNameSchema()`, Betreiber, Zugangsdaten) zählen weiter UTF-16-Einheiten. Sie sind nur bei Emoji strenger als das Backend und führen nie zu einem 400; Umlaute zählen dort schon heute als ein Zeichen.
- Neuer Test `frontend/src/lib/zeichen.test.ts`, Fall „zählt Umlaute und Emoji als je ein Zeichen": `maxZeichen(z.string(), 100)` nimmt `'ä'.repeat(100)` an, lehnt `'ä'.repeat(101)` ab; `kuerzeAufZeichen('a😀b', 2)` ergibt `'a😀'`.
- `CommentField.test.tsx`: Das `it.each` heißt „kürzt die Einfügung auf 100 Zeichen an einer Zeichengrenze (Fall %#)" und hat genau diese fünf Fälle (vorhanden, Cursor, Einfügung, erwartet, Cursor danach):
  - `''`, 0, `'ä'.repeat(101)`, `'ä'.repeat(100)`, 100
  - `''`, 0, `'a'.repeat(99) + '😀😀'`, `'a'.repeat(99) + '😀'`, 101
  - `'a'.repeat(100)`, 50, `'x'`, `'a'.repeat(100)`, 50
  - `'a'.repeat(98)`, 49, `'xyz'`, `'a'.repeat(49) + 'xy' + 'a'.repeat(49)`, 51
  - `'ä'.repeat(99)`, 0, `'öö'`, `'ö' + 'ä'.repeat(99)`, 1
- `docs/backlog.md`: Punkt „Zeichen gegen Bytes" löschen; im Punkt #145 entfallen der Satz zur Voraussetzung und „Offen: Freigabe des PRD." (Nico hat es freigegeben).

Demo: Ein Bestellkommentar aus 100 Umlauten lässt sich vollständig eingeben und bestellen.

### Acceptance criteria

- [x] Zod-Grenze und Kürzung zählen Codepoints (Fall „zählt Umlaute und Emoji als je ein Zeichen" in `frontend/src/lib/zeichen.test.ts`): `set -o pipefail; cd frontend && pnpm exec vitest run src/lib/zeichen.test.ts -t 'zählt Umlaute und Emoji als je ein Zeichen' --reporter=verbose | grep -- '✓ .*zählt Umlaute und Emoji als je ein Zeichen'`
- [x] Das Kommentarfeld kappt Einfügungen auf 100 Zeichen (fünf Fälle in `frontend/src/service/CommentField.test.tsx`): `cd frontend && out="$(pnpm exec vitest run src/service/CommentField.test.tsx -t 'kürzt die Einfügung auf 100 Zeichen' --reporter=verbose)" && test "$(printf '%s\n' "$out" | grep -c '✓ .*kürzt die Einfügung auf 100 Zeichen')" -eq 5`
- [x] Keine Byte-Zählung bleibt im Frontend: `test ! -e frontend/src/lib/utf8.ts && ! grep -rnE 'utf8Bytes|maxUtf8Bytes|kuerzeAufUtf8Bytes|KOMMENTAR_MAX_BYTES|UTF-8-Bytes' frontend/src`
- [x] Der Backlog-Punkt ist erledigt und entfernt: `! grep -q 'Zeichen gegen Bytes' docs/backlog.md`
- [x] Gate: `make check check-repo check-format`
- [x] Lint: `make lint`
- [x] E2E: `E2E_STACK_PROJECT=jotti-e2e-kommentar E2E_STACK_PORT=8091 make test-e2e`

## Phase 2: Positionskommentar am Tisch

**Depends on**: 1

### Context

- `backend/domain/kasse/bestellung.go — Position`, `PositionEventData`, `positionSchema`, `toPositionenEventData()`, `PositionenFromEventData()`.
- `backend/domain/kasse/tisch_session_events.go — BestellungAufgenommenV1Data`, `bestellungAufgenommenV1DataSchema`, `NewBestellungAufgenommenEvent()`, `buildBestellungFromEvent()`.
- `backend/domain/kasse/tisch_session.go — ApplyEvent()`, `backend/domain/kasse/storno_aufteilung.go — stornoBestand.apply()`, `backend/domain/kasse/fiskalische_projektion.go — FiskalischeProjektion()` — brauchen die Basisform.
- `backend/api/fiskal/dsfinvk/belege.go — belegAbleitung.bestellung()` — `positionen: data.Positionen` braucht die Basisform.
- `backend/domain/kasse/event_json_contract_test.go — TestEventContract_BestellungAufgenommenV1()`, `TestEventContract_BestellungUmgebuchtV1()` (Vorbild `omitempty`), `assertJSONKeyAbsent()`, `positionLiteral`.
- `backend/domain/kasse/kommentar_test.go — optionalKommentarTests`, `kommentarTestPositionen`.
- Tests, die `BestellungAufgenommenV1Data.Positionen` als Basisform nutzen und nach dem Typwechsel nicht mehr kompilieren:
  - Struct-Literale in `backend/domain/kasse/summen_positiv_test.go`, `fiskalische_projektion_test.go`, `backend/api/fiskal/dsfinvk/mapper_test.go`;
  - `backend/domain/kasse/event_json_contract_test.go — TestEventContract_BestellungAufgenommenV1()` übergibt `data.Positionen[0]` an `assertPosition()`;
  - `backend/api/kasse/kassenfuehrung/application/kassenabschluss_wiederanlauf_integration_test.go` ruft `kasse.PositionFromEventData(pos)` je Position; Build-Tag `integration`, nur `make check-integration` und `make lint-backend-full` sehen ihn.
- Frontend-Fixtures mit Bestellungs-Historie: `frontend/src/service/table/TischHistorie.test.tsx`, `TablePage.test.tsx`, `HistorieStornierungDrawer.test.tsx`, `HistorieUmbuchungDrawer.test.tsx`.
- `backend/api/schema_grenzen_test.go — lengthCases`.
- `backend/api/kasse/eingabe/eingabe.go — PositionInput`, `PositionInputSchema`, `ToPositionInputs()`.
- `backend/api/kasse/enrichment/enrichment.go — PositionInput`, `positionAus()`.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go — bestellungAufnehmenRequest` und sein Schema; `command_handler_test.go — mockCommand`, `TestCommandHandler_Abweisungen()`.
- `backend/api/kasse/tischgeschaeft/http/query_handler.go — position`, `bestellung`, `toBestellung()`; `query_handler_test.go — TestPositionResponsesIncludeSteuersatz()` (Vorbild).
- `backend/api/kasse/tischgeschaeft/application/command_test.go — TestBestellungAufnehmen_EnqueueArbeitsbonDruckauftraege()` (Vorbild).
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go — positionenMitKommentarEventData`, `unmarshalPositionenMitKommentar()`, `createStationsAuftraegeFromData()`; `arbeitsbon_policy_test.go`.
- `backend/api/druck/bondruck/application/escpos/formatter.go — FormatSammelBon()`, `FormatPositionBon()`; `formatter_test.go — TestFormatSammelBon_WithKommentar()`.
- `frontend/src/hooks/use-vorgang.ts — useVorgang()`; `use-vorgang.test.ts`.
- `frontend/src/service/schemas.ts — PositionInputSchema`, `KommentarSchema`.
- `frontend/src/service/table/Bestellung.ts — BestellungAufnehmenSchema`, `PositionSchema`, `BestellungSchema`.
- `frontend/src/service/drawerUtils.ts — toBestellungData()`, `toReceiptItems()`.
- `frontend/src/service/Receipt.tsx — ReceiptPosition`, `Receipt`; `frontend/src/service/CommentField.tsx`.
- `frontend/src/service/table/BestellungTab.tsx`, `BestellungAbschluss.tsx`, `BestellungAbschluss.test.tsx`, `TischHistorie.tsx`, `TischHistorie.test.tsx`.
- `docs/decisions.md`, `docs/language.md` (Zeilen „Kommentar", „BenutzerKommentar"), `docs/handbuch.md` §3.6 Tisch-Session-Events und §4.5 Arbeitsbon, `AGENTS.md` „Freeze discipline".

### What to build

Backend, Domäne:

- `kasse.Position` bekommt `Kommentar string` mit einem Satz Doku: gesetzt nur aus den zwei Erfassungs-Events und aus der Eingabe.
- `bestellung.go`: `KommentiertePositionEventData` und `KommentiertePositionV2EventData` wie in „Architectural decisions". `PositionsKommentarSchema` exportiert. `kommentiertePositionSchema` ist die Shape von `positionSchema` plus `"Kommentar": PositionsKommentarSchema`.
- Abbildungen: `toKommentiertePositionenEventData()` (schreibt `Kommentar`), `PositionenMitKommentar()` (liest ihn), `PositionenOhneKommentar()` (Basisform für Leser ohne Kommentar).
- `BestellungAufgenommenV1Data.Positionen` wird `[]KommentiertePositionEventData`. Das Schema prüft die Positionen mit `kommentiertePositionSchema`. `NewBestellungAufgenommenEvent()` schreibt den Kommentar; `buildBestellungFromEvent()` liest ihn mit.
- `ApplyEvent()`, `stornoBestand.apply()`, `FiskalischeProjektion()` und `belegAbleitung.bestellung()` lesen über `PositionenOhneKommentar()`.
- Die Tests aus der Context-Liste wechseln nur den Positionstyp oder greifen über `.PositionEventData` auf die Basisform zu; erwartete Werte bleiben.
- Event-Vertrag, neuer Test `TestEventContract_BestellungAufgenommenV1_Positionskommentar`:
  - Ein Alt-Event (`positionLiteral` ohne `kommentar`) dekodiert mit `Kommentar == ""` und besteht `bestellungAufgenommenV1DataSchema`.
  - `json.Marshal` dieses Events ist byte-gleich zu einem eingefrorenen kompakten Literal im Bestandsformat.
  - Ein Literal mit `"kommentar": "ohne Kümmel"` an der Position dekodiert den Text, und der Schlüssel steht nach `json.Marshal` in der Position.
- `kommentar_test.go`: `TestNewBestellungAufgenommenEvent_Positionskommentar` läuft `optionalKommentarTests` über den Positionskommentar.
- `historie_test.go`: `TestGetHistorieFromEvents_Positionskommentar` — eine Bestellung mit zwei Positionen, nur die erste kommentiert; `Positionen[0].Kommentar == "ohne Kümmel"`, `Positionen[1].Kommentar == ""`.
- `schema_grenzen_test.go`: Zeile `kasse.PositionsKommentarSchema`, shortest 0, longest 100, filler `'ä'`.

Backend, API und Druck:

- `eingabe.go`: `KommentiertePositionInput`, `KommentiertePositionInputSchema` (Felder von `PositionInputSchema` plus `"Kommentar": kasse.PositionsKommentarSchema`), `ToKommentiertePositionInputs()`.
- `enrichment.PositionInput.Kommentar`; `positionAus()` setzt `kasse.Position.Kommentar`.
- `bestellungAufnehmenRequest.Positionen` wird `[]eingabe.KommentiertePositionInput`.
- `command_handler_test.go`: `mockCommand` merkt sich die Positionen von `BestellungAufnehmen`. Neuer Test `TestBestellungAufnehmenHandler_ReichtPositionskommentarDurch`. `TestCommandHandler_Abweisungen` bekommt den Fall „aufnehmen: Positionskommentar über 100 Zeichen" (101 × `ä`, 400 `validation_error`).
- `query_handler.go`: DTO `bestellPosition` (bettet `position` ein, plus `Kommentar` mit `json:"kommentar"`). `bestellung.Positionen` nutzt es; `stornierbarePositionen` und `umbuchbarePositionen` bleiben `position`.
- `query_handler_test.go`: `TestBestellungResponseTraegtPositionskommentar` — `toBestellung()` liefert `"kommentar":"ohne Kümmel"` und `"kommentar":""` je Position.
- `application/command_test.go`: `TestBestellungAufnehmen_SchreibtPositionskommentar` — das geschriebene Event trägt `kommentar` nur an der kommentierten Position.
- `arbeitsbon_policy.go`: `positionenMitKommentarEventData.Positionen` wird `[]kasse.KommentiertePositionEventData`; `unmarshalPositionenMitKommentar()` nutzt `PositionenMitKommentar()`.
- `formatter.go`: `FormatSammelBon()` und `FormatPositionBon()` schreiben `pos.Kommentar` nach dem Layout aus „Architectural decisions".
- `formatter_test.go`: `TestFormatSammelBon_PositionskommentarUnterSeinerPosition` und `TestFormatPositionBon_PositionskommentarVorBestellkommentar` vergleichen gegen Byte-Literale.
- `arbeitsbon_policy_test.go`:
  - `TestCreateArbeitsbonAuftraege_PositionskommentarNurAnSeinerStation` ist die Regression aus #145: Essen mit Kommentar, Getränk ohne, zwei Stationen `pro_bestellung`; nur der Essenbon enthält den Text.
  - Die Policy-Tests nutzen ASCII-Kommentare wie „ohne Speck". Der Payload ist CP437, ein Umlaut-Text ließe die Negativprüfung auch ohne Umsetzung bestehen.
  - `TestCreateArbeitsbonAuftraege_ProPosition_PositionskommentarNurAufSeinemEinzelbon`: Station `pro_position`, zwei Positionen, eine kommentiert.

Frontend:

- `use-vorgang.ts`: `positionsKommentare`, `setPositionsKommentar()`, Verwerfen bei Menge null, Leeren mit der Auswahl.
- `use-vorgang.test.ts`: Fall „verwirft den Positionskommentar einer Variante, deren Menge auf null fällt": Variante 1 und 2 hinzufügen, 1 kommentieren, 1 entfernen, 1 wieder hinzufügen; der Kommentar von 1 ist leer. Fall „leert die Positionskommentare mit der Auswahl".
- `schemas.ts`: `KommentiertePositionInputSchema = PositionInputSchema.extend({ kommentar: KommentarSchema.optional() })`.
- `Bestellung.ts`: `BestellungAufnehmenSchema.positionen` nutzt `KommentiertePositionInputSchema`; `BestellungSchema.positionen` nutzt `PositionSchema.extend({ kommentar: z.string() })`. Die Fixtures aus der Context-Liste bekommen je Bestellposition `kommentar: ''`.
- `drawerUtils.ts`: `toBestellungData(products, mengen, positionsKommentare?)` legt den getrimmten, nicht leeren Kommentar an Eingabe- und Belegzeile. `toReceiptItems()` übernimmt `kommentar`, falls die Position einen trägt.
- `Receipt.tsx`: `ReceiptPosition.kommentar?`; Anzeige „…" unter der Zeile; optional `kommentarBearbeiten?: (id: string, text: string) => void` mit Schaltfläche „Kommentar" und `PositionsKommentarFeld` je Zeile.
- `CommentField.tsx`: `PositionsKommentarFeld` (einzeiliges Input, Platzhalter „Kommentar zur Position", gleiche Zeichen-Kappung).
- `BestellungTab.tsx` reicht `vorgang.positionsKommentare` an `toBestellungData()`. `BestellungAbschluss.tsx` setzt `kommentarBearbeiten` auf `setPositionsKommentar(Number(id), text)`.
- `BestellungAbschluss.test.tsx`: Fälle „sendet den Positionskommentar nur an seiner Position" und „erneuert den bestellungId, wenn sich ein Positionskommentar nach einem Fehlversuch ändert".
- `TischHistorie.test.tsx`: Fall „zeigt den Positionskommentar unter seiner Position im Detail".

Doku:

- `docs/decisions.md`, neue Zeile:
  „- **D22 Positionskommentar als additives Feld der Erfassungs-Events.** Die Positionen von `bestellung-aufgenommen:v1` und `direktverkauf-getaetigt:v2` tragen optional `kommentar`, ohne neue Event-Version. Das ist eine ausdrückliche Ausnahme vom Freeze der Event-Formate. Leer fehlt der Schlüssel, Alt-Events bleiben byte-identisch und lesen sich mit leerem Kommentar. Eine neue Version berührte jedes SQL-Literal beider Events und beide Idempotenz-Indexe, für ein rein operatives Feld. Der Kommentar erreicht weder Zahlung, Stornierung, Korrektur und Umbuchung noch TSE-Daten und DSFinV-K-Export. Jedes weitere Feld in einem bestehenden Event braucht eine eigene Entscheidung."
- `AGENTS.md` „Freeze discipline": Satz wie in Open question 1.
- `docs/language.md`, Tabelle „Weitere Typen und Felder (Kasse)" (Spalten Begriff, Bedeutung, Code-Mapping), zwei Zeilen nach dem PRD:
  - Positionskommentar: optionaler Freitext der Servicekraft zu einer Position einer Bestellung oder eines Direktverkaufs, gedruckt nur an deren Station; zu vermeiden Positionsnotiz, Notiz, Hinweis, Sonderwunsch. Code-Mapping: Go `Position.Kommentar` · JSON/TS `kommentar` im Positionsobjekt.
  - Bestellkommentar: Kommentar einer Bestellung, gedruckt auf jedem ihrer Arbeitsbons; zu vermeiden Bestellnotiz, Bonkommentar. Code-Mapping: Go `Bestellung.Kommentar` · JSON/TS `kommentar` auf Event-Ebene.
- `docs/handbuch.md`: In §3.6 trägt die Zeile `bestellung-aufgenommen:v1` „optional `kommentar` je Position (D22)". §4.5 Arbeitsbon-Inhalt: „Positionen (Art + Menge) mit ihrem Positionskommentar, Bestellkommentar am Ende".

Demo: Die Servicekraft bestellt Zwiebelkuchen mit „ohne Kümmel" und ein Bier ohne Kommentar. Der Essenbon zeigt den Kommentar unter dem Zwiebelkuchen, der Getränkebon keinen. Das Detail der Bestellung in der Historie zeigt „ohne Kümmel" unter der Position.

### Acceptance criteria

- [x] Event-Vertrag: Alt-Event byte-identisch, mit Kommentar lesbar, Schema nimmt das Alt-Event an (`TestEventContract_BestellungAufgenommenV1_Positionskommentar`, alle übrigen `TestEventContract_*` grün): `set -o pipefail; cd backend && go test ./domain/kasse/ -run '^TestEventContract_' -v | grep -- '--- PASS: TestEventContract_BestellungAufgenommenV1_Positionskommentar '`
- [x] Der Konstruktor begrenzt den Positionskommentar auf 100 Zeichen, und die Historie liest ihn je Position (`TestNewBestellungAufgenommenEvent_Positionskommentar`, `TestGetHistorieFromEvents_Positionskommentar`): `cd backend && out="$(go test ./domain/kasse/ -run '^(TestNewBestellungAufgenommenEvent_Positionskommentar|TestGetHistorieFromEvents_Positionskommentar)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '^--- PASS: (TestNewBestellungAufgenommenEvent_Positionskommentar|TestGetHistorieFromEvents_Positionskommentar) ')" -eq 2`
- [x] `kasse.PositionsKommentarSchema` ist mit 0 bis 100 Zeichen gepinnt: `cd backend && grep -q '"kasse.PositionsKommentarSchema"' api/schema_grenzen_test.go && go test ./api/ -run '^TestSchemaGrenzen$'`
- [x] Der Request reicht den Kommentar durch und lehnt 101 Zeichen ab (`TestBestellungAufnehmenHandler_ReichtPositionskommentarDurch`, Fall „aufnehmen: Positionskommentar über 100 Zeichen"): `cd backend && out="$(go test ./api/kasse/tischgeschaeft/http/ -run '^(TestBestellungAufnehmenHandler_ReichtPositionskommentarDurch|TestCommandHandler_Abweisungen)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '--- PASS: (TestBestellungAufnehmenHandler_ReichtPositionskommentarDurch|TestCommandHandler_Abweisungen/aufnehmen:_Positionskommentar_über_100_Zeichen) ')" -eq 2`
- [x] Das Historien-DTO trägt den Kommentar je Bestellposition (`TestBestellungResponseTraegtPositionskommentar`): `set -o pipefail; cd backend && go test ./api/kasse/tischgeschaeft/http/ -run '^TestBestellungResponseTraegtPositionskommentar$' -v | grep -- '--- PASS: TestBestellungResponseTraegtPositionskommentar '`
- [x] Der Command schreibt den Kommentar an seine Position (`TestBestellungAufnehmen_SchreibtPositionskommentar`): `set -o pipefail; cd backend && go test ./api/kasse/tischgeschaeft/application/ -run '^TestBestellungAufnehmen_SchreibtPositionskommentar$' -v | grep -- '--- PASS: TestBestellungAufnehmen_SchreibtPositionskommentar '`
- [x] Der Arbeitsbon trägt den Kommentar nur an der Station und auf dem Einzelbon seiner Position (Regression #145): `cd backend && out="$(go test ./api/druck/bondruck/application/ -run '^(TestCreateArbeitsbonAuftraege_PositionskommentarNurAnSeinerStation|TestCreateArbeitsbonAuftraege_ProPosition_PositionskommentarNurAufSeinemEinzelbon)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '^--- PASS: TestCreateArbeitsbonAuftraege_(PositionskommentarNurAnSeinerStation|ProPosition_PositionskommentarNurAufSeinemEinzelbon) ')" -eq 2`
- [x] Bonlayout: Positionskommentar unter seiner Zeile, Bestellkommentar am Ende: `cd backend && out="$(go test ./api/druck/bondruck/application/escpos/ -run '^(TestFormatSammelBon_PositionskommentarUnterSeinerPosition|TestFormatPositionBon_PositionskommentarVorBestellkommentar)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '^--- PASS: TestFormat(SammelBon_PositionskommentarUnterSeinerPosition|PositionBon_PositionskommentarVorBestellkommentar) ')" -eq 2`
- [x] Der Hook verwirft Kommentare bei Menge null und mit der Auswahl: `cd frontend && out="$(pnpm exec vitest run src/hooks/use-vorgang.test.ts --reporter=verbose)" && test "$(printf '%s\n' "$out" | grep -cE '✓ .*(verwirft den Positionskommentar einer Variante, deren Menge auf null fällt|leert die Positionskommentare mit der Auswahl)')" -eq 2`
- [x] Der Bestellabschluss sendet den Kommentar nur an seiner Position und zieht nach Änderung einen neuen Schlüssel: `cd frontend && out="$(pnpm exec vitest run src/service/table/BestellungAbschluss.test.tsx --reporter=verbose)" && test "$(printf '%s\n' "$out" | grep -cE '✓ .*(sendet den Positionskommentar nur an seiner Position|erneuert den bestellungId, wenn sich ein Positionskommentar nach einem Fehlversuch ändert)')" -eq 2`
- [x] Die Historie zeigt den Kommentar unter seiner Position: `set -o pipefail; cd frontend && pnpm exec vitest run src/service/table/TischHistorie.test.tsx -t 'zeigt den Positionskommentar unter seiner Position im Detail' --reporter=verbose | grep -- '✓ .*zeigt den Positionskommentar unter seiner Position im Detail'`
- [x] D22 steht, Handbuch und Glossar nennen den Positionskommentar, AGENTS.md nennt die Ausnahme: `grep -q '^- \*\*D22 Positionskommentar' docs/decisions.md && grep -q '^| Positionskommentar' docs/language.md && grep -q 'Positionskommentar' docs/handbuch.md && grep -q 'additive optional field' AGENTS.md`
- [x] Gate: `make check check-repo check-format`
- [x] Lint: `make lint`
- [x] Integration: `TEST_PG_CONTAINER=jotti-pg-kommentar TEST_PG_PORT=5447 make check-integration`
- [x] E2E: `E2E_STACK_PROJECT=jotti-e2e-kommentar E2E_STACK_PORT=8091 make test-e2e`

## Phase 3: Positionskommentar im Direktverkauf

**Depends on**: 2

### Context

- `backend/domain/kasse/direktverkauf_events.go — DirektverkaufGetaetigtV2Data`, `v1()`, `NewDirektverkaufGetaetigtEvent()`.
- `backend/domain/kasse/bestellung.go — KommentiertePositionV2EventData`, `kommentiertePositionSchema`, `PositionenOhneKommentar()` aus Phase 2.
- `backend/domain/kasse/direktverkauf.go — replayVorgang()`, `vorgangKern`, `BuildDirektverkaufHistorieEintrag()`; `backend/domain/kasse/pfandrueckgabe.go` — zweiter Nutzer von `replayVorgang()`.
- `backend/domain/kasse/direktverkauf_test.go — TestBuildDirektverkaufHistorieEintrag_PartialThenFullStorno()` (Vorbild).
- `backend/domain/kasse/event_json_contract_test.go — TestEventContract_DirektverkaufGetaetigtV2()`, `positionV2Literal`, `assertPositionV2()` — übergibt `data.Positionen` als `[]PositionV2EventData`.
- `backend/seed/engine_test.go` — übergibt `DirektverkaufGetaetigtV2Data.Positionen` an `pruefe(…, []kasse.PositionV2EventData)`.
- `backend/api/fiskal/dsfinvk/belege.go — belegAbleitung.direktverkauf()`, `verkaufsPositionen()`, `pfandJePosition()`; `mapper_test.go` (Fixture mit `PositionV2EventData`).
- `backend/api/kasse/direktverkauf/http/command_handler.go — direktverkaufTaetigenRequest` und Schema; `command_handler_test.go — mockCommand`, `TestCommandHandler_Abweisungen()`.
- `backend/api/kasse/direktverkauf/application/command_test.go — TestDirektverkaufTaetigen_WritesSingleEvent()` (Vorbild).
- `backend/api/kasse/vorgang/vorgang.go — Position`, `HistorieEintrag`, `ToHistorieEintrag()`.
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go — createAbholbonAuftraege()`; `arbeitsbon_policy_test.go — TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck()` (Vorbild).
- `frontend/src/service/direktverkauf/Direktverkauf.ts — DirektverkaufTaetigenSchema`; `frontend/src/service/schemas.ts — VorgangHistorieBasisSchema`.
- `frontend/src/service/kassieren/KassierAbschluss.tsx — KassierAbschlussProps`; `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx`, `DirektverkaufTab.tsx`, `DirektverkaufAbschluss.test.tsx`.
- `frontend/src/service/VorgangHistorie.tsx`, `VorgangHistorie.test.tsx`.
- Frontend-Fixtures mit Vorgangs-Historie: `frontend/src/service/direktverkauf/DirektverkaufPage.test.tsx`, `frontend/src/service/pfand/PfandrueckgabenAbschnitt.test.tsx`.
- `docs/handbuch.md` §3.6 Direktverkauf-Events und §4.5 Direktverkauf-Routing; `docs/backlog.md` Punkt #145.

### What to build

Backend:

- `DirektverkaufGetaetigtV2Data.Positionen` wird `[]KommentiertePositionV2EventData`.
- Neues `direktverkaufGetaetigtV2DataSchema`: wie das v1-Schema, die Positionen mit `kommentiertePositionSchema`. zog findet die eingebetteten Felder über `reflect.Type.FieldByName`. `NewDirektverkaufGetaetigtEvent()` validiert damit die v2-Daten direkt und schreibt den Kommentar. `DirektverkaufGetaetigtV2Data.v1()` hat danach keinen Nutzer und entfällt.
- Abbildungen in `bestellung.go`: `toKommentiertePositionenV2EventData()` (schreibt Pfand und Kommentar), `PositionenV2MitKommentar()` (liest den Kommentar in `kasse.Position`), `PositionenV2OhneKommentar()` (Basisform mit `pfand`).
- `belegAbleitung.direktverkauf()` reicht über `PositionenV2OhneKommentar()` an `verkaufsPositionen()` und `pfandJePosition()`.
- `vorgangKern.Positionen` wird `[]Position`; jeder Aufrufer bildet selbst ab. `BuildDirektverkaufHistorieEintrag()` dekodiert v1 und v2 als `DirektverkaufGetaetigtV2Data` und bildet über `PositionenV2MitKommentar()` ab. Die Pfandrückgabe-Closure in `pfandrueckgabe.go` ruft `PositionenFromEventData()`; das ist ihre einzige Änderung.
- `mapper_test.go`, `event_json_contract_test.go — TestEventContract_DirektverkaufGetaetigtV2()` und `backend/seed/engine_test.go` wechseln nur den Positionstyp oder greifen über `.PositionV2EventData` zu; erwartete Werte bleiben.
- Event-Vertrag, neuer Test `TestEventContract_DirektverkaufGetaetigtV2_Positionskommentar`:
  - Alt-Event ohne Schlüssel liest sich mit `""`, und `json.Marshal` ist byte-gleich zum eingefrorenen Literal.
  - Mit Schlüssel liest es den Text, und `DirektverkaufGetaetigtV1Data` dekodiert dasselbe Literal weiter.
- `kommentar_test.go`: `TestNewDirektverkaufGetaetigtEvent_Positionskommentar` über `optionalKommentarTests`.
- `direktverkauf_test.go`: `TestBuildDirektverkaufHistorieEintrag_Positionskommentar` — v2 mit Kommentar liest ihn; ein v1-Stream liest `""`.
- `vorgang.go`: DTO `KommentiertePosition` (bettet `Position` ein, plus `Kommentar` mit `json:"kommentar"`). `HistorieEintrag.Positionen` nutzt es; `OffenePositionen` bleibt `Position`.
- Neuer Test `backend/api/kasse/vorgang/vorgang_test.go — TestToHistorieEintrag_Positionskommentar`: `"kommentar":"ohne Senf"` und `"kommentar":""` je Position.
- `direktverkaufTaetigenRequest.Positionen` wird `[]eingabe.KommentiertePositionInput`; `pfand.positionen` bleibt `eingabe.PositionInput`.
- `direktverkauf/http/command_handler_test.go`: `mockCommand` merkt sich die Positionen von `DirektverkaufTaetigen`. `TestDirektverkaufTaetigenHandler_ReichtPositionskommentarDurch` und der Fall „tätigen: Positionskommentar über 100 Zeichen" in `TestCommandHandler_Abweisungen`.
- `direktverkauf/application/command_test.go`: `TestDirektverkaufTaetigen_SchreibtPositionskommentar` — das v2-Event trägt `kommentar` nur an der kommentierten Position.
- `arbeitsbon_policy_test.go`: `TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck_PositionskommentarAufJedemStueckbon` — 2 × Bratwurst mit „1 davon ohne Senf", 1 × Bier ohne; beide Bratwurst-Stückbons tragen den Text, der Bier-Stückbon nicht.

Frontend:

- `Direktverkauf.ts`: `DirektverkaufTaetigenSchema.positionen` nutzt `KommentiertePositionInputSchema`.
- `schemas.ts`: `VorgangHistorieBasisSchema.positionen` nutzt `GebuchtePositionSchema.extend({ kommentar: z.string() })`. Die Fixtures aus der Context-Liste und in `VorgangHistorie.test.tsx` bekommen je Position `kommentar: ''`.
- `KassierAbschluss.tsx`: optionale Prop `kommentarBearbeiten`, an `Receipt` durchgereicht.
- `DirektverkaufAbschluss.tsx` setzt sie auf `setPositionsKommentar(Number(id), text)`; `DirektverkaufTab.tsx` reicht `vorgang.positionsKommentare` an `toBestellungData()`. Die Tisch-Zahlung reicht nichts durch.
- `DirektverkaufAbschluss.test.tsx`: Fall „sendet den Positionskommentar nur an seiner Position".
- `VorgangHistorie.test.tsx`: Fall „zeigt den Positionskommentar unter seiner Position im Detail".

Doku:

- `docs/handbuch.md`: In §3.6 trägt die Zeile `direktverkauf-getaetigt:v2` „optional `kommentar` je Position (D22)". Im Direktverkauf-Routing von §4.5 steht: „Jeder Abholbon trägt den Positionskommentar seiner Position, `pro_stueck` auf jedem Stückbon."
- `docs/backlog.md`: Punkt „Kommentar je Position (#145)" löschen.

Demo: Im Direktverkauf bekommt „2 × Bratwurst" den Kommentar „1 davon ohne Senf". Mit Abholbon `pro_stueck` tragen beide Bratwurst-Bons den Text, der Bier-Bon nicht. Die Direktverkauf-Historie zeigt ihn unter der Position. Der Kassenbeleg des Verkaufs zeigt ihn nicht.

### Acceptance criteria

- [x] Event-Vertrag v2: Alt-Event byte-identisch, mit Kommentar lesbar, v1-Leser unberührt: `set -o pipefail; cd backend && go test ./domain/kasse/ -run '^TestEventContract_' -v | grep -- '--- PASS: TestEventContract_DirektverkaufGetaetigtV2_Positionskommentar '`
- [x] Konstruktor und Historie behandeln den Positionskommentar: `cd backend && out="$(go test ./domain/kasse/ -run '^(TestNewDirektverkaufGetaetigtEvent_Positionskommentar|TestBuildDirektverkaufHistorieEintrag_Positionskommentar)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '^--- PASS: (TestNewDirektverkaufGetaetigtEvent_Positionskommentar|TestBuildDirektverkaufHistorieEintrag_Positionskommentar) ')" -eq 2`
- [x] Das Vorgangs-DTO trägt den Kommentar je Position: `set -o pipefail; cd backend && go test ./api/kasse/vorgang/ -run '^TestToHistorieEintrag_Positionskommentar$' -v | grep -- '--- PASS: TestToHistorieEintrag_Positionskommentar '`
- [x] Der Request reicht den Kommentar durch und lehnt 101 Zeichen ab: `cd backend && out="$(go test ./api/kasse/direktverkauf/http/ -run '^(TestDirektverkaufTaetigenHandler_ReichtPositionskommentarDurch|TestCommandHandler_Abweisungen)$' -v)" && test "$(printf '%s\n' "$out" | grep -cE -- '--- PASS: (TestDirektverkaufTaetigenHandler_ReichtPositionskommentarDurch|TestCommandHandler_Abweisungen/tätigen:_Positionskommentar_über_100_Zeichen) ')" -eq 2`
- [x] Der Command schreibt den Kommentar an seine Position: `set -o pipefail; cd backend && go test ./api/kasse/direktverkauf/application/ -run '^TestDirektverkaufTaetigen_SchreibtPositionskommentar$' -v | grep -- '--- PASS: TestDirektverkaufTaetigen_SchreibtPositionskommentar '`
- [x] Jeder Stückbon trägt den Kommentar seiner Position: `set -o pipefail; cd backend && go test ./api/druck/bondruck/application/ -run '^TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck_PositionskommentarAufJedemStueckbon$' -v | grep -- '--- PASS: TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck_PositionskommentarAufJedemStueckbon '`
- [x] Der Direktverkauf-Abschluss sendet den Kommentar nur an seiner Position: `set -o pipefail; cd frontend && pnpm exec vitest run src/service/direktverkauf/DirektverkaufAbschluss.test.tsx -t 'sendet den Positionskommentar nur an seiner Position' --reporter=verbose | grep -- '✓ .*sendet den Positionskommentar nur an seiner Position'`
- [x] Die Direktverkauf-Historie zeigt den Kommentar unter seiner Position: `set -o pipefail; cd frontend && pnpm exec vitest run src/service/VorgangHistorie.test.tsx -t 'zeigt den Positionskommentar unter seiner Position im Detail' --reporter=verbose | grep -- '✓ .*zeigt den Positionskommentar unter seiner Position im Detail'`
- [x] Handbuch nennt den Kommentar am v2-Event und am Abholbon; der Backlog-Punkt #145 ist weg: `grep -q 'direktverkauf-getaetigt:v2.*kommentar' docs/handbuch.md && grep -q 'Stückbon' docs/handbuch.md && ! grep -q '#145' docs/backlog.md`
- [x] Gate: `make check check-repo check-format`
- [x] Lint: `make lint`
- [x] Integration: `TEST_PG_CONTAINER=jotti-pg-kommentar TEST_PG_PORT=5447 make check-integration`
- [x] E2E: `E2E_STACK_PROJECT=jotti-e2e-kommentar E2E_STACK_PORT=8091 make test-e2e`
