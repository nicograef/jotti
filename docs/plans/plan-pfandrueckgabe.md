# Plan: Pfandrückgabe und Verrechnung

> Source PRD: docs/prds/prd-pfandrueckgabe.md

## Goal

Ein Gast gibt Pfandgut zurück und erhält den Pfandbetrag bar oder verrechnet. Jede Rückgabe ist ein eigener, TSE-signierter Kassenvorgang. Sie mindert den Soll-Kassenbestand und den Umsatz am Steuersatz des Pfand-Produkts. Der DSFinV-K-Export führt sie als „PfandRueckzahlung". Die Serviceleitung storniert eine irrtümliche Rückgabe positionsgenau. Bestehende Events, Schemas und die Positiv-Regel der Geldsummen bleiben unberührt.

## Architectural decisions

**Stammdaten.**

- `produkte.pfand BOOLEAN NOT NULL DEFAULT false`. Das Domain-Feld heißt `produkt.Produkt.Pfand`.
- Das Kennzeichen steuert nur, welche Varianten die Rückgabe anbietet. Bestellung, Arbeitsbon und Kategorie bleiben unverändert.
- Der Steuersatz des Pfand-Produkts ist der Satz der Rückgabe. Die fette Position friert ihn mit Name und Preis ein.
- `service/get-aktive-produkte` liefert zusätzlich `pfandProdukte`: die Teilmenge mit `pfand = true`, gefiltert im Backend.

**Event-Verträge** (neue Typen, bestehende bleiben unverändert):

| Event                         | JSON-Felder                                                                                     | Regeln                                                                                      |
| ----------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `pfand-zurueckgegeben:v1`     | `rueckgabeId`, `tischId`, `positionen`, `gesamtbetragCents`, `kommentar`                         | `rueckgabeId` vom Client (UUID); `tischId` 0 = Direktverkauf; ≥ 1 Position; Summe ≥ 1; Kommentar ≤ 100, optional |
| `pfandrueckgabe-storniert:v1` | `stornierungId`, `rueckgabeId`, `positionen`, `gesamtStornierungCents`, `kommentar`              | `stornierungId` vom Client (UUID); ≥ 1 Position; Summe ≥ 1; Kommentar 3–100, Pflicht        |

- Go: `kasse.EventTypePfandZurueckgegebenV1`, `kasse.PfandZurueckgegebenV1Data`, `kasse.EventTypePfandrueckgabeStorniertV1`, `kasse.PfandrueckgabeStorniertV1Data`.
- Positionen sind `kasse.PositionEventData`, positiv, mit frischen `positionId`s. Das Vorzeichen „Auszahlung" trägt der Event-Typ.
- Die Stornierung nimmt je Position höchstens die zurückgegebene, noch nicht stornierte Menge zurück.

**Stream.** Jede Pfandrückgabe bekommt einen eigenen Stream `kassensitzung-{nr}/pfandrueckgabe-{rueckgabeId}`.

- `kasse.StreamTypePfandrueckgabe = "pfandrueckgabe"`, ohne Projektion. `kasse.PfandrueckgabeSubject()` baut das Subject.
- Die Rückgabe ist `version = 1`, jede Stornierung eine Folge-Version im selben Stream.
- `tischId` im Event ordnet die Rückgabe dem Tisch zu, mit und ohne Tisch-Session. Die Tisch-Projektion bleibt unberührt.
- Eine Rückgabe erzeugt kein Guthaben am Tisch. Verrechnet wird nur im selben Abschluss.

**Migration** `database/migrations/11_pfandrueckgabe.up.sql`, in `BEGIN; … COMMIT;`:

- `ALTER TABLE produkte ADD COLUMN pfand BOOLEAN NOT NULL DEFAULT false` samt `COMMENT`.
- `idx_kassenjournal_rueckgabe_id`: partieller UNIQUE-Index auf `(data->>'rueckgabeId')` für `pfand-zurueckgegeben:v1`.
- `idx_kassenjournal_pfand_stornierung_id`: partieller UNIQUE-Index auf `(data->>'stornierungId')` für `pfandrueckgabe-storniert:v1`.
- `kj_extract_pfand_cents(type, data)`: vorzeichenbehaftete Bargeldwirkung. Rückgabe `-gesamtbetragCents`, Stornierung `+gesamtStornierungCents`, sonst NULL.
- `CREATE OR REPLACE FUNCTION kj_extract_umsatz_pro_steuersatz`: zusätzlich beide Pfand-Typen. Rückgabe Faktor −1, Stornierung +1. Alte Typen liefern unverändert dasselbe.

**Endpunkte** (alle POST):

| Endpunkt                                   | Body                                                                             | Rolle                 |
| ------------------------------------------ | -------------------------------------------------------------------------------- | --------------------- |
| `service/pfand-zurueckgeben`               | `{rueckgabeId, tischId, positionen: [{produktId, varianteId, menge}], kommentar}` | alle                  |
| `service/zahlung-kassieren`                | bestehend, plus optional `pfand: {rueckgabeId, positionen}`                       | alle                  |
| `service/direktverkauf-taetigen`           | bestehend, plus optional `pfand: {rueckgabeId, positionen}`                       | alle                  |
| `serviceleitung/pfandrueckgabe-stornieren` | `{stornierungId, rueckgabeId, positionen: [{positionId, menge}], kommentar}`      | Serviceleitung, Admin |
| `service/get-pfandrueckgabe-historie`      | `{tischId}`; 0 = Direktverkauf                                                    | alle                  |
| `service/beleg-drucken`                    | zusätzlich `{rueckgabeId}` und `{rueckgabeId, stornierungId}`                      | alle                  |

- `admin/create-produkt` und `admin/update-produkt` nehmen `pfand` an. Die Produkt-DTOs beider Lesepfade tragen `pfand`.
- Pfand-Positionen müssen zu einem aktiven Pfand-Produkt gehören. Sonst antwortet der Endpunkt 400 `kein_pfand_produkt`, und nichts wird gebucht.
- Die Rückgabe-ID macht jeden Command mit Pfand idempotent. Ein zweiter Aufruf mit derselben ID antwortet Erfolg und bucht nichts.
- Ein verrechneter Abschluss schreibt Zahlung oder Verkauf und Rückgabe in einer Transaktion. Scheitert eins, fehlen beide.
- Bei einer reinen Rückgabe füllt das bestehende `KommentarField` des Abschlusses ihren `kommentar`.
- Der Kommentar eines verrechneten Abschlusses gehört zur Zahlung oder zum Verkauf; die Rückgabe trägt dann keinen.
- Die Abschluss-Barriere der Kassensitzung gilt: `kasse.IsAbschlussEventType()` meldet beide Typen als `false`.

**TSE.** Jede Rückgabe und jede Stornierung ist ein eigener Geschäftsvorfall mit eigener TSE-Transaktion (AEAO zu § 146a Nr. 1.10.1, § 2 Satz 1 KassenSichV).

- Rückgabe: `Kassenbeleg-V1`, `kasse.BuildKassenbelegProcessData(positionen, -summe, -1)`. Negative Bruttobeträge am Satz, negative Barzahlung.
- Stornierung der Rückgabe: `Kassenbeleg-V1` mit positiven Beträgen, Faktor +1.
- Ein verrechneter Abschluss erzeugt zwei Transaktionen. Die Signatur läuft über die bestehende Outbox samt Ausfall-Pfad.

**DSFinV-K.**

- Rückgabe: `BON_TYP` „Beleg", `GV_TYP` „PfandRueckzahlung" (Anhang C, S. 50). Negative Menge und Beträge je USt-Schlüssel (Tz. 4.2.5), `BON_STORNO` 0, Zahlart Bar, ohne Referenz.
- Ohne Referenz, weil Pfandrückgaben unabhängig vom Verkauf als Entgeltminderung gelten (Abschn. 10.1 Abs. 8 Satz 3 UStAE).
- Stornierung: positiver Bon mit `GV_TYP` „PfandRueckzahlung" und Referenz `Transaktion` auf die `rueckgabeId`.
- `ABRECHNUNGSKREIS` ist der Tischname aus `tischId`, im Direktverkauf leer.
- `businesscases.csv` führt „PfandRueckzahlung" je USt-Schlüssel. Der Pfand-Verkauf bleibt „Umsatz".

**Reporting.**

- Soll-Kassenbestand und Bareinnahmen addieren `kj_extract_pfand_cents`. Umsatz je Steuersatz zieht Rückgaben am Satz ab.
- `gesamt_umsatz_cents` und `kasse.ComputeAbschlussSummen()` ziehen Rückgaben ab und addieren deren Stornierungen. Der Direktverkauf-Umsatz bleibt Verkaufsumsatz.
- Kassenbericht und Live-Dashboard führen Anzahl der Rückgaben und deren Nettosumme. `GetLiveReporting()` liest dieselbe `GetReportingStats`, das Dashboard braucht keine eigene Query.
- Abrechnung pro Servicekraft: `AbzugebenCents = KassiertCents − RuecknahmenCents − PfandCents`. Es zählen Rückgaben mit `tischId > 0` beim Akteur der Rückgabe. Eine Stornierung zählt für den Akteur der ursprünglichen Rückgabe.
- Der abzugebende Betrag darf negativ werden. Eigene Übersicht und Live-Dashboard rechnen genauso.
- Die Produktstatistik bleibt auf Bestellbasis.

**Beleg.** Der Pfandbeleg ist auf Verlangen druckbar und nachdruckbar.

