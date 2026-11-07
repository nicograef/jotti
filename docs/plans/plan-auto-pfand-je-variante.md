# Plan: Auto-Pfand je Variante

> Source PRD: `docs/prds/prd-auto-pfand-je-variante.md` (freigegeben mit Entscheidung 2: Abholbon ohne Pfand-Positionen in jedem Bonmodus, selber Release; Entscheidung 3: Regel „gleicher Steuersatz" ohne Kombi-Sonderfall)

## Goal

Der Admin verknüpft eine Variante mit einer Pfand-Variante gleichen Steuersatzes.
Plus auf diese Variante legt im Warenkorb je Stück ein Pfand als gewöhnliche Position dazu, am Tisch und im Direktverkauf.
Events, Bestell-API, Zahlung, Stornierung, Umbuchung, Pfandrückgabe und Export bleiben unverändert.
Der Abholbon des Direktverkaufs lässt Pfand-Positionen in jedem Bonmodus weg.

## Architectural decisions

- **Schema:** Eine additive Migration `17_pfand_verknuepfung.up.sql` (Nummer: nächste freie ab 17; landet #145 vorher mit eigener Migration, wird es 18; der Lead legt die Nummer beim Landen fest).
  - In `BEGIN; … COMMIT;`; der Kopfkommentar übernimmt aus Migration 09 den Satz, dass die Feldliste der View `jsonVariante` spiegelt.
  - `ALTER TABLE produkt_varianten ADD COLUMN pfand_variante_id INT REFERENCES produkt_varianten(id)`, nullable, mit `COMMENT ON COLUMN`.
  - `CREATE OR REPLACE VIEW produkt_varianten_json` mit zusätzlichem Schlüssel `'pfandVarianteId', pfand_variante_id`; Spaltenliste der View bleibt gleich.
  - Kein Index, kein CHECK: Dutzende Varianten je Verein, Selbstverweis schließt die Domänenregel aus. Kein Datenschreiben.
- **Choke files nur in Phase 2:** Migration, `backend/sqlc/queries/produkte.sql` und `backend/sqlc/dbgen/` (per `make sqlc`). Auch die Lösch-Query `EntfernePfandVerweise` entsteht dort, Phase 3 nutzt sie nur.
- **Domäne:** `produkt.Variante.PfandVarianteID *int` (nil = kein Pfand), ohne json-Tag.
  - `produkt.PruefePfandVerknuepfung(quelle Produkt, ziel Produkt, zielVarianteID int) error` ist die reine Regel.
  - Reihenfolge der Prüfung: Ziel-Variante fehlt, ist gelöscht oder gehört zu keinem Pfand-Produkt → `ErrPfandZielUngueltig`; Quelle ist Pfand-Produkt → `ErrPfandVerkettet`; Steuersatz verschieden → `ErrPfandSteuersatzAbweichend`.
  - `produkt.PruefePfandVerknuepfungen(katalog []Produkt, produktID int) error` (Phase 3) prüft jede Verknüpfung, die von Varianten des Produkts ausgeht oder auf sie zeigt.
  - Verliert das Ziel-Produkt dabei sein Pfand-Kennzeichen, meldet sie `ErrPfandNochVerknuepft` (Phase 3) statt `ErrPfandZielUngueltig`.
- **Fehlercodes (400), Texte zentral in `frontend/src/lib/errorMessages.ts`:**

  | Code                          | Meldung                                                                                                     |
  | ----------------------------- | ----------------------------------------------------------------------------------------------------------- |
  | `pfand_ziel_ungueltig`        | Das gewählte Pfand ist keine Pfand-Variante mehr. Bitte neu laden und erneut wählen.                        |
  | `pfand_verkettet`             | Ein Pfand-Produkt bringt selbst kein Pfand mit. Bitte zuerst die Pfand-Verknüpfungen seiner Varianten entfernen. |
  | `pfand_steuersatz_abweichend` | Pfand und Variante brauchen denselben Steuersatz.                                                          |
  | `pfand_noch_verknuepft`       | Varianten verknüpfen dieses Produkt als Pfand. Bitte zuerst ihre Pfand-Verknüpfung entfernen.              |

- **Anwendungs-Sentinels** sind Aliase der Domänenfehler in `application/errors.go`, wie `ErrDatabase = db.ErrDatabase`; der Vertragstest der Fehlerabbildung erzwingt ihre Nennung im HTTP-Layer.
- **Prüfung beim Schreiben:** Variante anlegen und ändern mit Verweis sowie Produkt ändern laden den Admin-Katalog (`GetAllProdukte`) vollständig und prüfen in Go. Keine Sperre gegen zwei gleichzeitig schreibende Admins.
- **API:** `admin/create-variante` und `admin/update-variante` nehmen `pfandVarianteId` (`*int`, `z.Ptr(dom.IDSchema)`; null oder fehlend = kein Pfand). Das Frontend sendet das Feld immer.
- **Lesepfade:** Die Varianten-DTO trägt `pfandVarianteId` (Zahl oder null) in `admin/get-all-produkte` und `service/get-aktive-produkte`.
  - Admin: gespeicherter Verweis, auch auf ein inaktives Ziel.
  - Service: `application.Query.GetActiveProdukte` nullt jeden Verweis, dessen Ziel nicht als Variante derselben Antwort erscheint. Die Anwendungsschicht formt diese Antwort schon heute (`PfandProdukte`).
- **Schreibfalle:** `UpdateVariante` schreibt `pfand_variante_id` immer mit. Jede Variante, die in einen Schreibpfad geht, stammt aus `GetVariante` oder `GetProdukt` (beide lesen den Verweis). `GetVariantenByIDs` liest ihn nicht und speist nie einen Schreibpfad.
- **Löschen:** Variante löschen und Produkt löschen entfernen in derselben Transaktion alle Verweise auf die gelöschten Varianten (`EntfernePfandVerweise`, setzt `updated_at`). Deaktivieren lässt Verweise stehen.
- **Frontend-Admin:** Auswahl und Lösch-Hinweis leitet das Frontend aus der Admin-Produktliste ab (reine Funktionen in `admin/products/pfandVerknuepfung.ts`). `Products` reicht die Liste als `katalog` an `ProductItem`, `VariantChip`, `NewVariantDialog` und `EditVariantDialog` weiter.
- **Frontend-Service:** Reine Kopplung `mitPfand()` in `service/pfandKopplung.ts`; `useMengen().setAll` nimmt zusätzlich eine Updater-Funktion, damit Plus und Minus funktional auf den aktuellen Stand wirken. Kein globaler Store.
- **Abholbon:** `createAbholbonAuftraege` filtert Pfand-Positionen vor der Bonmodus-Verzweigung; eine leere Restliste ergibt keinen Auftrag. Der Rückfall ohne Abholbon-IP an die Produktstationen bleibt unverändert.
- **Parallelität:** Phasen 1, 2 und 3 teilen `docs/handbuch.md` und laufen nacheinander. Phase 4 berührt keine Datei der Phasen 1 und 3 und darf neben jeder von ihnen laufen.
- **Gates:** Backend-Teil: `make check check-repo check-format && make lint && TEST_PG_PORT=5447 make check-integration`.
  - Phasen 2 und 3 zusätzlich `E2E_STACK_PROJECT=jotti-autopfand-e2e E2E_STACK_PORT=8091 make test-e2e`.
  - Phase 4 nutzt einen eigenen E2E-Stack (`jotti-autopfand-e2e-4`, Port 8092), weil sie neben Phase 3 läuft.
  - Ein `-run`- oder `-t`-Filter ohne Treffer läuft grün durch: Ein Häkchen braucht den genannten Test in der Ausgabe (`-v` bzw. Vitest-Bericht).

## Inventory

- `database/migrations/09_produkt_varianten_json.up.sql` — View `produkt_varianten_json`; die neue Migration ersetzt sie.
- `backend/sqlc/queries/produkte.sql` — `GetVariante`, `CreateVariante`, `UpdateVariante`; Kommentar „Migration 09" an `GetProdukt`.
- `backend/domain/produkt/variante.go` — `Variante`, `VarianteSchema`, `NewVariante()`, `UpdateDetails()`.
- `backend/domain/produkt/produkt.go` — `Produkt.Pfand`, `Produkt.Steuersatz`.
- `backend/repository/produkt_repo/types.go` — `jsonVariante`, `toDomain()`, `varianteRowToDomain()`.
- `backend/repository/produkt_repo/variant.go` — `CreateVariante()`, `UpdateVariante()`.
- `backend/repository/produkt_repo/repo.go` — `DeleteProduktMitVarianten()`, `updateVarianteParams()`.
- `backend/repository/produkt_repo/batch.go` — `GetVariantenByIDs()`: liest den Verweis nicht, kein Schreibpfad.
- `backend/repository/repotest/produkt.go` — `ProduktRepo`: Fake der Anwendungstests.
- `backend/api/stammdaten/produkt/application/command.go` — `CreateVariante()`, `UpdateVariante()`, `UpdateProdukt()`, `DeleteVariante()`, `DeleteProdukt()`.
- `backend/api/stammdaten/produkt/application/query.go` — `GetActiveProdukte()`, `AktiveProdukte`.
- `backend/api/stammdaten/produkt/http/command_handler.go` — Request-Structs, zog-Schemas, `helper.MapError`-Tabellen.
- `backend/api/stammdaten/produkt/http/query_handler.go` — DTO `variante`, `toVariante()`.
- `backend/api/error_mapping_contract_test.go` — `TestErrorMappingContract`: erzwingt einen Code je Sentinel.
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go` — `createDirektverkaufAuftraege()`, `createAbholbonAuftraege()`, `unmarshalPositionenMitKommentar()`.
- `backend/domain/kasse/bestellung.go` — `PositionV2EventData`, `PositionFromEventData()`.
- `backend/seed/bondruck_test.go` — verlangt genau einen Sammel-Abholbon je Direktverkauf; das Drehbuch verkauft im Direktverkauf kein Pfand.
- `frontend/src/lib/produktSchemas.ts` — `VarianteSchema` (geteilt von beiden Lesepfaden).
- `frontend/src/admin/products/ProduktBackend.ts` — `CreateVarianteSchema`, `UpdateVarianteSchema`.
- `frontend/src/admin/products/` — `Products.tsx`, `ProductItem.tsx`, `VariantChip.tsx`, `NewVariantDialog.tsx`, `EditVariantDialog.tsx`, `SteuersatzField.tsx` (Vorbild der Auswahl).
- `frontend/src/hooks/use-mengen.ts` — `useMengen()`, `setAll()`.
- `frontend/src/service/ProductList.tsx` — `ProductList`, `VariantRow`.
- `frontend/src/service/table/BestellungTab.tsx`, `frontend/src/service/direktverkauf/DirektverkaufTab.tsx` — `onAdd`/`onRemove` der Produktliste.
- `frontend/src/service/drawerUtils.ts` — `toBestellungData()`: macht das Pfand ohne Änderung zur eigenen Position.
- `docs/handbuch.md` — §4.1 Produkt-Aggregat, §4.5 Direktverkauf-Routing.
- `docs/language.md` — Abschnitt „Pfand".
- `docs/leitfaden/belege-steuersaetze.md` — Pfand-Absatz.
- `docs/decisions.md` — neue Entscheidungszeile.

## Open questions / Risks

1. **Frontend leitet Pfand-Auswahl und Lösch-Hinweis ab.** AGENTS.md: das Backend filtert und formt. Die Admin-Liste trägt alle nötigen Daten; das Backend erzwingt die Regeln trotzdem. Alternative: ein Backend-Feld `pfandOptionen` je Produkt. Empfehlung: Ableitung im Frontend, wie `groupProdukteByKategorie`.
2. **Service-Filter in der Anwendungsschicht statt im Repository.** Das PRD testet den Service-Lesepfad gegen Postgres. Der Plan filtert in `Query.GetActiveProdukte` und testet dort mit dem Fake. Den Feldweg View → Domäne pinnt ein Repository-Integrationstest. Empfehlung: so übernehmen.
3. **Direktverkauf ohne Abholbon-IP** druckt Pfand weiter an der Station seiner Kategorie; Entscheidung 2 betrifft nur den Abholbon. Empfehlung: unverändert lassen, gleiche Regel wie am Tisch (Pfand-Produkt in eine Kategorie ohne Druckstation legen).
4. **Demo-Drehbuch:** Der Seed bekommt keine Verknüpfung. Eine Verknüpfung änderte die E2E-Ausgangslage nicht, aber jede Spec, die eine verknüpfte Variante tippt. Empfehlung: Seed unverändert (D14).
5. **CHANGELOG:** Kein Abschnitt `[Unreleased]` existiert. Empfehlung: Eintrag schreibt der Release-Schritt, nicht dieser Plan.
6. **Nummern:** Migrationsnummer (nächste freie ab 17) und die D-Nummer der Entscheidungszeile vergibt der Lead beim Landen; #145 fügt ebenfalls eine Zeile in `docs/decisions.md` an.
7. **Risiko #145 (landet vorher):** gibt `direktverkauf-getaetigt:v2` eine eigene Positionsform mit `pfand` und `kommentar` und ändert `positionenMitKommentarEventData` sowie `arbeitsbon_policy_test.go`. Phase 1 baut auf dieser Form auf und wird nach dem Landen von #145 rebased.
8. **Risiko Zusatzdrucker (landet danach):** schreibt das Direktverkauf-Routing in `arbeitsbon_policy.go` um. Sein Plan muss den Pfand-Filter des Abholbons übernehmen.
9. **Risiko `fix/q-cross`:** fügt Idempotenzschlüssel für Tisch-/Direktverkauf-Storno und Umbuchung hinzu. Überschneidung: `frontend/src/hooks/use-mengen.ts` (Storno- und Umbuchungs-Drawer nutzen `useMengen`), `frontend/src/lib/errorMessages.ts` (neue Codes) und die `DirektverkaufTab`-Tests. Der Kassenbeleg mit TSE-Start- und Endzeit berührt diesen Plan nicht. Die Feature-Spur startet erst nach dem Programm-Abschluss; jede Phase baut auf dem Stand von main mit `fix/q-cross` auf und rebased bei Konflikt dorthin.
10. **Bekannte Unschärfe (PRD):** Nach dem Abzug für ein eigenes Glas zieht ein weiteres Minus am Getränk noch ein Pfand ab; die Vorschau zeigt das Ergebnis. Empfehlung: hinnehmen wie im PRD; eine Zuordnung Pfand je Stück ist Out of Scope.

## Phase 1: Abholbon ohne Pfand-Positionen

**Depends on**: none

### Context

- `backend/api/druck/bondruck/application/arbeitsbon_policy.go — unmarshalPositionenMitKommentar()` — dekodiert heute `[]kasse.PositionEventData` und verliert das Pfand-Kennzeichen.
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go — createDirektverkaufAuftraege(), createAbholbonAuftraege()` — Routing und Bonmodus-Verzweigung des Abholbons.
- `backend/domain/kasse/bestellung.go — PositionV2EventData, PositionFromEventData()` — Positionsform mit `pfand`; v1-Events lesen sich ohne Schlüssel als `false`.
- `backend/api/druck/bondruck/application/arbeitsbon_policy_test.go — TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProBestellung()` und Geschwister — Vorbild der neuen Fälle.
- `backend/seed/bondruck_test.go` — bleibt grün, weil kein Seed-Direktverkauf Pfand enthält.
- `docs/handbuch.md` — §4.5, Absatz „Direktverkauf-Routing".

### What to build

Ein Direktverkauf mit Bier und Glaspfand druckt am Abholbon nur das Bier, in jedem Bonmodus.
Die Policy dekodiert die Positionen mit Pfand-Kennzeichen (`kasse.PositionV2EventData`, nach #145 dessen Direktverkauf-Positionsform) und setzt `kasse.Position.Pfand`.
`createAbholbonAuftraege()` filtert Positionen mit `Pfand` vor der Bonmodus-Verzweigung; bleibt nichts übrig, entsteht kein Auftrag.
Tischbestellungen und der Rückfall an die Produktstationen bleiben unverändert.
`docs/handbuch.md` §4.5 nennt im Abholbon-Punkt: Pfand-Positionen erscheinen auf keinem Abholbon.

### Acceptance criteria

- [ ] Neuer Tabellentest `TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_OhnePfand` über `pro_bestellung`, `pro_position` und `pro_stueck`: Bier ×2 plus Pfand ×2 ergibt nur Abholbons mit Bier, ein reiner Pfand-Verkauf keinen Auftrag. Verify: `cd backend && go test ./api/druck/bondruck/application/ -run 'TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_OhnePfand' -v`
- [ ] Handbuch §4.5 nennt die Pfand-Regel des Abholbons. Verify: `grep -n 'Pfand-Positionen' docs/handbuch.md`
- [ ] Gate grün. Verify: `make check check-repo check-format && make lint && TEST_PG_PORT=5447 make check-integration`

## Phase 2: Pfand-Verknüpfung einrichten

**Depends on**: none

### Context

- `database/migrations/09_produkt_varianten_json.up.sql` — Feldliste der View, die `jsonVariante` spiegelt.
- `backend/sqlc/queries/produkte.sql — GetVariante, CreateVariante, UpdateVariante, GetProdukt` — Spalte ergänzen, Kommentar „Migration 09" auf die neue Migration umstellen.
- `backend/domain/produkt/variante.go — Variante, VarianteSchema, NewVariante(), UpdateDetails()` — Feld `PfandVarianteID`, Regel-Funktion und drei Domänenfehler.
- `backend/domain/produkt/variante_test.go — TestNewVariante_Preisgrenze()` — Vorbild des Tabellentests.
- `backend/repository/produkt_repo/types.go — jsonVariante, toDomain(), varianteRowToDomain()` — Verweis lesen; Kommentar auf die neue Migration umstellen.
- `backend/repository/produkt_repo/variant.go — CreateVariante()` und `backend/repository/produkt_repo/repo.go — updateVarianteParams()` — Verweis schreiben.
- `backend/repository/produkt_repo/repo_test.go — TestUpdateVariante()` — Vorbild des Integrationstests.
- `backend/api/stammdaten/produkt/application/command.go — CreateVariante(), UpdateVariante(), produktRepo` — Prüfung mit Katalog; Interface um `GetAllProdukte` erweitern.
- `backend/api/stammdaten/produkt/application/errors.go` — Aliase `ErrPfandZielUngueltig`, `ErrPfandVerkettet`, `ErrPfandSteuersatzAbweichend`.
- `backend/api/stammdaten/produkt/application/query.go — GetActiveProdukte()` — Service-Filter.
- `backend/api/stammdaten/produkt/application/query_test.go — TestGetActiveProdukte_PfandProdukteNurMitKennzeichen()` — Vorbild.
- `backend/api/stammdaten/produkt/application/command_test.go` — Fake `repotest.NewProduktRepo`.
- `backend/api/stammdaten/produkt/http/command_handler.go — createVarianteRequest, updateVarianteRequest, CreateVarianteHandler(), UpdateVarianteHandler()` — Feld, Schema, Fehlercodes.
- `backend/api/stammdaten/produkt/http/command_handler_test.go — TestCommandHandler_Abweisungen()` — Vorbild.
- `backend/api/stammdaten/produkt/http/query_handler.go — variante, toVariante()` — DTO-Feld `pfandVarianteId`.
- `frontend/src/lib/produktSchemas.ts — VarianteSchema` — `pfandVarianteId: VarianteIdSchema.nullable()`.
- `frontend/src/admin/products/ProduktBackend.ts — CreateVarianteSchema, UpdateVarianteSchema` — Feld ergänzen.
- `frontend/src/admin/products/SteuersatzField.tsx` — Vorbild der Auswahl (Radix Select: „Kein Pfand" braucht einen nicht leeren Sentinel-Wert).
- `frontend/src/admin/products/Products.tsx — Products` und `frontend/src/admin/products/ProductItem.tsx — ProductItem` — Katalog als `katalog` durchreichen.
- `frontend/src/admin/products/VariantChip.tsx — VariantChip` — Hinweis „+ Pfand“, Katalog an den Dialog.
- `frontend/src/admin/products/NewVariantDialog.tsx — NewVariantDialog` und `frontend/src/admin/products/EditVariantDialog.tsx — EditVariantDialog` — Feld „Pfand“.
- `frontend/src/admin/products/Products.test.tsx` — Vorbild der Admin-Tests.
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — drei Codes.
- Variante-Fixtures in `frontend/src/test/fixtures.ts`, `frontend/src/admin/products/Products.test.tsx`, `frontend/src/admin/products/productGrouping.test.ts`, `frontend/src/service/drawerUtils.test.ts`, `frontend/src/service/table/TablePage.test.tsx`, `frontend/src/service/table/ZahlungAbschluss.test.tsx`, `frontend/src/service/direktverkauf/DirektverkaufPage.test.tsx`, `frontend/src/service/direktverkauf/DirektverkaufAbschluss.test.tsx`, `frontend/src/service/direktverkauf/DirektverkaufTab.test.tsx` — bekommen `pfandVarianteId: null`, soweit sie Varianten literal bauen.
- `docs/handbuch.md` — §4.1 Produkt-Aggregat.

### What to build

Der Admin wählt im Variantendialog ein Pfand, das Backend prüft und speichert den Verweis, beide Lesepfade liefern ihn.

- Migration `database/migrations/17_pfand_verknuepfung.up.sql` wie in den Architekturentscheidungen; danach `make sqlc`.
- Die Queries `GetVariante`, `CreateVariante` und `UpdateVariante` lesen und schreiben `pfand_variante_id`.
- Neue Query `EntfernePfandVerweise :exec` (`UPDATE produkt_varianten SET pfand_variante_id = NULL, updated_at = $1 WHERE pfand_variante_id = ANY($2::int[])`); Phase 3 ruft sie auf.
- Domäne: Feld, `NewVariante(name, preisCents, pfandVarianteID *int)`, `UpdateDetails(name, preisCents, pfandVarianteID *int)`.
- Domäne: `PruefePfandVerknuepfung()` und die drei Fehler `ErrPfandZielUngueltig`, `ErrPfandVerkettet`, `ErrPfandSteuersatzAbweichend`.
- Repository: `jsonVariante.PfandVarianteID *int` mit `json:"pfandVarianteId"`, `varianteRowToDomain()`, `CreateVariante()`, `updateVarianteParams()`.
- Anwendung: `CreateVariante(ctx, produktID, name, preisCents, pfandVarianteID *int)` und `UpdateVariante(ctx, varianteID, name, preisCents, pfandVarianteID *int)`.
  - Mit Verweis laden beide den Katalog per `GetAllProdukte` und suchen Quell- und Zielprodukt.
  - Danach rufen sie `PruefePfandVerknuepfung()`; ein Regelbruch schreibt nichts.
- Anwendung: `Query.GetActiveProdukte()` nullt Verweise, deren Ziel-ID nicht unter den Varianten der Antwort steht.
- HTTP: Feld `PfandVarianteID *int` mit `z.Ptr(dom.IDSchema)` in beiden Requests; die Codes `pfand_ziel_ungueltig`, `pfand_verkettet`, `pfand_steuersatz_abweichend` in beiden `MapError`-Tabellen.
- HTTP: DTO `variante.PfandVarianteID *int` mit `json:"pfandVarianteId"`.
- Frontend: Schemas, neue Komponente `frontend/src/admin/products/PfandVarianteField.tsx` mit Label „Pfand".
- Frontend: reine Funktion `pfandOptionen(katalog, produkt)` in neuer Datei `frontend/src/admin/products/pfandVerknuepfung.ts`.
  - Sie liefert die Varianten aller Pfand-Produkte mit dem Steuersatz des Produkts.
  - Label `<Produkt> <Variante> · <Preis>` über `formatPositionName` und `formatEuro`, Zusatz „(inaktiv)" bei inaktiver Variante oder inaktivem Produkt.
- Frontend: Das Feld erscheint in `NewVariantDialog` und `EditVariantDialog` nur, wenn das Produkt kein Pfand-Produkt ist und `pfandOptionen` nicht leer ist. Erste Option „Kein Pfand".
- Frontend: `VariantChip` zeigt bei gesetztem `pfandVarianteId` den Hinweis „+ Pfand".
- `errorMessages.ts`: die drei Codes mit den Texten der Tabelle oben.
- `docs/handbuch.md` §4.1: Absatz „Pfand-Verknüpfung" mit Spalte, JSON-Feld, Regeln, Fehlercodes und Service-Filter.

### Acceptance criteria

- [ ] Migration läuft auf frischer DB, sqlc ist synchron. Verify: `make check-sqlc && TEST_PG_PORT=5447 TEST_PKG='./repository/produkt_repo/' make check-integration`
- [ ] Tabellentest `TestPruefePfandVerknuepfung`: gültig; Ziel ohne Pfand-Kennzeichen; Ziel-Variante gelöscht; Quelle ist Pfand-Produkt; `regel` gegen `ermaessigt`; `kombi` gegen `kombi` gültig; `kombi` gegen `regel` abgelehnt. Verify: `cd backend && go test ./domain/produkt/ -run TestPruefePfandVerknuepfung -v`
- [ ] Integrationstest `TestPfandVerknuepfung_Roundtrip`: Der Verweis kommt aus `GetProdukt`, `GetAllProdukte` und `GetVariante` zurück und übersteht Deaktivieren und Aktivieren der verknüpfenden Variante über `UpdateVariante()`. Verify: `TEST_PG_PORT=5447 TEST_PKG='./repository/produkt_repo/' TEST_RUN='^TestPfandVerknuepfung' make check-integration`
- [ ] `CreateVariante` und `UpdateVariante` lehnen einen Steuersatzbruch mit `ErrPfandSteuersatzAbweichend` ab und schreiben nichts; ein gültiger Verweis wird gespeichert. Verify: `cd backend && go test ./api/stammdaten/produkt/application/ -run 'TestVariantePfandVerknuepfung' -v`
- [ ] Der Service-Lesepfad liefert null für eine inaktive Pfand-Variante, der Admin-Lesepfad ihre ID. Verify: `cd backend && go test ./api/stammdaten/produkt/application/ -run 'TestGetActiveProdukte_PfandVerweis' -v`
- [ ] `create-variante` nimmt Bodies ohne `pfandVarianteId` und mit `null` an und weist `0` mit 400 ab. Verify: `cd backend && go test ./api/stammdaten/produkt/http/ -run 'TestCommandHandler_PfandVarianteId' -v`
- [ ] Jeder neue Sentinel hat einen Fehlercode. Verify: `cd backend && go test ./api/ -run TestErrorMappingContract`
- [ ] Admin, Tests mit „Pfand-Verknüpfung" im Namen: Ohne passende Pfand-Variante fehlt das Feld „Pfand"; mit Pfand-Produkten zu 19 % und 7 % bietet es an einem 19-%-Produkt nur das 19-%-Pfand an. Die Auswahl sendet `pfandVarianteId`, und der Chip zeigt „+ Pfand". Verify: `cd frontend && pnpm exec vitest run src/admin/products/Products.test.tsx -t 'Pfand-Verknüpfung'`
- [ ] Frontend-Typen und Fixtures passen. Verify: `cd frontend && pnpm typecheck`
- [ ] Handbuch §4.1 beschreibt die Pfand-Verknüpfung. Verify: `grep -n 'pfandVarianteId' docs/handbuch.md`
- [ ] Gate grün. Verify: `make check check-repo check-format && make lint && TEST_PG_PORT=5447 make check-integration && E2E_STACK_PROJECT=jotti-autopfand-e2e E2E_STACK_PORT=8091 make test-e2e`

## Phase 3: Verknüpfungen bleiben konsistent

**Depends on**: 2

### Context

- `backend/api/stammdaten/produkt/application/command.go — UpdateProdukt(), DeleteVariante(), DeleteProdukt(), produktRepo` — Katalogprüfung und transaktionales Löschen.
- `backend/api/stammdaten/produkt/application/errors.go` — Alias `ErrPfandNochVerknuepft`.
- `backend/domain/produkt/variante.go — PruefePfandVerknuepfung()` — aus Phase 2; daneben entstehen `PruefePfandVerknuepfungen()` und `ErrPfandNochVerknuepft`.
- `backend/repository/produkt_repo/repo.go — DeleteProduktMitVarianten()` — Verweise in derselben Transaktion entfernen.
- `backend/repository/produkt_repo/variant.go` — neue Methode `DeleteVariante(ctx, v produkt.Variante)`.
- `backend/sqlc/dbgen/produkte.sql.go — EntfernePfandVerweise()` — aus Phase 2.
- `backend/repository/repotest/produkt.go — ProduktRepo` — Fake um `DeleteVariante` ergänzen.
- `backend/repository/produkt_repo/delete_integration_test.go — TestDeleteProduktMitVarianten_HappyPath()` — Vorbild.
- `backend/api/stammdaten/produkt/application/command_test.go — TestUpdateProdukt(), TestDeleteVariante()` — Vorbild.
- `backend/api/stammdaten/produkt/http/command_handler.go — UpdateProduktHandler()` — Codes `pfand_verkettet`, `pfand_steuersatz_abweichend`, `pfand_noch_verknuepft`.
- `frontend/src/admin/products/pfandVerknuepfung.ts` — neue Funktion `verliertPfand(katalog, varianteIds)`.
- `frontend/src/admin/products/ProductItem.tsx` — Lösch-Dialog des Produkts.
- `frontend/src/admin/products/EditVariantDialog.tsx` — Lösch-Dialog der Variante.
- `frontend/src/admin/products/Products.test.tsx` — Vorbild.
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — Code `pfand_noch_verknuepft`.
- `docs/handbuch.md` — §4.1, Absatz „Pfand-Verknüpfung" aus Phase 2.

### What to build

Keine Produktänderung und kein Löschen hinterlässt eine Verknüpfung, die eine Regel bricht.

- Domäne: `PruefePfandVerknuepfungen()` und `ErrPfandNochVerknuepft` wie in den Architekturentscheidungen.
- `UpdateProdukt()` lädt den Katalog, ersetzt darin das geänderte Produkt und ruft `PruefePfandVerknuepfungen(katalog, produktID)`.
  - Ein Steuersatzwechsel mit bestehender Verknüpfung scheitert mit `ErrPfandSteuersatzAbweichend`.
  - Pfand-Kennzeichen aus, solange Varianten auf das Produkt zeigen, scheitert mit `ErrPfandNochVerknuepft`.
  - Pfand-Kennzeichen an, solange eigene Varianten ein Pfand verknüpfen, scheitert mit `ErrPfandVerkettet`.
- `DeleteVariante()` schreibt über `Repository.DeleteVariante()`: Status `deleted` und `EntfernePfandVerweise([id])` in einer Transaktion.
- `DeleteProduktMitVarianten()` ruft in seiner Transaktion `EntfernePfandVerweise` mit allen Varianten-IDs des Produkts.
- Frontend: `verliertPfand(katalog, varianteIds)` meldet, ob eine nicht gelöschte Variante auf eine der IDs zeigt.
  - Trifft das zu, ergänzen beide Lösch-Dialoge den Satz „Verknüpfte Varianten verlieren ihr Pfand."
- Die Produktänderung zeigt die Backend-Meldung über die zentrale Fehlerabbildung.
- `docs/handbuch.md` §4.1 ergänzt die Prüfung bei Produktänderung und das Entfernen der Verweise beim Löschen.

### Acceptance criteria

- [ ] Steuersatzwechsel, Kennzeichen aus und Kennzeichen an scheitern bei bestehender Verknüpfung mit ihrem Sentinel und schreiben nichts; ohne Verknüpfung gelingt jede dieser Änderungen. Verify: `cd backend && go test ./api/stammdaten/produkt/application/ -run 'TestUpdateProdukt_PfandVerknuepfung' -v`
- [ ] `TestDeleteVariante_EntferntPfandVerweise` und `TestDeleteProduktMitVarianten_EntferntPfandVerweise`: Löschen einer Pfand-Variante und eines Pfand-Produkts setzt jeden Verweis darauf auf NULL; ein Verweis auf eine andere Variante bleibt. Verify: `TEST_PG_PORT=5447 TEST_PKG='./repository/produkt_repo/' TEST_RUN='^TestDelete.*_EntferntPfandVerweise$' make check-integration`
- [ ] Jeder neue Sentinel hat einen Fehlercode. Verify: `cd backend && go test ./api/ -run TestErrorMappingContract`
- [ ] Admin: Beide Lösch-Dialoge nennen „Verknüpfte Varianten verlieren ihr Pfand." genau dann, wenn eine Variante auf das Gelöschte zeigt. Verify: `cd frontend && pnpm exec vitest run src/admin/products/Products.test.tsx -t 'verlieren ihr Pfand'`
- [ ] Handbuch §4.1 beschreibt Prüfung und Löschen. Verify: `grep -n 'pfand_noch_verknuepft' docs/handbuch.md`
- [ ] Gate grün. Verify: `make check check-repo check-format && make lint && TEST_PG_PORT=5447 make check-integration && E2E_STACK_PROJECT=jotti-autopfand-e2e E2E_STACK_PORT=8091 make test-e2e`

## Phase 4: Bestellen mit Auto-Pfand

**Depends on**: 2

### Context

- `frontend/src/hooks/use-mengen.ts — useMengen(), MengenSteuerung.setAll` — nimmt zusätzlich einen Updater `(vorher) => nachher`.
- `frontend/src/hooks/use-mengen.test.ts` — Vorbild des Tabellentests.
- `frontend/src/service/ProductList.tsx — ProductList, VariantRow` — Pfand-Hinweis unter dem Preis.
- `frontend/src/service/table/BestellungTab.tsx — BestellungTab` — `onAdd`/`onRemove` koppeln.
- `frontend/src/service/direktverkauf/DirektverkaufTab.tsx — DirektverkaufTab` — dieselbe Kopplung.
- `frontend/src/service/pfand/PfandAbschnitt.tsx` — nutzt `ProductList` für die Pfandrückgabe ohne Kopplung; Pfand-Varianten verknüpfen nichts, der Hinweis erscheint dort nie.
- `frontend/src/service/drawerUtils.ts — toBestellungData()` — unverändert; das Pfand wird darüber zur eigenen Position.
- `frontend/src/service/table/BestellungTab.test.tsx`, `frontend/src/service/direktverkauf/DirektverkaufTab.test.tsx` — Vorbild der Request-Prüfung.
- `docs/language.md` — Abschnitt „Pfand".
- `docs/leitfaden/belege-steuersaetze.md` — Pfand-Absatz.
- `docs/decisions.md` — neue Zeile.

### What to build

Plus auf „Bier 0,5 l" legt im Warenkorb ein Glaspfand dazu, am Tisch wie im Direktverkauf, und der Request trägt beide Positionen.

- Neue Datei `frontend/src/service/pfandKopplung.ts`:
  - `pfandVerknuepfungen(produkte)` liefert `Map<varianteId, { varianteId, preisCents }>` aus den Varianten mit `pfandVarianteId`, deren Ziel in `produkte` steht.
  - `mitPfand(mengen, varianteId, schritt: 'plus' | 'minus', verknuepfungen)` ist rein.
  - Plus erhöht Variante und Pfand um je eins. Minus senkt die Variante um eins und das Pfand um eins, solange es über null liegt; steht die Variante auf null, ändert Minus nichts.
  - Plus und Minus am Pfand selbst ändern nur das Pfand.
- `useMengen().setAll` nimmt einen Wert oder einen Updater.
- `BestellungTab` und `DirektverkaufTab` übergeben `ProductList` die Handler `onAdd={(id) => setAll((m) => mitPfand(m, id, 'plus', verknuepfungen))}` und entsprechend `onRemove`.
- `VariantRow` zeigt bei verknüpfter Variante unter dem Preis „+ 2,00 € Pfand" (Betrag der Pfand-Variante über `formatEuro`); `ProductList` berechnet `pfandVerknuepfungen(props.products)` selbst.
- `docs/language.md`: Einleitungssatz zu Pfand umschreiben (Pfand kommt aus der Verknüpfung oder wird getippt).
  - Dazu zwei Zeilen „Pfand-Variante" und „Pfand-Verknüpfung" mit Code-Mapping: Go `Variante.PfandVarianteID`, JSON/TS `pfandVarianteId`, DB `produkt_varianten.pfand_variante_id`, TS `mitPfand`.
- `docs/leitfaden/belege-steuersaetze.md`: Pfand-Absatz um die Einrichtung ergänzen: Dialoge „Neue Variante anlegen" und „Variante bearbeiten", Feld „Pfand", nur gleicher Steuersatz, Abzug beim eigenen Glas. Kein Kombi-Rat.
- `docs/decisions.md`: eine Zeile (nächste freie D-Nummer) ohne Dateipfad.
  - Inhalt: Auto-Pfand als Stammdaten-Verknüpfung einer Variante auf eine Pfand-Variante gleichen Steuersatzes, ohne Kombi-Sonderfall. Der Warenkorb legt das Pfand sichtbar als gewöhnliche Position dazu; Events, Bestell-API und Export bleiben unverändert.
  - Begründung: Ein Aufschlag am Preis bräche D20, ein Anhängen im Backend versteckte das Pfand bis nach dem Buchen.
  - Dazu: Der Abholbon des Direktverkaufs lässt Pfand-Positionen in jedem Bonmodus weg.

### Acceptance criteria

- [ ] Tabellentest `mitPfand`: Plus; Minus; Boden bei null für Variante und Pfand; zwei Varianten mit einem Pfand summieren sich; Plus und Minus am Pfand von Hand ändern nur das Pfand. Verify: `cd frontend && pnpm exec vitest run src/service/pfandKopplung.test.ts`
- [ ] Am Tisch zeigt die verknüpfte Zeile „+ 2,00 € Pfand", und Plus auf sie sendet Variante und Pfand mit je Menge 1 an `service/bestellung-aufnehmen` (Testname enthält „Auto-Pfand"). Verify: `cd frontend && pnpm exec vitest run src/service/table/BestellungTab.test.tsx -t 'Auto-Pfand'`
- [ ] Im Direktverkauf sendet Plus auf eine verknüpfte Variante Variante und Pfand an `service/direktverkauf-taetigen` (Testname enthält „Auto-Pfand"). Verify: `cd frontend && pnpm exec vitest run src/service/direktverkauf/DirektverkaufTab.test.tsx -t 'Auto-Pfand'`
- [ ] Der Leitfaden beschreibt die Einrichtung, und jedes zitierte UI-Label existiert im Frontend. Verify: `grep -n 'Neue Variante anlegen' docs/leitfaden/belege-steuersaetze.md && ./scripts/check-ui-labels.sh`
- [ ] Glossar und Entscheidung stehen. Verify: `grep -q 'Pfand-Verknüpfung' docs/language.md && grep -q 'Pfand-Verknüpfung' docs/decisions.md`
- [ ] Gate grün. Verify: `make check check-repo check-format && make lint && E2E_STACK_PROJECT=jotti-autopfand-e2e-4 E2E_STACK_PORT=8092 make test-e2e`