- Er ist ein Kassenbeleg mit negativen Positionen, negativer Steuermatrix und negativem Gesamtbetrag.
- Der Beleg zur Stornierung ist ein Stornobeleg mit positiven Beträgen und „Storno zu Bon-Nr" der Rückgabe.

**Angewandte KISS-Schnitte** gegenüber dem PRD:

- Eigener Stream für jede Rückgabe statt Tisch-Session-Stream. Das entfernt die Fallunterscheidung nach bestehender Tisch-Session, jede Projektionsänderung und zwei Storno-Orte. Die Zuordnung zum Tisch trägt `tischId` ohnehin.
- `tagesabschluss-erstellt:v1` bekommt keine Folgeversion. Sein `umsatzGesamtCents` ist netto; Anzahl und Summe der Rückgaben zeigt der Kassenbericht aus dem Journal.
- Pfandrückgabe-Stornierungen erscheinen nicht in der Storno-Liste des Kassenberichts. Sie stehen in der Pfand-Historie und wirken über die Nettosumme.
- Der Pfandbeleg bekommt keinen eigenen Titel. Der Kassenbeleg mit negativen Beträgen und Positionen namens „Pfand" genügt.

## Inventory

- `database/migrations/01_initial.up.sql — kj_extract_umsatz_pro_steuersatz()` — wird in Migration 11 ersetzt; Vorbild `idx_kassenjournal_verkauf_id`.
- `database/migrations/README.md` — Regeln für additive, transaktionale Migrationen.
- `backend/domain/kasse/direktverkauf_events.go — NewDirektverkaufGetaetigtEvent()`, `NewDirektverkaufStorniertEvent()` — Vorbild der beiden Event-Konstruktoren.
- `backend/domain/kasse/direktverkauf.go — ComputeNichtStornierteVerkaufPositionen()`, `BuildDirektverkaufHistorieEintrag()` — Vorbild für Replay und Historie.
- `backend/domain/kasse/positionen.go — ValidatePositionRefs()`, `ResolvePositionen()` — Storno-Validierung, unverändert wiederverwendet.
- `backend/domain/kasse/subject.go — DirektverkaufSubject()` — Vorbild des Pfand-Subjects.
- `backend/domain/kasse/stream_type.go` — neuer Stream-Typ.
- `backend/domain/kasse/fiskalische_projektion.go — FiskalischeProjektion()` — einzige Stelle der Signaturpflicht; `exhaustive`-Switch.
- `backend/domain/kasse/tse_processdata.go — BuildKassenbelegProcessData()` — negative Beträge per Faktor −1.
- `backend/domain/kasse/tagesabschluss_summen.go — ComputeAbschlussSummen()` — Z-Bon-Summen; `exhaustive`-Switch.
- `backend/domain/kasse/kassensitzung_events.go — IsAbschlussEventType()` — Abschluss-Barriere; `exhaustive`-Switch.
- `backend/domain/kasse/event_json_contract_test.go — TestEventContract_AllTypesPinned()`, `TestEventContract_SQLLiterals()` — eingefrorene Verträge.
- `backend/domain/kasse/summen_positiv_test.go — TestGeldsummen_MussPositiv()` — Repo-Regel der Positiv-Summen.
- `backend/domain/produkt/produkt.go — Produkt`, `NewProdukt()`, `UpdateDetails()` — Pfand-Kennzeichen.
- `backend/api/kasse/enrichment/enrichment.go — EnrichPositionen()` — fette Positionen aus Eingaben.
- `backend/api/kasse/direktverkauf/application/command.go — DirektverkaufTaetigen()`, `persistVerkaufEvent()` — Idempotenz per Client-ID, Vorbild.
- `backend/api/kasse/direktverkauf/application/command.go — DirektverkaufStornieren()` — Vorbild der Pfand-Stornierung; Phase 4 liest es nur hier, Phase 3 bearbeitet die Datei.
- `backend/repository/repotest/kassenjournal.go — KassenjournalRepo` — gemeinsamer Fake der `command_test.go` von tischgeschaeft und direktverkauf; nur Phase 3 bearbeitet ihn.
- `scripts/test-integration.sh` — Integrationslauf; Phase 1 ergänzt die Filter `TEST_PKG` und `TEST_RUN`.
- `backend/api/kasse/tischgeschaeft/application/command.go — ZahlungKassieren()`, `loadTischState()` — Verrechnung am Tisch.
- `backend/repository/kassenjournal_repo/repo.go — writeEventInTx()`, `WriteTischSessionEventsAtomic()`, `EventExistsByTypeAndVorgangsID()` — Schreibpfad, Outbox, Idempotenz-Lookup.
- `backend/sqlc/queries/kassensitzungen.sql — GetKassenbestand` — Soll-Formel.
- `backend/sqlc/queries/reporting.sql — GetReportingStats`, `GetUmsatzPositionszeilen`, `GetKassiertProServicekraft`, `GetEigeneUebersicht` — Reporting.
- `backend/api/reporting/application/query.go — aggregateAbrechnungProServicekraft()`, `mergeServicekraefteLive()` — Abrechnung pro Servicekraft.
- `backend/api/fiskal/dsfinvk/mapper.go — belegeFromEvents()`, `abrechnungskreis()`, `buildBusinesscases()` — Export.
- `backend/dsfinvkpruefung/pruefung_integration_test.go — TestSeedExportBefundfrei()` — befundfreier Export der Seed-Daten.
- `backend/api/druck/beleg/application/kassenbeleg_command.go — KassenbelegDrucken()`, `resolveBelegQuelle()` — Belegdruck.
- `backend/api/druck/bondruck/application/escpos/formatter.go — FormatKassenbeleg()` — druckt Stornobelege und negative Beträge, bleibt unverändert.
- `backend/seed/engine.go — sitzungsBuilder`, `backend/seed/szenario.go — demoSzenario()` — Demo- und E2E-Daten.
- `frontend/src/service/table/ZahlungAbschluss.tsx — ZahlungAbschluss()` — Abschluss am Tisch.
- `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx — DirektverkaufAbschluss()` — Abschluss im Direktverkauf.
- `frontend/src/service/table/drawerUtils.ts — calculateZahlungsbetraege()` — Rückgeld und Trinkgeld.
- `frontend/src/admin/products/NewProductDialog.tsx`, `frontend/src/admin/products/EditProductDialog.tsx` — Admin-Formular.
- `frontend/src/lib/produktSchemas.ts — ProduktSchema` — gemeinsamer Produkt-Vertrag.
- `e2e/tests/direktverkauf-storno.mobile.spec.ts` — Vorbild der E2E-Spec.
- `docs/verfahrensdokumentation.md` §2, §3, §4 — Geschäftsvorfälle, Subject-Struktur, Prozesstyp-Tabelle.

## Open questions / Risks

- **Planer-Entscheidung, Owner-Bestätigung offen: eigener Stream je Rückgabe.** Das PRD legt Rückgaben am Tisch in den Tisch-Session-Stream. Der Plan nimmt immer einen eigenen Stream; Begründung unter „Angewandte KISS-Schnitte". Folge: Die Tisch-Historie zeigt Rückgaben über `tischId`, auch ohne Tisch-Session.
- **Planer-Entscheidung, Owner-Bestätigung offen: Z-Bon ohne Folgeversion.** `tagesabschluss-erstellt:v1` trägt keine Pfand-Kennzahlen. Anzahl und Summe stehen im Kassenbericht. Eine `:v2` lohnt erst, wenn der Z-Bon-Druck sie braucht.
- **Planer-Entscheidung, Owner-Bestätigung offen: Stornobeleg der Rückgabe.** Das PRD nennt nur den Pfandbeleg. Der Plan macht auch die Stornierung druckbar, weil jeder signierte Vorgang einen erstellbaren Beleg braucht (§ 146a Abs. 2 AO).
- **PRD nachziehen.** Bestätigt der Owner die drei Entscheidungen, schreibt der Lead vor Phase 1 das PRD um. Betroffen: §„Streams und Projektion", die Zeile „Kassenabschluss und Tagesabschluss führen …" unter §„Kasse und Reporting" und §„Beleg". Lehnt er eine ab, ändert sich zuerst der Plan.
- **Prüfpunkte aus dem PRD bleiben offen.** BMF-Schreiben vom 20.10.2014 (Warenumschließung oder Transporthilfsmittel) liegt nicht in den Rechtsquellen. Die DSFinV-K zitiert § 17 Abs. 1 Satz 7 UStG, der Gesetzestext hat Satz 8. Beides ist Steuerberater-Sache, kein Blocker.
- **`docs/decisions.md`** braucht eine Zeile „Pfandrückgabe als eigener Kassenvorgang, nie als Negativpreis". Der Lead trägt sie beim Landen ein; dieser Plan fasst die Datei nicht an.
- **Migrationsnummer.** 11 ist beim Schreiben frei. `docs/prds/prd-stationsmodell.md` plant ebenfalls die nächste freie Nummer. Landet sie zuerst, wird diese Datei vor dem Landen umbenannt.
- **Überschneidung im Abschluss.** `docs/prds/prd-kommentar-je-position.md` ändert `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx` ebenfalls. Parallel geplant, gibt es dort einen Merge-Konflikt.
- **Upgrade-Pfad nur in CI.** Der Job `upgrade-path` in `.github/workflows/ci.yml` prüft Spalte und ersetzte Funktion auf einer befüllten Bestandsinstanz. Lokal deckt `make check-integration` nur die Frischinstallation ab.
- **Kriterien für neue Tests scheitern, wenn der Test fehlt.** `go test -run` und `vitest -t` enden ohne Treffer mit 0. Deshalb sucht jedes Kriterium die PASS-Zeile des genannten Tests:
  - Go: `go test <pkg> -run '^TestX$' -v | grep -- '--- PASS: TestX '`.
  - Integration: `TEST_PKG=<pkg> TEST_RUN='^TestX$' make check-integration | grep -- '--- PASS: TestX '`. `TEST_RUN` setzt auch `-v`.
  - Vitest: `pnpm exec vitest run <file> -t '<name>' --reporter=verbose | grep -- '✓ .*<name>'`.
  - Mehrere Tests: `test "$(… | grep -cE -- '<muster>')" -eq N`.
  - Ein erweiterter bestehender Test bekommt einen neuen Untertest oder Testnamen, sonst besteht er auch ohne Erweiterung.
- **Dateibesitz bei parallelen Phasen.** Phasen 3 bis 6 laufen parallel. `backend/repository/repotest/kassenjournal.go` gehört Phase 3; Phase 4 nutzt einen paketlokalen Fake wie `mockHistorieRepo` in `backend/api/kasse/direktverkauf/application/query_test.go`.
- **`exhaustive`-Linter.** Jeder neue Event-Typ braucht einen Fall in jedem `switch` über `kasse.EventType`. `make lint-backend-full` findet fehlende.
- **Prosa-Gate.** `scripts/check-prose.sh` prüft alle getrackten Dateien. Code-Kommentare dürfen nicht „Phase N" oder Vorher-Wörter enthalten.

## Phase 1: Pfand-Produkt in den Stammdaten

**Depends on**: none

### Context

- `database/migrations/10_abholbon_kein_bon.up.sql` — Form einer additiven Migration.
- `backend/sqlc/queries/produkte.sql — GetProdukt`, `GetAlleProdukte`, `GetAktiveProdukte`, `CreateProdukt`, `UpdateProdukt` — lesen und schreiben das Kennzeichen.
- `backend/repository/produkt_repo/types.go — produktRowToDomain()` — Mapping Zeile → Domain.
- `backend/repository/produkt_repo/batch.go — GetProdukteByIDs()` — Roh-SQL, braucht die Spalte für die Anreicherung.
- `backend/repository/produkt_repo/repo.go — CreateProdukt()`, `UpdateProdukt()`.
- `backend/domain/produkt/produkt.go — Produkt`, `NewProdukt()`, `UpdateDetails()`.
- `backend/api/stammdaten/produkt/application/command.go — CreateProdukt()`, `UpdateProdukt()`.
- `backend/api/stammdaten/produkt/http/command_handler.go — createProduktRequest`, `updateProduktRequest` — zog-Schemas.
- `backend/api/stammdaten/produkt/http/query_handler.go — produkt` — gemeinsames DTO beider Lesepfade.
- `frontend/src/lib/produktSchemas.ts — ProduktSchema`.
- `frontend/src/admin/products/Produkt.ts`, `frontend/src/admin/products/ProduktBackend.ts`, `frontend/src/admin/products/NewProductDialog.tsx`, `frontend/src/admin/products/EditProductDialog.tsx`, `frontend/src/admin/products/ProductItem.tsx`.
- `docs/leitfaden/belege-steuersaetze.md` — Leitfaden zu Steuersätzen.
- `scripts/test-integration.sh` — startet Postgres, migriert, ruft `go test -tags=integration … ./...`.

### What to build

- `scripts/test-integration.sh` nimmt `TEST_PKG` und `TEST_RUN` aus der Umgebung. `TEST_PKG` ist eine Paketliste relativ zu `backend/`, Standard `./...`. `TEST_RUN` setzt `-run "$TEST_RUN" -v`. Der Usage-Kopf nennt beide.
- `database/migrations/11_pfandrueckgabe.up.sql` mit dem vollständigen Inhalt aus „Architectural decisions". Die Datei enthält Spalte, beide Indexe und beide Funktionen; spätere Phasen legen keine Migration an.
- `make sqlc` erneuert `backend/sqlc/dbgen/models.go` und `backend/sqlc/dbgen/produkte.sql.go`.
- `produkt.Produkt` bekommt `Pfand bool`. `NewProdukt()` und `UpdateDetails()` nehmen `pfand bool` an.
- Repo, Application und Handler reichen `pfand` durch. `GetProdukteByIDs()` liest die Spalte mit.
- Das DTO `produkt` in `backend/api/stammdaten/produkt/http/query_handler.go` trägt `pfand`. Zod-`ProduktSchema` bekommt `pfand: z.boolean()`.
- Neues und Bearbeiten-Formular zeigen den Schalter „Pfand-Produkt", Standard aus. Der Request trägt `pfand`.
- `frontend/src/admin/products/ProductItem.tsx` markiert ein Pfand-Produkt mit einem Badge „Pfand".
- `docs/leitfaden/belege-steuersaetze.md` erklärt: Pfand teilt den Satz der Hauptleistung. Gläserpfand zu Getränken 19 %, Tellerpfand zu Speisen 7 %.
- Der Leitfaden nennt den Schalter „Pfand-Produkt" und rät zu je einem Pfand-Produkt pro Satz.

Demo: Der Admin legt „Pfand" mit Variante „Glas 2,00 €" an und kennzeichnet es. Nach dem Neuladen steht das Badge.

### Acceptance criteria

- [ ] sqlc-Code passt zu Queries und Migration 11: `make check-sqlc`
- [ ] `scripts/test-integration.sh` besteht den Shell-Lint; seine Filter belegt das nächste Kriterium: `make check-shell`
- [ ] Anlegen und Bearbeiten speichern und lesen `pfand = true` (`TestCreateProdukt_StoresPfand`, `TestUpdateProdukt_StoresPfand` in `backend/repository/produkt_repo/repo_test.go`): `test "$(TEST_PKG=./repository/produkt_repo/ TEST_RUN='^Test(Create|Update)Produkt_StoresPfand$' make check-integration | grep -cE -- '^--- PASS: Test(Create|Update)Produkt_StoresPfand ')" -eq 2`
- [ ] Das Formular sendet `pfand: true`, wenn der Schalter an ist (Fall „sends pfand when the Pfand switch is on" in `frontend/src/admin/products/Products.test.tsx`): `cd frontend && pnpm exec vitest run src/admin/products/Products.test.tsx -t 'sends pfand when the Pfand switch is on' --reporter=verbose | grep -- '✓ .*sends pfand when the Pfand switch is on'`
- [ ] Jedes Leitfaden-Zitat existiert im Frontend: `./scripts/check-ui-labels.sh`
- [ ] Prosa-Gate grün: `./scripts/check-prose.sh`

## Phase 2: Reine Pfandrückgabe mit Barauszahlung

**Depends on**: 1

### Context

- `backend/domain/kasse/direktverkauf_events.go — NewDirektverkaufGetaetigtEvent()` — Vorbild Konstruktor, PositionIDs, Summe.
- `backend/domain/kasse/stream_type.go`, `backend/domain/kasse/subject.go — DirektverkaufSubject()`.
- `backend/domain/kasse/event_json_contract_test.go`, `backend/domain/kasse/summen_positiv_test.go`.
- `backend/domain/kasse/fiskalische_projektion.go — FiskalischeProjektion()`, `kassenbelegVorgang()`.
- `backend/domain/kasse/tagesabschluss_summen.go — ComputeAbschlussSummen()`.
- `backend/domain/kasse/kassensitzung_events.go — IsAbschlussEventType()`.
- `backend/seed/bondruck.go` — zwei `exhaustive`-Switches über `kasse.EventType`.
- `backend/api/kasse/enrichment/enrichment.go — EnrichPositionen()`.
- `backend/api/kasse/direktverkauf/application/command.go — getOffeneKassensitzungOderFehler()`, `writeEventOCC()` — Vorbild für Barriere und OCC.
- `backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go` — Vorbild Idempotenz-Test.
- `backend/repository/kassenjournal_repo/repo.go — writeEventInTx()` — Stream-Typ-Switch.
- `backend/sqlc/queries/kassensitzungen.sql — GetKassenbestand`, `backend/sqlc/queries/reporting.sql — GetReportingStats`, `GetUmsatzPositionszeilen`.
- `backend/repository/kassenjournal_repo/repo_test.go — TestGetKassenbestand_DirektverkaufIncreasesThenStornoDecreases()` — Vorbild Soll-Test.
- `backend/repository/reporting_repo/repo_test.go — TestGetReporting_UmsatzProSteuersatzZiehtWarenruecknahmeAb()` — Vorbild Umsatz-Test.
- `backend/repository/reporting_repo/summen_abschluss_test.go — TestComputeAbschlussSummen_AequivalenzMitSQLReporting()`.
- `backend/api/stammdaten/produkt/application/query.go`, `backend/api/stammdaten/produkt/http/query_handler.go — GetActiveProdukteHandler()`.
- `backend/api/service.go` — Routen des Service-Bereichs.
- `frontend/src/service/product/ProduktBackend.ts — getAktiveProdukte()`, `frontend/src/service/product/hooks.ts`.
- `frontend/src/service/table/ZahlungAbschluss.tsx`, `frontend/src/service/table/ZahlungDrawer.tsx`, `frontend/src/service/table/ZahlungTab.tsx`, `frontend/src/service/table/TablePage.tsx`.
- `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx`, `frontend/src/service/direktverkauf/DirektverkaufDrawer.tsx`, `frontend/src/service/direktverkauf/DirektverkaufTab.tsx`.
- `frontend/src/service/table/DockActionButton.tsx`, `frontend/src/service/table/AbschlussLeer.tsx`.

### What to build

Backend, Domäne:

- `kasse.StreamTypePfandrueckgabe` und `kasse.PfandrueckgabeSubject(zNr, rueckgabeID)`.
- Neue Datei `backend/domain/kasse/pfandrueckgabe_events.go` mit beiden Event-Typen, Data-Structs und zog-Schemas aus „Architectural decisions". `TischID` ist `z.Int().GTE(0).LTE(math.MaxInt32)` ohne `Required()`, weil 0 gültig ist.
- `kasse.NewPfandZurueckgegebenEvent(subject, rueckgabeID, tischID, userID, userName, positionen, kommentar)` vergibt PositionIDs und rechnet die Summe.
- `kasse.NewPfandrueckgabeStorniertEvent(subject, stornierungID, rueckgabeID, userID, userName, positionen, gesamtStornierungCents, kommentar)`.
- Vertragstests `TestEventContract_PfandZurueckgegebenV1` und `TestEventContract_PfandrueckgabeStorniertV1`. Beide Typen in `allEventTypes`, `contractedTypes` und `TestEventContract_SQLLiterals`.
- `TestGeldsummen_MussPositiv` bekommt beide Summen-Felder als Fälle.
- `FiskalischeProjektion()`: Rückgabe `kassenbelegVorgang(pos, -summe, -1)`, Stornierung `kassenbelegVorgang(pos, summe, 1)`.
- `ComputeAbschlussSummen()`: Rückgabe mindert `UmsatzCents`, Stornierung erhöht ihn.
- `IsAbschlussEventType()` und beide Switches in `backend/seed/bondruck.go` führen die Typen als leere Fälle.

Backend, Anwendung:

- `enrichment.EnrichPfandPositionen()` ruft `EnrichPositionen()` und lehnt Produkte ohne `Pfand` mit `enrichment.ErrKeinPfandProdukt` ab.
- `writeEventInTx()` kennt `StreamTypePfandrueckgabe` ohne Projektion.
- Neues Paket `backend/api/kasse/pfand/` mit `backend/api/kasse/pfand/application/command.go — Command.PfandZurueckgeben()`, `backend/api/kasse/pfand/application/errors.go` und `backend/api/kasse/pfand/http/command_handler.go — PfandZurueckgebenHandler()`.
- `PfandZurueckgeben()` prüft die offene Kassensitzung samt Barriere. Bei `tischId > 0` muss der Tisch existieren und aktiv sein.
- Es reichert an, schreibt `version = 1` in den eigenen Stream und behandelt den OCC-Konflikt per `EventExistsByTypeAndVorgangsID(…, "rueckgabeId")` wie `DirektverkaufTaetigen()`.
- Fehler-Mapping: `kasse_nicht_geoeffnet` und `kasse_wird_abgeschlossen` 409, `kein_pfand_produkt` 400, `tisch_not_found` 404, `tisch_not_active` 409.
- Route `service/pfand-zurueckgeben` in `backend/api/service.go`. `TestBerechtigungsMatrix` prüft jede Route des Bereichs ohne Änderung.

Backend, SQL:

- `GetKassenbestand`: `soll_bestand_cents` und `bareinnahmen_cents` addieren `kj_extract_pfand_cents`; die `type IN`-Liste kennt beide Typen.
- `GetReportingStats.gesamt_umsatz_cents` addiert `kj_extract_pfand_cents`; die `WHERE type IN`-Liste von `GetReportingStats` kennt beide Typen. Ohne sie bleibt der Pfand-Term immer 0.
- `GetUmsatzPositionszeilen` filtert zusätzlich auf beide Typen.
- `make sqlc` erneuert `backend/sqlc/dbgen/kassensitzungen.sql.go` und `backend/sqlc/dbgen/reporting.sql.go`.
- `service/get-aktive-produkte` liefert zusätzlich `pfandProdukte`, gefiltert in `backend/api/stammdaten/produkt/application/query.go`.

Frontend:

- Neuer Ordner `frontend/src/service/pfand/`: `frontend/src/service/pfand/Pfand.ts` (`PfandPositionInputSchema`, `PfandZurueckgebenSchema`), `frontend/src/service/pfand/PfandBackend.ts — pfandZurueckgeben()`, `frontend/src/service/pfand/PfandAbschnitt.tsx`. Den Abschnitt prüfen die Tests der beiden Abschlüsse.
- `PfandAbschnitt` heißt „Pfand zurück" und listet jede Variante der `pfandProdukte` mit Mengen-Stepper. Er meldet Mengen und Summe nach oben.
- `ProduktBackend.getAktiveProdukte()` liefert `produkte` und `pfandProdukte`; die Aufrufer reichen `pfandProdukte` an Zahlung- und Direktverkauf-Tab.
- `frontend/src/service/table/TablePage.test.tsx` und `frontend/src/service/direktverkauf/DirektverkaufPage.test.tsx` antworten auf `service/get-aktive-produkte` zusätzlich mit `pfandProdukte: []`.
- `ZahlungAbschluss` und `DirektverkaufAbschluss` zeigen den Abschnitt, sobald es Pfand-Produkte gibt. Er ist nur aktiv, solange keine Position gewählt ist; Verrechnung kommt erst in Phase 3.
- Ohne Positionen, aber mit Pfand, heißt der Button „Pfand auszahlen". Er ruft `pfandZurueckgeben()` mit `tischId` des Tisches oder 0 und dem Kommentar aus `KommentarField`.
- Jede Rückgabe trägt eine `rueckgabeId` aus `crypto.randomUUID()`. Ein neuer Vorgang erneuert sie, ein Wiederholversuch behält sie.
- Der Dock-Button in `ZahlungDrawer` und `DirektverkaufDrawer` öffnet den Abschluss auch ohne Auswahl, wenn Pfand-Produkte existieren; Label „Pfand zurück".
- Der Zahlung-Tab ist an einem Tisch ohne Bestellung erreichbar und bietet dort nur „Pfand zurück". Erfolgsmeldung „Pfand ausgezahlt."

Demo: Im Direktverkauf zwei Gläser zurücknehmen. Der Soll-Kassenbestand sinkt um 4,00 €, der Umsatz 19 % um 4,00 € brutto.

### Acceptance criteria

- [ ] Beide Event-Verträge sind eingefroren und vollständig gelistet: `cd backend && test "$(go test ./domain/kasse/ -run '^TestEventContract_(PfandZurueckgegebenV1|PfandrueckgabeStorniertV1|AllTypesPinned|SQLLiterals)$' -v | grep -cE -- '^--- PASS: TestEventContract_(PfandZurueckgegebenV1|PfandrueckgabeStorniertV1|AllTypesPinned|SQLLiterals) ')" -eq 4`
- [ ] Beide Summen-Felder lehnen 0 ab (Fälle `PfandZurueckgegeben.GesamtbetragCents`, `PfandrueckgabeStorniert.GesamtStornierungCents`): `cd backend && test "$(go test ./domain/kasse/ -run '^TestGeldsummen_MussPositiv$' -v | grep -cE -- '--- PASS: TestGeldsummen_MussPositiv/(PfandZurueckgegeben\.GesamtbetragCents|PfandrueckgabeStorniert\.GesamtStornierungCents)/lehntNullAb ')" -eq 2`
- [ ] Die Rückgabe erzeugt negatives `Kassenbeleg-V1`, die Stornierung positives (Fälle „pfand zurueckgegeben: …" und „pfandrueckgabe storniert: …" in `TestFiskalischeProjektion`): `cd backend && test "$(go test ./domain/kasse/ -run '^TestFiskalischeProjektion$' -v | grep -cE -- '--- PASS: TestFiskalischeProjektion/(pfand_zurueckgegeben|pfandrueckgabe_storniert):')" -eq 2`
- [ ] Z-Bon-Summen und SQL-Reporting stimmen mit Rückgabe und Stornierung überein (`TestComputeAbschlussSummen_AequivalenzMitPfandrueckgabe` in `backend/repository/reporting_repo/summen_abschluss_test.go`): `TEST_PKG=./repository/reporting_repo/ TEST_RUN='^TestComputeAbschlussSummen_AequivalenzMitPfandrueckgabe$' make check-integration | grep -- '--- PASS: TestComputeAbschlussSummen_AequivalenzMitPfandrueckgabe '`
- [ ] Eine Rückgabe mindert Soll und Bareinnahmen, ihre Stornierung hebt das auf (`TestGetKassenbestand_PfandrueckgabeMindertSoll` in `backend/repository/kassenjournal_repo/repo_test.go`): `TEST_PKG=./repository/kassenjournal_repo/ TEST_RUN='^TestGetKassenbestand_PfandrueckgabeMindertSoll$' make check-integration | grep -- '--- PASS: TestGetKassenbestand_PfandrueckgabeMindertSoll '`
- [ ] Umsatz je Steuersatz zieht eine Rückgabe am Satz des Pfand-Produkts ab (`TestGetReporting_UmsatzProSteuersatzZiehtPfandrueckgabeAb` in `backend/repository/reporting_repo/repo_test.go`): `TEST_PKG=./repository/reporting_repo/ TEST_RUN='^TestGetReporting_UmsatzProSteuersatzZiehtPfandrueckgabeAb$' make check-integration | grep -- '--- PASS: TestGetReporting_UmsatzProSteuersatzZiehtPfandrueckgabeAb '`
- [ ] Zweimal dieselbe `rueckgabeId` bucht einmal, mit genau einem Signaturauftrag (`TestPfandZurueckgeben_DuplikatRueckgabeId_IdempotenterErfolg` in `backend/api/kasse/pfand/application/idempotenz_integration_test.go`): `TEST_PKG=./api/kasse/pfand/application/ TEST_RUN='^TestPfandZurueckgeben_DuplikatRueckgabeId_IdempotenterErfolg$' make check-integration | grep -- '--- PASS: TestPfandZurueckgeben_DuplikatRueckgabeId_IdempotenterErfolg '`
- [ ] Eine Nicht-Pfand-Variante wird mit `kein_pfand_produkt` abgelehnt, ohne Event (`TestPfandZurueckgeben_KeinPfandProdukt_KeinEvent` in `backend/api/kasse/pfand/application/command_test.go`): `cd backend && go test ./api/kasse/pfand/application/ -run '^TestPfandZurueckgeben_KeinPfandProdukt_KeinEvent$' -v | grep -- '--- PASS: TestPfandZurueckgeben_KeinPfandProdukt_KeinEvent '`
- [ ] Der Abschluss ohne Positionen sendet nur `pfandZurueckgeben` mit der Tisch-ID (Fall „sends only the Pfand return with the table id" in `frontend/src/service/table/ZahlungAbschluss.test.tsx`): `cd frontend && pnpm exec vitest run src/service/table/ZahlungAbschluss.test.tsx -t 'sends only the Pfand return with the table id' --reporter=verbose | grep -- '✓ .*sends only the Pfand return with the table id'`
- [ ] Der Direktverkauf-Abschluss ohne Positionen sendet `tischId: 0` (Fall „sends the Pfand return with table id 0" in `frontend/src/service/direktverkauf/DirektverkaufAbschluss.test.tsx`): `cd frontend && pnpm exec vitest run src/service/direktverkauf/DirektverkaufAbschluss.test.tsx -t 'sends the Pfand return with table id 0' --reporter=verbose | grep -- '✓ .*sends the Pfand return with table id 0'`
- [ ] Tisch- und Direktverkaufsseite laden mit `pfandProdukte` im Response: `cd frontend && pnpm exec vitest run src/service/table/TablePage.test.tsx src/service/direktverkauf/DirektverkaufPage.test.tsx`
- [ ] Alle Switches über `kasse.EventType` sind vollständig: `make lint-backend-full`

## Phase 3: Verrechnung beim Kassieren und im Direktverkauf

**Depends on**: 2

### Context

- `backend/repository/kassenjournal_repo/repo.go — WriteTischSessionEventsAtomic()`, `WriteEventWithDruckauftraege()`, `writeEventInTx()` — Vorbild der atomaren Schreibweise.
- `backend/repository/kassenjournal_repo/repo_test.go — TestWriteUmbuchung_RollsBackWhenTargetWriteFails()` — Vorbild Rollback-Test.
- `backend/repository/repotest/kassenjournal.go — KassenjournalRepo.WriteTischSessionEventsAtomic()` — Fake der `command_test.go`-Dateien.
- `backend/api/kasse/tischgeschaeft/application/command.go — ZahlungKassieren()`, `loadTischState()`.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go — zahlungKassierenRequest`, `zahlungKassierenSchema`.
- `backend/api/kasse/direktverkauf/application/command.go — DirektverkaufTaetigen()`.
- `backend/api/kasse/direktverkauf/http/command_handler.go — direktverkaufTaetigenRequest`, `direktverkaufTaetigenSchema`.
- `backend/api/kasse/enrichment/enrichment.go — EnrichPfandPositionen()`.
- `frontend/src/service/table/Zahlung.ts — ZahlungKassierenSchema`, `frontend/src/service/direktverkauf/Direktverkauf.ts — DirektverkaufTaetigenSchema`.
- `frontend/src/service/table/ZahlungAbschluss.tsx`, `frontend/src/service/direktverkauf/DirektverkaufAbschluss.tsx`, `frontend/src/service/table/BarzahlungFelder.tsx`, `frontend/src/service/table/drawerUtils.ts — calculateZahlungsbetraege()`.

### What to build

Backend:

- `kassenjournal_repo.Repository.WriteEventsAtomic(ctx, kassensitzungNr, []kassenjournal_repo.NeuesEvent)` schreibt mehrere Events über Stream-Typen hinweg in einer Transaktion.
- `NeuesEvent` trägt `Event`, `StreamType` und optional `BuildAuftraege`. Der Signatur-Worker wird nach dem Commit einmal angestoßen.
- `repotest.KassenjournalRepo` in `backend/repository/repotest/kassenjournal.go` bekommt `WriteEventsAtomic()`, alles oder nichts wie `WriteTischSessionEventsAtomic()`.
- `enrichment.PfandEingabe{RueckgabeID, Positionen}` ist die gemeinsame Eingabe beider Commands.
- `ZahlungKassieren()` nimmt `pfand *enrichment.PfandEingabe`. Mit Pfand prüft es zuerst per `EventExistsByTypeAndVorgangsID()` die Rückgabe-ID; existiert sie, antwortet es Erfolg ohne Buchung.
- Danach validiert es Zahlung und Pfand-Positionen vollständig, bevor es schreibt. Zahlung (`version = LastEventVersion + 1`) und Rückgabe (`version = 1`, `tischId` des Tisches) gehen per `WriteEventsAtomic()` hinaus.
- Ein `ErrAlreadyExists` beim Schreiben führt zum Lookup der Rückgabe-ID: gefunden heißt Erfolg, sonst 409.
- `DirektverkaufTaetigen()` nimmt dasselbe `pfand`. Verkauf mit Arbeitsbon-Builder und Rückgabe mit `tischId = 0` gehen per `WriteEventsAtomic()` hinaus. Der OCC-Pfad über `verkaufId` bleibt.
- Beide Handler nehmen optional `pfand: {rueckgabeId, positionen}` an, geprüft per zog.

Frontend:

- `ZahlungKassierenSchema` und `DirektverkaufTaetigenSchema` bekommen optional `pfand`.
- Der Pfand-Abschnitt ist auch neben gewählten Positionen aktiv.
- Beide Abschlüsse zeigen bei Positionen und Pfand eine Netto-Zeile: Kaufsumme, „Pfand zurück" negativ, Differenz.
- Ist die Differenz positiv, rechnen `BarzahlungFelder` mit ihr Rückgeld und Trinkgeld.
- Ist sie null oder negativ, zeigt der Abschluss „Auszahlen" mit dem Betrag; die Barzahlungsfelder entfallen.
- Der Submit sendet Zahlung oder Verkauf mit `pfand` in einem Request.

Demo: Am Tisch zwei Bier für 9,00 € kassieren und zwei Gläser verrechnen. Der Drawer zeigt 5,00 € netto, das Journal hat zwei signierte Vorgänge.

### Acceptance criteria

- [ ] `WriteEventsAtomic` rollt beide Events zurück, wenn das zweite scheitert (`TestWriteEventsAtomic_RollsBackBothEvents` in `backend/repository/kassenjournal_repo/repo_test.go`): `TEST_PKG=./repository/kassenjournal_repo/ TEST_RUN='^TestWriteEventsAtomic_RollsBackBothEvents$' make check-integration | grep -- '--- PASS: TestWriteEventsAtomic_RollsBackBothEvents '`
- [ ] Ein wiederholter verrechneter Abschluss am Tisch bucht Zahlung und Rückgabe einmal (`TestZahlungKassieren_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg` in `backend/api/kasse/tischgeschaeft/application/idempotenz_integration_test.go`): `TEST_PKG=./api/kasse/tischgeschaeft/application/ TEST_RUN='^TestZahlungKassieren_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg$' make check-integration | grep -- '--- PASS: TestZahlungKassieren_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg '`
- [ ] Ein wiederholter verrechneter Direktverkauf bucht Verkauf und Rückgabe einmal (`TestDirektverkaufTaetigen_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg` in `backend/api/kasse/direktverkauf/application/idempotenz_integration_test.go`): `TEST_PKG=./api/kasse/direktverkauf/application/ TEST_RUN='^TestDirektverkaufTaetigen_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg$' make check-integration | grep -- '--- PASS: TestDirektverkaufTaetigen_MitPfand_DuplikatRueckgabeId_IdempotenterErfolg '`
- [ ] Eine ungültige Pfand-Position verhindert auch die Zahlung, kein Event wird geschrieben (`TestZahlungKassieren_UngueltigesPfand_KeinEvent` in `backend/api/kasse/tischgeschaeft/application/command_test.go`): `cd backend && go test ./api/kasse/tischgeschaeft/application/ -run '^TestZahlungKassieren_UngueltigesPfand_KeinEvent$' -v | grep -- '--- PASS: TestZahlungKassieren_UngueltigesPfand_KeinEvent '`
- [ ] Der Tisch-Abschluss zeigt Nettobetrag und Rückgeld und sendet `pfand` (Fall „nets the Pfand against the payment" in `frontend/src/service/table/ZahlungAbschluss.test.tsx`): `cd frontend && pnpm exec vitest run src/service/table/ZahlungAbschluss.test.tsx -t 'nets the Pfand against the payment' --reporter=verbose | grep -- '✓ .*nets the Pfand against the payment'`
- [ ] Der Direktverkauf-Abschluss zeigt „Auszahlen" bei mehr Pfand als Kaufsumme (Fall „shows Auszahlen when the Pfand exceeds the sale" in `frontend/src/service/direktverkauf/DirektverkaufAbschluss.test.tsx`): `cd frontend && pnpm exec vitest run src/service/direktverkauf/DirektverkaufAbschluss.test.tsx -t 'shows Auszahlen when the Pfand exceeds the sale' --reporter=verbose | grep -- '✓ .*shows Auszahlen when the Pfand exceeds the sale'`

## Phase 4: Historie, Stornierung und Pfandbeleg

**Depends on**: 2

### Context

- `backend/domain/kasse/direktverkauf.go — ComputeNichtStornierteVerkaufPositionen()`, `BuildDirektverkaufHistorieEintrag()`.
- `backend/domain/kasse/direktverkauf_test.go` — Vorbild Storno-Validierung.
- `backend/api/kasse/direktverkauf/application/query.go`; das Vorbild `DirektverkaufStornieren()` steht im Inventory.
- `backend/api/kasse/direktverkauf/application/query_test.go — mockHistorieRepo` — Vorbild des paketlokalen Fakes.
- `backend/api/kasse/direktverkauf/http/query_handler.go — GetDirektverkaufHistorieHandler()` — Vorbild DTO.
- `backend/sqlc/queries/kassenjournal.sql — ReadDirektverkaufEvents` — Vorbild Lese-Query.
- `backend/api/serviceleitung.go`, `backend/api/service.go` — Routen.
- `backend/api/druck/beleg/application/kassenbeleg_command.go — KassenbelegDruckenCommand`, `belegQuelle`, `resolveBelegQuelle()`, `resolveDirektverkaufStornobeleg()`.
- `backend/api/druck/beleg/http/command_handler.go — readKassenbelegCommand()`, `KassenbelegDruckenHandler()`.
- `backend/api/druck/bondruck/application/escpos/formatter.go — KassenbelegData` — `Stornobeleg` und vorab negierte Beträge, unverändert genutzt.
- `frontend/src/service/direktverkauf/DirektverkaufHistorie.tsx`, `frontend/src/service/direktverkauf/DirektverkaufStornoDrawer.tsx` — Vorbild UI.
- `frontend/src/service/table/TablePage.tsx`, `frontend/src/service/direktverkauf/DirektverkaufPage.tsx` — Historie-Tabs.
- `frontend/src/lib/Auth.ts — AuthSingleton.canCancel`.

### What to build

Backend:

- Neue Datei `backend/domain/kasse/pfandrueckgabe.go`: `ComputeNichtStorniertePfandPositionen(events)` und `BuildPfandrueckgabeHistorieEintrag(events)`.
- `PfandrueckgabeHistorieEintrag` trägt `RueckgabeID`, `TischID`, `UserName`, `ZurueckgegebenAm`, `Positionen`, `GesamtbetragCents`, `Kommentar`, `OffenePositionen`, `GesamtStorniertCents`, `Stornierungen`.
- `pfand.Command.PfandrueckgabeStornieren()` in `backend/api/kasse/pfand/application/command.go` liest den Stream der Rückgabe. Leer heißt `rueckgabe_not_found` 404.
- Eine bekannte `stornierungId` antwortet Erfolg ohne Buchung. Zu große Mengen enden in `position_nicht_stornierbar` 409.
- Sonst schreibt es `version = n + 1` mit `ResolvePositionen()`. Ein Konflikt führt zum Lookup der Stornierungs-ID.
- `pfand.Query.GetPfandrueckgabeHistorie(ctx, tischID)` in neuer Datei `backend/api/kasse/pfand/application/query.go` liest per neuer Query `ReadPfandrueckgabeEvents` alle Pfand-Events der offenen Kassensitzung. Es gruppiert nach Subject, filtert auf `tischId` und sortiert neueste zuerst.
- `backend/api/kasse/pfand/application/query_test.go` nutzt einen paketlokalen Fake mit `ReadPfandrueckgabeEvents()`. `backend/repository/repotest/kassenjournal.go` bleibt in dieser Phase unberührt.
- Repo-Methode `ReadPfandrueckgabeEvents()` in neuer Datei `backend/repository/kassenjournal_repo/pfandrueckgabe.go`; `make sqlc` erneuert `backend/sqlc/dbgen/kassenjournal.sql.go`.
- `PfandrueckgabeStornierenHandler()` in `backend/api/kasse/pfand/http/command_handler.go`, `GetPfandrueckgabeHistorieHandler()` in neuer Datei `backend/api/kasse/pfand/http/query_handler.go`.
- Routen `serviceleitung/pfandrueckgabe-stornieren` in `backend/api/serviceleitung.go` und `service/get-pfandrueckgabe-historie` in `backend/api/service.go`.
- Belegdruck: `KassenbelegDruckenCommand.RueckgabeID`; die Formen `{rueckgabeId}` und `{rueckgabeId, stornierungId}` in `readKassenbelegCommand()`.
- `belegQuelle` trennt `Negativ` (Beträge negieren) von `Stornobeleg` (Titel). Bestehende Stornobelege setzen beides.
- Der Pfandbeleg setzt nur `Negativ`. Der Beleg zur Stornierung setzt nur `Stornobeleg` und nennt als „Storno zu Bon-Nr" die Rückgabe.
- Fehler-Mapping: `rueckgabe_not_found` 404, `stornierung_not_found` 404.

Frontend:

- `frontend/src/service/pfand/Pfand.ts` bekommt `PfandrueckgabeHistorieEintragSchema`, `PfandrueckgabeStornierenSchema` und `PfandbelegDruckenSchema`.
- `PfandBackend` bekommt `getPfandrueckgabeHistorie()`, `pfandrueckgabeStornieren()` und `pfandbelegDrucken()`. Ein Hook in `frontend/src/service/pfand/hooks.ts` lädt die Historie.
- `frontend/src/service/pfand/PfandrueckgabeHistorie.tsx` listet die Rückgaben mit Uhrzeit, Servicekraft, Betrag und Badge „storniert" oder „teilweise storniert".
- Jede Zeile bietet „Beleg drucken". `canCancel` sieht zusätzlich „Stornieren…".
- `frontend/src/service/pfand/PfandStornoDrawer.tsx` wählt Positionen und Mengen aus `offenePositionen`, Kommentar Pflicht.
- Die Tisch-Historie in `frontend/src/service/table/TablePage.tsx` zeigt den Abschnitt „Pfandrückgaben" mit der Tisch-ID. Die Direktverkauf-Historie in `frontend/src/service/direktverkauf/DirektverkaufPage.tsx` zeigt ihn mit 0.

Demo: Eine irrtümliche Rückgabe am Tisch als Serviceleitung stornieren. Die Zeile trägt „storniert", der Soll-Kassenbestand steigt zurück.

### Acceptance criteria

- [ ] Offene Mengen sinken mit jeder Teilstornierung, die Historie summiert die Stornierungen (`TestComputeNichtStorniertePfandPositionen_PartialThenFullStorno`, `TestBuildPfandrueckgabeHistorieEintrag_AggregatesStornos` in `backend/domain/kasse/pfandrueckgabe_test.go`): `cd backend && test "$(go test ./domain/kasse/ -run '^(TestComputeNichtStorniertePfandPositionen_PartialThenFullStorno|TestBuildPfandrueckgabeHistorieEintrag_AggregatesStornos)$' -v | grep -cE -- '^--- PASS: (TestComputeNichtStorniertePfandPositionen_PartialThenFullStorno|TestBuildPfandrueckgabeHistorieEintrag_AggregatesStornos) ')" -eq 2`
- [ ] Eine Menge über der offenen Rückgabe endet in `position_nicht_stornierbar`, ohne Event (`TestPfandrueckgabeStornieren_ZuGrosseMenge_KeinEvent` in `backend/api/kasse/pfand/application/command_test.go`): `cd backend && go test ./api/kasse/pfand/application/ -run '^TestPfandrueckgabeStornieren_ZuGrosseMenge_KeinEvent$' -v | grep -- '--- PASS: TestPfandrueckgabeStornieren_ZuGrosseMenge_KeinEvent '`
- [ ] Zweimal dieselbe `stornierungId` bucht einmal (`TestPfandrueckgabeStornieren_DuplikatStornierungId_IdempotenterErfolg` in `backend/api/kasse/pfand/application/idempotenz_integration_test.go`): `TEST_PKG=./api/kasse/pfand/application/ TEST_RUN='^TestPfandrueckgabeStornieren_DuplikatStornierungId_IdempotenterErfolg$' make check-integration | grep -- '--- PASS: TestPfandrueckgabeStornieren_DuplikatStornierungId_IdempotenterErfolg '`
- [ ] Die Historie liefert je Tisch nur dessen Rückgaben, 0 nur die des Direktverkaufs (`TestGetPfandrueckgabeHistorie_FiltertNachTischID` in `backend/api/kasse/pfand/application/query_test.go`): `cd backend && go test ./api/kasse/pfand/application/ -run '^TestGetPfandrueckgabeHistorie_FiltertNachTischID$' -v | grep -- '--- PASS: TestGetPfandrueckgabeHistorie_FiltertNachTischID '`
- [ ] Die Stornierung liegt nur im Serviceleitungs-Bereich, die Historie im Service-Bereich; `TestBerechtigungsMatrix` prüft dann die Rollen beider Bereiche: `grep -q '"/pfandrueckgabe-stornieren"' backend/api/serviceleitung.go && ! grep -q '"/pfandrueckgabe-stornieren"' backend/api/service.go && grep -q '"/get-pfandrueckgabe-historie"' backend/api/service.go`
- [ ] Der Pfandbeleg trägt negative Beträge, der Stornobeleg zur Rückgabe positive mit Verweis auf deren Bon-Nr (`TestKassenbelegDrucken_Pfandrueckgabe_NegativeBetraege`, `TestKassenbelegDrucken_PfandStorno_DruckbarAlsStornobeleg` in `backend/api/druck/beleg/application/kassenbeleg_command_test.go`): `cd backend && test "$(go test ./api/druck/beleg/application/ -run '^(TestKassenbelegDrucken_Pfandrueckgabe_NegativeBetraege|TestKassenbelegDrucken_PfandStorno_DruckbarAlsStornobeleg)$' -v | grep -cE -- '^--- PASS: (TestKassenbelegDrucken_Pfandrueckgabe_NegativeBetraege|TestKassenbelegDrucken_PfandStorno_DruckbarAlsStornobeleg) ')" -eq 2`
- [ ] Stornierte Rückgaben sind markiert, „Stornieren…" fehlt ohne `canCancel`, und der Drawer sendet gewählte Positionen samt Kommentar (Fälle „marks cancelled returns" und „hides Stornieren without canCancel" in `frontend/src/service/pfand/PfandrueckgabeHistorie.test.tsx`, „sends the selected positions with the comment" in `frontend/src/service/pfand/PfandStornoDrawer.test.tsx`): `cd frontend && test "$(pnpm exec vitest run src/service/pfand/ --reporter=verbose | grep -cE -- '✓ .*(marks cancelled returns|hides Stornieren without canCancel|sends the selected positions with the comment)')" -eq 3`

## Phase 5: Abrechnung und Kassenabschluss

**Depends on**: 2

### Context

- `backend/sqlc/queries/reporting.sql — GetReportingStats`, `GetKassiertProServicekraft`, `GetEigeneUebersicht`.
- `backend/repository/reporting_repo/repo.go — GetReporting()`, `GetLiveReporting()`, `GetEigeneUebersicht()`.
- `backend/domain/reporting/reporting.go — Summary`, `AbrechnungServicekraft`, `ServicekraftLive`, `EigeneUebersicht`.
- `backend/api/reporting/application/query.go — aggregateAbrechnungProServicekraft()`, `mergeServicekraefteLive()`.
- `backend/api/reporting/http/query_handler.go` — Response-DTOs.
- `backend/api/reporting/application/abrechnung_integration_test.go — TestAbrechnung_DirektverkaufBleibtAussen()` — Vorbild.
- `frontend/src/admin/reporting/types.ts`, `frontend/src/admin/reporting/ReportingResults.tsx`, `frontend/src/admin/reporting/LiveReportingSection.tsx`, `frontend/src/admin/reporting/SummaryCard.tsx`.
- `frontend/src/service/table/Tisch.ts — EigeneUebersichtSchema`, `frontend/src/service/table/EigeneUebersicht.tsx`.
- `backend/sqlc/queries/reporting.sql — GetEigeneUebersicht`, `backend/domain/reporting/reporting.go`, `frontend/src/admin/reporting/types.ts`, `frontend/src/service/table/Tisch.ts` — Kommentare mit „nie negativ" und der alten Formel.

### What to build

Backend:

- `GetReportingStats` liefert `anzahl_pfandrueckgaben` und `pfandrueckgaben_cents`, die Nettosumme `-SUM(kj_extract_pfand_cents)`.
- Neue Query `GetPfandProServicekraft`: je Akteur der Rückgabe die Nettosumme aller Rückgaben mit `tischId > 0`. Stornierungen werden über `rueckgabeId` dem Akteur der Rückgabe zugeordnet.
- `GetEigeneUebersicht` liefert `pfand_cents` nach derselben Regel. `abzugeben_cents = zahlungen − ruecknahmen − pfand`. Der Query-Kommentar „nie negativ" entfällt.
- Diese Kommentare nennen danach die neue Formel und „darf negativ werden": `backend/domain/reporting/reporting.go — AbrechnungServicekraft` (Doc-Kommentar) und `EigeneUebersicht.AbzugebenCents` („nie negativ").
- `make sqlc` erneuert `backend/sqlc/dbgen/reporting.sql.go`.
- `reporting.Summary` bekommt `AnzahlPfandrueckgaben` und `PfandrueckgabenCents`.
- `AbrechnungServicekraft`, `ServicekraftLive` und `EigeneUebersicht` bekommen `PfandCents`.
- `aggregateAbrechnungProServicekraft()` nimmt die Pfand-Zeilen dazu. Eine Servicekraft nur mit Rückgaben bekommt eine eigene Zeile.
- `AbzugebenCents = KassiertCents − RuecknahmenCents − PfandCents`. `mergeServicekraefteLive()` reicht `PfandCents` durch.
- Die DTOs in `backend/api/reporting/http/query_handler.go` tragen `anzahlPfandrueckgaben`, `pfandrueckgabenCents` und `pfandCents`.

Frontend:

- Die Zod-Schemas kennen die neuen Felder. `abzugebenCents` ist schon `z.number().int()` und erlaubt negative Werte.
- Die Formel-Kommentare in `frontend/src/admin/reporting/types.ts` und `frontend/src/service/table/Tisch.ts` ziehen das Pfand mit ab.
- Kassenbericht und Live-Dashboard zeigen die Kennzahl „Pfandrückgaben" mit Anzahl und Summe. Die Abrechnung pro Servicekraft zeigt die Spalte „Pfand".
- Die Eigene Übersicht zeigt „Pfand ausgezahlt" und einen negativen abzugebenden Betrag als solchen.

Demo: Eine Servicekraft kassiert 20,00 € und zahlt am Tisch 4,00 € Pfand aus. Die Abrechnung zeigt 16,00 € abzugeben, der Kassenbericht 1 Rückgabe mit 4,00 €.

### Acceptance criteria

- [ ] Eine Rückgabe am Tisch mindert den abzugebenden Betrag des Akteurs, auch ohne Bestellung am Tisch (`TestAbrechnung_PfandrueckgabeAmTischMindertAbzugeben` in `backend/api/reporting/application/abrechnung_integration_test.go`): `TEST_PKG=./api/reporting/application/ TEST_RUN='^TestAbrechnung_PfandrueckgabeAmTischMindertAbzugeben$' make check-integration | grep -- '--- PASS: TestAbrechnung_PfandrueckgabeAmTischMindertAbzugeben '`
- [ ] Eine Rückgabe im Direktverkauf ändert keine Servicekraft-Zeile (`TestAbrechnung_PfandrueckgabeImDirektverkaufBleibtAussen`, gleiche Datei): `TEST_PKG=./api/reporting/application/ TEST_RUN='^TestAbrechnung_PfandrueckgabeImDirektverkaufBleibtAussen$' make check-integration | grep -- '--- PASS: TestAbrechnung_PfandrueckgabeImDirektverkaufBleibtAussen '`
- [ ] Eine Stornierung zählt für den Akteur der Rückgabe, nicht für die Serviceleitung (`TestAbrechnung_PfandStornoTrifftDenAkteurDerRueckgabe`, gleiche Datei): `TEST_PKG=./api/reporting/application/ TEST_RUN='^TestAbrechnung_PfandStornoTrifftDenAkteurDerRueckgabe$' make check-integration | grep -- '--- PASS: TestAbrechnung_PfandStornoTrifftDenAkteurDerRueckgabe '`
- [ ] Die Eigene Übersicht zieht ausgezahltes Pfand ab und darf negativ werden (`TestEigeneUebersicht_PfandrueckgabeMindertAbzugeben`, gleiche Datei): `TEST_PKG=./api/reporting/application/ TEST_RUN='^TestEigeneUebersicht_PfandrueckgabeMindertAbzugeben$' make check-integration | grep -- '--- PASS: TestEigeneUebersicht_PfandrueckgabeMindertAbzugeben '`
- [ ] Der Kassenbericht zählt Rückgaben und deren Nettosumme (`TestGetReporting_ZaehltPfandrueckgaben` in `backend/repository/reporting_repo/repo_test.go`): `TEST_PKG=./repository/reporting_repo/ TEST_RUN='^TestGetReporting_ZaehltPfandrueckgaben$' make check-integration | grep -- '--- PASS: TestGetReporting_ZaehltPfandrueckgaben '`
- [ ] Der Kassenbericht zeigt die Kennzahl „Pfandrückgaben" (Fall „shows the Pfandrückgaben figure" in `frontend/src/admin/reporting/ReportingResults.test.tsx`): `cd frontend && pnpm exec vitest run src/admin/reporting/ReportingResults.test.tsx -t 'shows the Pfandrückgaben figure' --reporter=verbose | grep -- '✓ .*shows the Pfandrückgaben figure'`
- [ ] Die Eigene Übersicht zeigt einen negativen abzugebenden Betrag (Fall „shows a negative amount to hand over" in `frontend/src/service/table/EigeneUebersicht.test.tsx`): `cd frontend && pnpm exec vitest run src/service/table/EigeneUebersicht.test.tsx -t 'shows a negative amount to hand over' --reporter=verbose | grep -- '✓ .*shows a negative amount to hand over'`

## Phase 6: DSFinV-K-Export und Demo-Daten

**Depends on**: 2

### Context

- `backend/api/fiskal/dsfinvk/mapper.go — belegeFromEvents()`, `beleg`, `abrechnungskreis()`, `gvTypReihenfolge`, `buildBusinesscases()`.
- `backend/api/fiskal/dsfinvk/mapper_test.go — TestMapDirektverkaufUndStorno()`, `TestMapWarenruecknahmeNegativeWithZahlungReference()` — Vorbild Golden Rows.
- `backend/api/reporting/application/query_export_konsistenz_test.go — TestUmsatzProSteuersatz_KonsistentMitDSFinVKBusinesscases()`.
- `backend/dsfinvkpruefung/inhalt.go — pruefeStornoReferenzen()` — prüft nur `GV_TYP` „Umsatz"; Pfand-Bons lösen die Regel nicht aus.
- `backend/dsfinvkpruefung/pruefung_integration_test.go — TestSeedExportBefundfrei()`.
- `backend/seed/szenario.go — demoSzenario()`, `direktverkauf`, `direktverkaufStorno`, `samstagsAktionen()`.
- `backend/seed/engine.go — sitzungsBuilder.process()`, `tagesSummen.UmsatzGesamtCents()`, `kassensturz()`.
- `backend/seed/writer.go — writeStammdaten()`, `backend/sqlc/queries/seed.sql — SeedInsertProdukt`.
- `backend/seed/engine_test.go` — Invarianten der Seed-Engine.
- `docs/compliance.md` §3.3, §6.3, §6.6, §6.8.
- `docs/verfahrensdokumentation.md` §2, §3, §4.
- `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/20231215_DSFinV_K_2_4.pdf` — Anhang C S. 50, Tz. 4.2.4, Tz. 4.2.5.

### What to build

Export:

- Konstante `gvTypPfandRueckzahlung = "PfandRueckzahlung"`, in `gvTypReihenfolge` direkt nach „Umsatz".
- `belegeFromEvents()`: Rückgabe als `beleg` mit `bonID = rueckgabeId`, `BON_TYP` „Beleg", Zahlart Bar und `storno: true` für das negative Vorzeichen. Ohne Referenz.
- Stornierung als positiver `beleg` mit `bonID = stornierungId` und `refBonIDs = [rueckgabeId]`.
- `abrechnungskreis()` bekommt eine Variante über die Tisch-ID; 0 ergibt keinen Abrechnungskreis.
- Die Rückgabe nimmt `tischId` aus ihrem Event. `pfandrueckgabe-storniert:v1` trägt keine `tischId`. Deshalb führt `belegeFromEvents()` beim Durchlauf eine Map `rueckgabeId → tischId`, wie `herkunft` für Bestellungen.
- `buildBusinesscases()` summiert „PfandRueckzahlung" je USt-Schlüssel ohne weitere Änderung.

Demo-Daten:

- `SeedInsertProdukt` schreibt `pfand`; `make sqlc` erneuert `backend/sqlc/dbgen/seed.sql.go`. Das Szenario bekommt ein Pfand-Produkt „Pfand", Kategorie getraenk, Steuersatz regel, Variante „Glas" zu 2,00 €.
- Aktionen `pfandrueckgabe{RueckgabeID, Tisch, User, Posten, Kommentar}` (Tisch 0 = Theke) und `pfandStorno{RueckgabeID, StornierungID, User, Posten, Kommentar}`.
- Die Engine bucht sie über die Domain-Konstruktoren, mindert `bestandCents` und führt sie in `tagesSummen`. `UmsatzGesamtCents()` zieht sie ab.
- Der Samstag bekommt Pfand-Verkäufe, eine Rückgabe am Tisch, eine an der Theke und eine Stornierung.
- Literale in `backend/seed/engine_test.go` und `backend/seed/bondruck_test.go`, die Szenario-Zählungen pinnen, ziehen nach.

Doku `docs/compliance.md`:

- Prozesstyp-Tabelle: Pfandrückgabe und deren Stornierung als `Kassenbeleg-V1`.
- Bontypen und Geschäftsvorfälle: „PfandRueckzahlung".
- Storno-Handling: Rückgabe ohne Referenz, Stornierung mit Referenz.
- Die Zeile zu `subitems.csv` nennt Pfand nicht mehr als Beispiel.

Doku `docs/verfahrensdokumentation.md`:

- §2, Bounded-Context-Tabelle: Pfandrückgabe unter den Geschäftsvorfällen der Kasse.
- §3, „Strukturierung (Subjects)": jede Pfandrückgabe bildet einen eigenen Strom, wie ein Theken-Direktverkauf.
- §4, Prozesstyp-Tabelle: Pfandrückgabe und deren Stornierung als `Kassenbeleg-V1`.

Demo: Den Samstag exportieren. `lines.csv` hat Zeilen „PfandRueckzahlung" mit negativer Menge, `businesscases.csv` eine Summe je Satz.

### Acceptance criteria

- [ ] Rückgabe und Stornierung ergeben die erwarteten Zeilen in Bonkopf, Positionen, Referenzen, Abrechnungskreis und Geschäftsvorfällen, auch den Tischnamen der Stornierung (`TestMapPfandrueckgabeUndStorno` in `backend/api/fiskal/dsfinvk/mapper_test.go`): `cd backend && go test ./api/fiskal/dsfinvk/ -run '^TestMapPfandrueckgabeUndStorno$' -v | grep -- '--- PASS: TestMapPfandrueckgabeUndStorno '`
- [ ] Umsatz je Steuersatz und „Umsatz" plus „PfandRueckzahlung" stimmen je USt-Schlüssel überein (`TestUmsatzProSteuersatz_KonsistentMitPfandrueckzahlung` in `backend/api/reporting/application/query_export_konsistenz_test.go`): `cd backend && go test ./api/reporting/application/ -run '^TestUmsatzProSteuersatz_KonsistentMitPfandrueckzahlung$' -v | grep -- '--- PASS: TestUmsatzProSteuersatz_KonsistentMitPfandrueckzahlung '`
- [ ] Das Szenario enthält Rückgaben und eine Stornierung, und die Seed-Engine hält Soll, Z-Bon-Summe und Tisch-Salden ein: `cd backend && grep -q 'pfandrueckgabe{' seed/szenario.go && grep -q 'pfandStorno{' seed/szenario.go && go test ./seed/`
- [ ] Der Export aller Seed-Sitzungen ist befundfrei, samt Pfand-Bons (`TestSeedExportBefundfrei`): `TEST_PKG=./dsfinvkpruefung/ TEST_RUN='^TestSeedExportBefundfrei$' make check-integration | grep -- '--- PASS: TestSeedExportBefundfrei '`
- [ ] Prosa-Gate grün: `./scripts/check-prose.sh`

## Phase 7: Ende-zu-Ende und Dokumentation

**Depends on**: 3, 4, 5, 6

### Context

- `e2e/tests/direktverkauf-storno.mobile.spec.ts` — Vorbild am Handy-Viewport.
- `e2e/support/seed.ts — resetAndSeed()`, `e2e/support/servicekraft.ts — waehleVariante()`, `e2e/support/anmelden.ts — anmelden()`.
- `docker-compose.e2e.yml` — eigener E2E-Stack, gestartet und abgewartet wie im CI-Job `e2e` in `.github/workflows/ci.yml`.
- `Makefile — website-screenshots`, `test-e2e` — Vorbild für Stack-Start, Warteschleife und Abbau per `trap`.
- `frontend/src/admin/kasse/KassensitzungPage.tsx` — zeigt den Soll-Kassenbestand.
- `docs/handbuch.md` §3.3, §3.6, §3.8, §3.9, §4.1, §4.5, §5.1, §7.1, §7.2.
- `docs/anforderungen.md` — Kasse-Tabelle, nächste freie Kennung K-27.
- `docs/language.md` — Abschnitt Kasse.
- `docs/prds/prd-pfandrueckgabe.md` — Glossar-Einträge unter „Further Notes".

### What to build

- Make-Target `test-e2e-stack`: startet `docker-compose.e2e.yml` per `up -d --build` auf `E2E_STACK_PORT` (Standard 8080) unter eigenem Projektnamen. Es wartet wie `website-screenshots` auf HTTP 200 und scheitert nach 60 Versuchen.
- Danach läuft `playwright test $(E2E_ARGS)` mit `E2E_BASE_URL` auf den Stack. Ein `trap` ruft `down -v`, der Exit-Code ist der des Testlaufs.
- `e2e/tests/pfandrueckgabe.mobile.spec.ts`: Eine Servicekraft verkauft im Direktverkauf Bier mit Pfand. Danach nimmt sie das Glas über „Pfand zurück" an.
- Der Admin liest den Soll-Kassenbestand vor und nach dem Ablauf. Die Differenz ist Verkauf minus Pfand.
- `docs/handbuch.md`: Subject `kassensitzung-{nr}/pfandrueckgabe-{rueckgabeId}`, Event-Tabelle „Pfand-Events" und Soll-Formel mit Pfand.
- §3.8, Stream-Typ-Tabelle: Zeile `"pfandrueckgabe"` ohne Projektion, wie `"direktverkauf"`.
- Ebenso das Pfand-Kennzeichen am Produkt-Aggregat, die Belegformen und Berechtigungszeilen.
- Und die Invariante „abzugeben = kassiert − Rücknahmen − Pfand, darf negativ sein".
- `docs/anforderungen.md`: K-27 „Pfandrückgabe" mit Verrechnung und Stornierung.
- `docs/language.md`: Pfand-Produkt, Pfandrückgabe, Verrechnung, Pfandrückgabe-Stornierung mit den zu vermeidenden Begriffen. „Gutschrift" bleibt aus Code und Oberfläche fern.

Demo: Die Spec läuft grün gegen den E2E-Stack.

### Acceptance criteria

- [ ] Verkauf mit Pfand, Rückgabe und Soll-Kassenbestand stimmen am Handy-Viewport: `make test-e2e-stack E2E_ARGS=tests/pfandrueckgabe.mobile.spec.ts`
- [ ] Die übrigen Specs überstehen das Pfand-Produkt in den Seed-Daten: `make test-e2e-stack`
- [ ] E2E-Suite formatiert, typisiert und gelintet: `make check-e2e`
- [ ] Doku-Links und Anker gültig: `./scripts/check-links.sh`
- [ ] Prosa-Gate grün: `./scripts/check-prose.sh`
