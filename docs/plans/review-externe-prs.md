# Review externer PRs #109, #110, #111

> Quelle: Multi-Linsen-Review (Security/Supply-Chain, Korrektheit, Konventionen, Frontend) je PR,
> jeder Befund von drei Skeptikern gegengeprüft, Synthese mit Merge-Urteil. Stand: PR-Köpfe vom
> 11./12.08.2026 auf Basis `main` @ 2ee9cbaa. 60 Befunde, 35 bestätigt, 25 verworfen.
> Verwendung: Plan 2 Phasen 4, 5 und 10. Der Eigentümer kommentiert die PRs selbst; die
> Kommentar-Entwürfe werden nicht gepostet. Das Dokument hält die Pflicht-Fixes fest: Phase 4
> setzt die zu #110 um, Phase 5 die zu #109; für #111 gilt ADR 10 (abgelehnt).
> Löschen, sobald alle drei PRs gemerged oder geschlossen sind.

## Urteile

| PR   | Urteil            | Security-Risiko | Bestätigte Befunde                |
| ---- | ----------------- | --------------- | --------------------------------- |
| #109 | Änderungen nötig  | none            | 10 (0 Blocker, 4 Major, 6 Minor)  |
| #110 | mergen nach Fixes | none            | 9 (0 Blocker, 6 Major, 3 Minor)   |
| #111 | Änderungen nötig  | none            | 16 (2 Blocker, 12 Major, 2 Minor) |

## PR #109

### Begründung

Supply chain and security are clean: no dependency, CI, Docker, script or dotfile changes; sqlc/dbgen reproduces byte-identically with the pinned sqlc v1.31.1; the three new endpoints sit behind the admin role guard, are POST-only via the shared middleware, are automatically covered by the permission-matrix test, and all values flow through $N parameters. Migration 07 is additive, forward-only, correctly numbered and backfills reihenfolge = id, so existing installs keep their order. But the feature's core operation is broken in a reachable, permanent and silent way: VerschiebeProdukt swaps reihenfolge values instead of ranks, and equal values inside one kategorie are producible by an ordinary admin action, because CreateProdukt numbers per kategorie starting at 1 while UpdateProdukt never reassigns reihenfolge on a kategorie change. Reproduced end to end against PostgreSQL: after moving a produkt into another kategorie the arrows return HTTP 200 and the list stays byte-identical forever, with no error and no UI repair path. The seeder hits the same wall from the other side: SeedInsertProdukt/SeedInsertVariante omit reihenfolge, so every seeded row lands on DEFAULT 0 and reordering is dead in demo, staging and the whole e2e suite. On the frontend the new right chevron in VariantChip overlaps the Switch's invisible 12 px hit-area extender across a 6 px gap, so taps on its left edge toggle the variante active/inactive instead of moving it (reproduced in Chromium via elementFromPoint and mouse.click; adding `relative` to chevronClass removes the dead zone), and both chevrons are 20x20 px while the same PR gives the product-level chevrons size="icon-sm" (32 px). The PR's own category-boundary test is vacuous: it still passes with the kategorie filter removed from GetProduktVorgaenger. None of this is a data-integrity or compliance risk — the kassenjournal, event contracts and DSFinV-K are untouched — but the shipped feature does not reliably work, so it must not merge as is.

### Pflicht-Fixes vor dem Merge

- backend/repository/produkt_repo/repo.go + backend/sqlc/queries/produkte.sql: swap ranks, not raw values — renumber the scope densely with row_number() OVER (ORDER BY reihenfolge, id) before swapping (the pattern SortiereVariantenAlphabetisch already uses), so equal values cannot neutralise the move
- backend/sqlc/queries/produkte.sql UpdateProdukt: when the kategorie changes, reassign reihenfolge to MAX(reihenfolge)+1 of the target kategorie, so a re-categorised produkt lands at the end instead of an arbitrary (or colliding) position
- backend/sqlc/queries/seed.sql: write reihenfolge in SeedInsertProdukt and SeedInsertVariante, so seeded data (demo, staging, e2e, /test/reset-and-seed) is not all on DEFAULT 0
- frontend/src/admin/products/VariantChip.tsx: add `relative` to chevronClass (and to the name button) or raise the spacing next to the Switch to >= 12 px, so taps on the chevron edge no longer toggle the variante
- frontend/src/admin/products/VariantChip.tsx: give the chevrons a real touch target — reuse Button size="icon-sm" (32 px) like the product-level chevrons in ProductItem.tsx
- backend/repository/produkt_repo/reihenfolge_integration_test.go: make TestVerschiebeProdukt_BleibtInSeinerKategorie meaningful (different reihenfolge values across the kategorie boundary, assert the reihenfolge column directly) and add a case with equal values inside one kategorie
- backend/sqlc/queries/produkte.sql: shorten the COLLATE "de-DE-x-icu" comment to the correct rationale (deterministic German ordering independent of the cluster locale) and use a real accent (e.g. "Café Crème") in the collation test
- backend/repository/produkt_repo/mock.go: either add command-level tests that read Verschiebung/ProduktVerschiebungen/VarianteVerschiebungen/SortierteProdukte, or delete the fields together with the comment claiming tests assert on them

### Bestätigte Befunde

| Schwere | Datei                                                             | Befund                                                                                                                                                                                                                                                                                                       |
| ------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| major   | `backend/repository/produkt_repo/repo.go`                         | The reorder swap is a permanent silent no-op whenever two rows in the same kategorie share the same `reihenfolge` value — reachable through the ordinary admin action of changing a produkt's kategorie, because `UpdateProdukt` never reassigns `reihenfolge` while `CreateProdukt` numbers per-kategorie s |
| major   | `backend/repository/produkt_repo/repo.go`                         | Der Tausch vertauscht nur die `reihenfolge`-Werte, nicht den Rang. Haben zwei Nachbarn denselben `reihenfolge`-Wert, ist das Verschieben ein stiller No-Op: HTTP 200, keine Fehlermeldung, unveränderte Liste. Zwei Wege dorthin sind reproduziert.                                                          |
| major   | `frontend/src/admin/products/VariantChip.tsx`                     | The new right chevron ("nach hinten", VariantChip.tsx:99-109) is partly covered by the Switch's invisible hit-area extender, so taps on its left ~6 px toggle the variant active/inactive instead of moving it.                                                                                              |
| major   | `frontend/src/admin/products/VariantChip.tsx`                     | The two chevrons are 20x20 px tap targets with no hit-area extension — far below the 44 px the mobile-first product requires, and the smallest interactive elements in the app.                                                                                                                              |
| minor   | `backend/repository/produkt_repo/repo.go`                         | Two concurrent moves in the same kategorie can produce duplicate `reihenfolge` values, which then wedges that pair permanently via the defect above. The transaction runs at the default READ COMMITTED isolation and the neighbour lookups take no row lock.                                                |
| minor   | `backend/repository/produkt_repo/reihenfolge_integration_test.go` | `TestVerschiebeProdukt_BleibtInSeinerKategorie` prüft nicht, was Name und Kommentar behaupten: Der Test bleibt grün, auch wenn der `kategorie`-Filter aus `GetProduktVorgaenger` komplett entfernt wird.                                                                                                     |
| minor   | `backend/repository/produkt_repo/mock.go`                         | Die neuen Aufzeichnungsfelder des Mocks werden von keinem Test gelesen; der Kommentar, der ihre Existenz begründet, ist falsch. Für `VerschiebeProdukt`, `VerschiebeVariante` und `SortiereVariantenAlphabetisch` gibt es auf Command-Ebene gar keinen Unit-Test.                                            |
| minor   | `backend/sqlc/queries/produkte.sql`                               | Die Begründung des `COLLATE "de-DE-x-icu"` ist sachlich falsch, und der zugehörige Test prüft die behauptete Akzentbehandlung nicht — er enthält keinen einzigen Akzent.                                                                                                                                     |
| minor   | `frontend/src/admin/products/VariantChip.tsx`                     | Products with a single variante render two permanently disabled chevrons on their only chip; every chip pays ~40 px of extra width for controls that can never fire.                                                                                                                                         |
| minor   | `backend/sqlc/queries/produkte.sql`                               | Changing a produkt's kategorie keeps its old `reihenfolge`, so the product lands at an arbitrary position in the target category's list — visible in the admin price list and on the service handhelds.                                                                                                      |

### Kommentar-Entwurf für GitHub

```markdown
Danke für den PR — die explizite Sortierung ist eine sinnvolle Ergänzung, und die Grundlagen stimmen: Migration 07 ist additiv und forward-only, der Backfill `reihenfolge = id` erhält die bisherige Ordnung, `sqlc generate` reproduziert `sqlc/dbgen` byte-genau, die neuen Endpunkte hängen korrekt am Admin-Guard und sind POST-only, und alle Werte gehen über Parameter. Vor dem Merge brauche ich aber noch ein paar Korrekturen.

**1. Verschieben ist bei Wert-Gleichstand ein stiller No-Op (wichtigster Punkt)**
`VerschiebeProdukt` tauscht die `reihenfolge`-Werte, nicht die Ränge. Haben zwei Zeilen derselben Kategorie denselben Wert, schreiben beide UPDATEs den Wert zurück, den die Zeile schon hat: HTTP 200, keine Fehlermeldung, Liste unverändert — dauerhaft, in beide Richtungen, ohne Reparaturmöglichkeit im UI.

Der Gleichstand ist über eine ganz normale Admin-Aktion erreichbar: `CreateProdukt` nummeriert pro Kategorie ab 1, `UpdateProdukt` fasst `reihenfolge` beim Kategoriewechsel aber nicht an. Nachgestellt gegen PostgreSQL mit deinen Statements: Pommes+Currywurst in `essen` (1,2), Cola+Bier in `getraenk` (1,2), dann Pommes nach `getraenk` verschieben → Pommes und Cola liegen beide auf 1 und lassen sich nie wieder aneinander vorbeischieben.

Bitte:

- beim Verschieben den Geltungsbereich vor dem Tausch dicht durchnummerieren (`row_number() OVER (ORDER BY reihenfolge, id)`) — genau das Muster, das `SortiereVariantenAlphabetisch` schon verwendet, und
- in `UpdateProdukt` bei Kategoriewechsel `reihenfolge` auf `MAX+1` der Zielkategorie setzen. Das behebt nebenbei auch, dass ein umkategorisiertes Produkt sonst an einer beliebigen Stelle der neuen Kategorie landet statt am Ende.

**2. Seeder schreibt `reihenfolge` nicht**
`SeedInsertProdukt` und `SeedInsertVariante` listen die Spalte nicht auf, es greift `DEFAULT 0`. Damit liegen alle geseedeten Produkte und Varianten auf 0 — Verschieben ist in Demo/Staging, in der kompletten e2e-Suite und über `/test/reset-and-seed` wirkungslos. Bitte die Spalte in beiden Seed-Queries mitschreiben.

**3. Rechter Chevron im Varianten-Chip liegt unter der Hit-Area des Switch**
Der `Switch` bringt `after:absolute after:-inset-x-3` mit, also 12 px unsichtbare Trefferfläche pro Seite; der Chip hat aber nur `gap-1.5` (6 px). Im Browser gemessen: die linken ~5–6 px des 20-px-Buttons treffen den Switch, nicht den Pfeil — ein Tap dort schaltet die Variante aktiv/inaktiv, statt sie zu verschieben. Der Name-Button links vom Switch ist genauso betroffen. `relative` auf `chevronClass` (und auf den Name-Button) beseitigt die tote Zone, alternativ Abstand ≥ 12 px.

**4. Touch-Targets der Chevrons**
Die beiden Chevrons sind 20x20 px. Im selben PR bekommen die Produkt-Pfeile in `ProductItem.tsx` `size="icon-sm"` (32 px) — bitte für die Varianten-Chevrons genauso, sonst ist das Feature in sich inkonsistent und auf dem Handy schwer zu treffen.

**5. Der Kategorie-Test greift nicht**
`TestVerschiebeProdukt_BleibtInSeinerKategorie` bleibt grün, auch wenn man den `kategorie`-Filter aus `GetProduktVorgaenger` komplett entfernt: Pommes und Cola haben beide `reihenfolge` 1, der Tausch schreibt 1↔1, und `GetAlleProdukte` sortiert ohnehin zuerst nach Kategorie. Bitte unterschiedliche Werte über die Kategoriegrenze hinweg wählen und die `reihenfolge`-Spalte direkt assertieren — plus einen Fall mit Gleichstand innerhalb einer Kategorie.

**Kleinigkeiten**

- Die Begründung am `COLLATE "de-DE-x-icu"` stimmt nicht: „Cafe Creme" ist reines ASCII und sortiert nirgends hinter „Cz", und die Deploy-DB läuft auf glibc `en_US.utf8`, wo Ä ohnehin bei A steht. Klausel behalten, Begründung auf „deterministische deutsche Sortierung unabhängig von der Cluster-Locale" kürzen und im Test ein echtes Akzentzeichen verwenden („Café Crème").
- Die neuen Aufzeichnungsfelder im Mock (`ProduktVerschiebungen`, `VarianteVerschiebungen`, `SortierteProdukte`) liest kein Test; der Kommentar behauptet das Gegenteil. Entweder Tests ergänzen oder Felder samt Kommentar löschen.

Wenn 1–5 drin sind, schaue ich es mir direkt wieder an.
```

## PR #110

### Begründung

Small, well-targeted change (one commit, two frontend files) that fixes a genuine correctness hazard: at 412 px the old VariantRow clamped the name to a ~149 px box, so "Schorle weiß, sauer" and "Schorle weiß, süß" collapsed to the same visible text — a real mis-booking risk. Security and supply chain are clean: no manifests, lockfiles, CI, Docker or scripts touched, no dangerous DOM or network APIs, no hidden or bidi characters, both rendered values are plain React text children. Type-check, lint, format and the full vitest suite pass. Two things must be fixed before merge. First, the new minusNurAbEins flag unmounts 88 px of stepper at quantity 0, so the first tap shrinks the name column from 324 to 236 px on a Pixel 7, names of ~30-40 characters rewrap, the row grows 61 -> 78.3 px and every row below shifts down — measured up to ~52 px for the last row, on a mobile-first ordering screen where the stepper's own unchanged contract comment promises no layout shift. Reserving the slot instead of unmounting keeps the PR's real win and removes the shift. Second, the PR falsifies committed statements without rewriting them (rule 18): VariantNamePreis.tsx still names VariantRow as a consumer although the only remaining consumer is VariantChip, the e2e overflow spec still claims the VariantNamePreis unit test covers that row, and the PR's own new comment in ProductList.tsx claims wrapping no longer happens. A unit test for the new public Stepper prop is also missing. The stale handbuch.md sentence, the website screenshots and the untouched truncation in the storno/umbuchung selection list are maintainer-side follow-ups, not contributor obligations.

### Pflicht-Fixes vor dem Merge

- frontend/src/service/components/table/ProductList.tsx / Stepper.tsx: reserve the stepper slot instead of unmounting it (fixed-width wrapper at the call site, or keep minus + count mounted with invisible/pointer-events-none at quantity 0), so the name column width stays constant and rows below do not shift
- frontend/src/components/common/VariantNamePreis.tsx: rewrite the contract comment — VariantChip is the only consumer left, and the Stepper half of "Stepper bzw. Switch" no longer applies
- e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts: drop the claim that the VariantNamePreis unit test covers the Bestellen row (no fixed price column, no truncate any more)
- frontend/src/service/components/table/ProductList.tsx: correct the new comment claiming the name "braucht in der Praxis keinen Umbruch mehr" — with the stepper mounted the column is 236 px and long names do wrap
- frontend/src/service/components/Stepper.test.tsx: add a case for minusNurAbEins (minus and quantity absent at 0, present and usable from 1); the existing three cases only exercise the default

### Bestätigte Befunde

| Schwere | Datei                                                      | Befund                                                                                                                                                                                                                                                                                                       |
| ------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| major   | `frontend/src/components/common/VariantNamePreis.tsx`      | The PR makes the contract comment of VariantNamePreis factually false: it still names VariantRow as a consumer, but VariantRow no longer uses the component. AGENTS.md rule 18 ("Nur der aktuelle Stand") requires the same change to rewrite a statement it makes false.                                    |
| major   | `frontend/src/components/common/VariantNamePreis.tsx`      | The PR drops the last service-side consumer of `VariantNamePreis` but does not update its contract comment, which still names that consumer (rule 18).                                                                                                                                                       |
| major   | `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts` | The e2e comment describing the Bestellen screen's layout contract becomes false with this PR and was not updated (rule 18).                                                                                                                                                                                  |
| major   | `frontend/src/service/components/Stepper.test.tsx`         | The PR adds a new public prop and a second visual contract with zero tests, and does not satisfy the acceptance criteria the repo's own plan defines for this change.                                                                                                                                        |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | `minusNurAbEins` makes the Stepper's width jump from 44px to 132px on the first tap, so the name column shrinks from 324px to 236px on a Pixel 7 and names of ~30–40 characters rewrap to two lines. The row grows 61px → 78.3px and every row below it is pushed down, which silently swallows the next tap |
| major   | `frontend/src/components/common/VariantNamePreis.tsx`      | Removing `VariantNamePreis` from `VariantRow` leaves two committed comments asserting a consumer relationship that no longer exists. AGENTS.md rule 18 ("Nur der aktuelle Stand") requires the same change to rewrite statements it makes false; the PR touches neither file.                                |
| minor   | `frontend/src/service/components/PositionAuswahlListe.tsx` | The mis-booking risk the PR names is only fixed in the order list; the storno/umbuchung selection list still truncates variant names to a single line, where picking the wrong entry is more damaging than in ordering.                                                                                      |
| minor   | `docs/handbuch.md`                                         | §6.3 UI-Patterns still claims card styling for products, which this PR removes from the service order list.                                                                                                                                                                                                  |
| minor   | `website/src/assets/screenshots/bestellansicht-light.png`  | Committed website screenshots of the order and Direktverkauf screens show the removed card layout and were not regenerated.                                                                                                                                                                                  |

### Kommentar-Entwurf für GitHub

```markdown
Danke für den PR — das Problem ist real und gut getroffen. Auf 390–412 px war der Variantenname vorher auf ~149 px geklemmt, „Schorle weiß, sauer" und „Schorle weiß, süß" sahen identisch aus; genau daraus entstehen Fehlbuchungen. Der Umbruch mit Preis in der zweiten Zeile beseitigt das vollständig. tsc, ESLint, Prettier und die Vitest-Suite laufen bei mir grün.

Drei Punkte, dann merge ich:

**1. `minusNurAbEins` erzeugt einen Layout-Shift**
Bei Menge 0 ist der Stepper 44 px breit, ab Menge 1 sind es 132 px. Auf einem Pixel 7 schrumpft die Namensspalte damit beim ersten Tap von 324 px auf 236 px, Namen ab ~30 Zeichen brechen dann um, die Zeile wächst von 61 px auf 78,3 px und alle Zeilen darunter rutschen nach — die letzte in meiner Messung um ~52 px, bei 44 px hohen Plus-Buttons. Ein zweiter Tap auf die vorher gemerkte Position landet dann daneben und wird kommentarlos verschluckt.

Der Kommentar im Stepper sagt oberhalb der neuen Zeile selbst: „Die Menge in der Mitte hat feste Breite, damit der Zustandswechsel keinen Layout-Shift auslöst." Bitte den Platz reservieren statt die Controls auszuhängen — entweder feste Breite am Aufrufort (`<div className="flex w-[132px] shrink-0 justify-end">`) oder Minus + Menge bei 0 mit `invisible` / `pointer-events-none` stehen lassen. Der eigentliche Gewinn des PRs (kein Kürzen mehr) bleibt dabei erhalten.

**2. Kommentare, die der PR falsch macht**

- `frontend/src/components/common/VariantNamePreis.tsx`: nennt weiterhin „Service-Bestellen (VariantRow)" als Consumer — nach dem PR ist `VariantChip` der einzige. Auch „Stepper bzw. Switch" trifft nur noch auf den Switch zu.
- `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts`: behauptet, die feste Preisspalte (Name `flex-1 truncate`, Preis `shrink-0`) decke der Unit-Test von `VariantNamePreis` ab. Beides gilt für die Bestellen-Zeile nicht mehr.
- Der neue Kommentar in `ProductList.tsx` („braucht deshalb in der Praxis keinen Umbruch mehr") stimmt so nicht: mit gemountetem Stepper bleiben 236 px, und lange Namen brechen um — was ja auch beabsichtigt ist.

**3. Test für die neue Prop**
`Stepper.test.tsx` deckt bisher nur den Default ab. Bitte einen Fall für `minusNurAbEins` ergänzen: Minus und Mengenanzeige bei 0 nicht im DOM, ab 1 vorhanden und bedienbar.

Handbuch (§6.3 „Karten für Produkte") und die Website-Screenshots ziehe ich nach dem Merge selbst nach, das musst du nicht anfassen. Die Storno-/Umbuchungs-Auswahlliste kürzt Namen übrigens noch genauso — das ist bewusst nicht Teil dieses PRs, ich mache dafür ein eigenes Issue auf.
```

## PR #111

### Begründung

Security and supply chain are clean (6 frontend files, no manifests, CI, scripts or dangerous APIs, no hidden characters, consistent authorship). But this PR is not ready, on two levels. Design first: the Bestellen screen already filters by category pills, and the new product tile grid puts an extra tap plus a screen change in front of every product — including single-variant products, where the tile itself says "1 Variante" and there is no fast path. That is a change to the core ordering path of a mobile-first app used by volunteers under stress, proposed unsolicited by an outside contributor, so it needs a maintainer design decision before any code work; product conservatism argues for deciding this before polishing the implementation. Second, even if the design is accepted, the branch is blocking: it deletes the product <h2> heading that e2e/support/servicekraft.ts anchors on and hides variant rows one level deeper, without touching a single e2e file. waehleVariante and everything built on it (bestellePosition, nimmLangeBestellungAuf) resolve to nothing across roughly ten specs, plus tischservice-viewport-ueberlauf.mobile.spec.ts:44 and e2e/website/screenshots.mjs:139; the e2e job runs on every pull request and is not covered by make check/verify, so the green unit tests prove nothing. On top of that ProductTile renders a literal ', ' as visible text (", 1 Variante") for a jsdom-only accessible-name artefact no test depends on, the whole new level ships without a single assertion, the only exit from the variant level is a ~24 px control in an area that documents 44 px elsewhere, and lg:grid-cols-4 is keyed to the viewport although the grid lives in ServiceSplitLayout's ~584 px column, making tiles at 1024-1104 px narrower than on a 360 px phone. Note the branch also contains the #110 commit, so it inherits #110's layout-shift and stale-comment items and will need a rebase once #110 lands.

### Pflicht-Fixes vor dem Merge

- Design decision first: the category pills already filter the list — clarify whether the extra product level is wanted at all, and if so how single-variant products avoid a pointless drill-down (inline stepper on the tile or auto-add)
- e2e/support/servicekraft.ts: teach waehleVariante the new navigation level (open the product tile, locate the variant row, tolerate an already-open product for consecutive calls) and rewrite the now-false comments about the flat list and the group heading
- e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts:44 and e2e/website/screenshots.mjs:139: open the owning product tile before asserting on / clicking the variant row
- frontend/src/service/components/table/ProductList.tsx: delete the visible {', '} text node and its comment in ProductTile — the tile currently renders ", 1 Variante"; no test depends on it (the helpers match by RegExp)
- Add tests for the new level: ProductList.test.tsx (tile with variant count and aggregated "N gewählt", drill-down, back button, category switch) and a Stepper case for minusNurAbEins
- frontend/src/service/components/table/ProductList.tsx: give the "Produkte" back control a real touch target (min-h-11), it is currently ~24 px tall
- frontend/src/service/components/table/ProductList.tsx: size the tile grid by its container, not the viewport (xl:grid-cols-4 or fewer columns) — inside ServiceSplitLayout's ~584 px column four columns are narrower than on a 360 px phone
- Rebase onto main once #110 is merged; the #110 items (reserve the stepper width, stale VariantNamePreis / e2e comments) apply here as well

### Bestätigte Befunde

| Schwere | Datei                                                      | Befund                                                                                                                                                                                                                                                                             |
| ------- | ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| blocker | `e2e/support/servicekraft.ts`                              | The PR removes the product `<h2>` heading that the shared e2e helper `waehleVariante` locates the variant group by, and hides variant rows behind a new navigation level, without touching a single e2e file. Every e2e spec that orders anything will fail, and e2e runs in CI.   |
| blocker | `e2e/support/servicekraft.ts`                              | The PR deletes the only product heading from ProductList, which is the anchor every Playwright ordering helper uses. `waehleVariante` (and everything built on it) can no longer find a variant row, so the CI `e2e` job fails.                                                    |
| major   | `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts` | Two more e2e entry points assert that variant names/rows are visible immediately on the Bestellen and Verkaufen screens. With the product level they are not rendered until a tile is opened.                                                                                      |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | `ProductTile` renders a literal `', '` as visible text, so every product tile shows a stray leading comma on its own line: ", 1 Variante" / ", 2 Varianten". The comma is a hack for the accessible-name computation but is not visually hidden.                                   |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | The entire new product level ships without a single test. The four touched test files only insert a navigation click so the _pre-existing_ assertions keep passing; nothing asserts any new behaviour. No `ProductList.test.tsx` exists, and the new `Stepper` branch is untested. |
| major   | `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts` | The 390px overflow regression spec asserts a variant name is visible on the Bestellen tab; after the PR that name only exists once its product tile has been opened.                                                                                                               |
| major   | `e2e/website/screenshots.mjs`                              | The website screenshot script clicks a variant row that is no longer reachable without first opening its product tile.                                                                                                                                                             |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | `ProductTile` renders a stray leading comma into the visible German label, and the accessible-name hack it is justified with does not even work for the second half of the label.                                                                                                  |
| major   | `frontend/src/components/common/VariantNamePreis.tsx`      | The PR removes the service-side consumer of `VariantNamePreis` but leaves the component's contract comment naming it — AGENTS.md rule 18 ("Nur der aktuelle Stand").                                                                                                               |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | A whole new navigation level ships with zero tests; the four test files touched are only patched to keep the pre-existing flows passing.                                                                                                                                           |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | Every product tile renders a stray leading comma: the second line literally reads ", 1 Variante". The workaround it implements is a jsdom-only artifact and is not needed by any test.                                                                                             |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | The "‹ Produkte" back button — the only way out of the new variant level — is roughly 24 px tall, far below the 44 px touch standard the service area documents for itself.                                                                                                        |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | `lg:grid-cols-4` is keyed to the viewport but the grid lives inside the ~584 px left column of ServiceSplitLayout, so at 1024–1279 px each tile is narrower than on a 360 px phone.                                                                                                |
| major   | `frontend/src/service/components/table/ProductList.tsx`    | Every product now costs one extra tap plus a full screen change, including products with a single variant, where the tile itself already says "1 Variante" — a regression on the core ordering path with no fast path.                                                             |
| minor   | `frontend/src/service/components/table/ProductList.tsx`    | Switching category does not reset the drill-down: coming back to the original category reopens the previously selected product instead of the product grid. The comment claiming otherwise is only half true.                                                                      |
| minor   | `frontend/src/components/common/VariantNamePreis.tsx`      | The PR removes the last service-side consumer of VariantNamePreis but leaves its doc comment claiming it is used by the service order list — a statement the change makes false (AGENTS.md rule 18).                                                                               |

### Kommentar-Entwurf für GitHub

```markdown
Danke, dass du dir die Bestellansicht so gründlich vorgenommen hast — und danke insbesondere für den Umbruch der Variantennamen, den ich über #110 separat übernehme.

**Zuerst eine Designfrage, bevor du weiter Arbeit reinsteckst**
Die Kategorie-Chips filtern die Liste heute schon. Die zusätzliche Produktebene kostet auf dem Kernpfad pro Produkt einen Extra-Tap plus Bildschirmwechsel — auch bei Produkten mit nur einer Variante, wo die Kachel selbst „1 Variante" anzeigt und trotzdem erst aufgeklappt werden muss. jotti wird von Ehrenamtlichen unter Zeitdruck bedient, und Bestellen ist die häufigste Aktion überhaupt; ich möchte deshalb erst entscheiden, ob die Ebene grundsätzlich rein soll, bevor wir die Details polieren. Falls ja: Produkte mit genau einer Variante brauchen einen direkten Weg (Stepper auf der Kachel oder Hinzufügen beim Tap), sonst verdoppeln wir den Aufwand für den häufigsten Fall.

Falls wir uns für die Ebene entscheiden, sind das die Punkte:

**1. Die e2e-Suite bricht komplett (blockierend)**
`e2e/support/servicekraft.ts` grenzt die Variantenzeile über `getByRole('heading', { name: produkt })` ein — das `<h2>` mit dem Produktnamen entfällt in diesem PR, der Name steht jetzt in einem `<span>` in der Kachel-Button. Außerdem existiert vor dem Öffnen einer Kachel überhaupt kein „Variante hinzufügen"-Button. `waehleVariante` findet damit nichts, und mit ihr `bestellePosition` und `nimmLangeBestellungAuf` — betroffen sind rund zehn Specs. Dazu kommen `tischservice-viewport-ueberlauf.mobile.spec.ts:44` (prüft direkt auf den Variantennamen „Fr: Schnitzel mit Pommes") und `e2e/website/screenshots.mjs:139` (klickt die Currywurst-Zeile). Im PR ist keine einzige Datei unter `e2e/` angefasst. Der `e2e`-Job läuft bei jedem PR und ist nicht Teil von `make check`/`make verify`, die grünen Unit-Tests sagen dazu also nichts. Bitte die Helfer auf die neue Ebene umstellen (Kachel öffnen, Zeile suchen; mehrere Varianten desselben Produkts nacheinander müssen ohne Zurück-Navigation funktionieren) und die Kommentare über der „flachen Liste" mitziehen.

**2. Sichtbares Komma auf jeder Kachel**
`ProductTile` rendert `{', '}` als echten Textknoten; die Kachel zeigt deshalb in der zweiten Zeile wörtlich „, 1 Variante". Der Grund im Kommentar greift nur in jsdom — im Browser sind beide Spans Flex-Items und werden ohnehin getrennt. Deine eigenen Test-Helfer matchen per `new RegExp(name)`, hängen also nicht daran. Bitte die Zeile samt Kommentar streichen.

**3. Keine Tests für die neue Ebene**
Die vier Testdateien bekommen nur den `produktOeffnen`-Klick, damit die bestehenden Fälle weiterlaufen — neue Assertions gibt es keine. Ungetestet sind: Kachel mit Variantenzahl und aggregiertem „N gewählt", Zurück-Navigation, das implizite Zurückfallen beim Kategoriewechsel (steht bisher nur als Kommentar im Code) und `minusNurAbEins`. Für Letzteres gibt es mit `Stepper.test.tsx` schon eine Datei, die genau dieses Verhalten absichert.

**4. Zurück-Button zu klein**
„‹ Produkte" ist mit `py-1` und 13-px-Text ~24 px hoch und ist der einzige Weg zurück. Der Stepper im selben Bereich dokumentiert 44 px als Standard — bitte `min-h-11`.

**5. Kachel-Grid richtet sich nach dem Viewport statt nach dem Container**
`lg:grid-cols-4` greift ab 1024 px, dort steckt das Grid aber in der ~584 px breiten linken Spalte von `ServiceSplitLayout`. Ergebnis: zwischen 1024 und ~1100 px sind die Kacheln schmaler als auf einem 360-px-Handy. `xl:grid-cols-4` oder generell 2–3 Spalten passt besser; die anderen Grids in der Split-Spalte machen es genauso.

**Noch zwei Kleinigkeiten:** der Kontrakt-Kommentar in `VariantNamePreis.tsx` nennt weiterhin `VariantRow` als Consumer, und ein Kategoriewechsel setzt die aufgeklappte Produktansicht nicht zurück — beim Zurückwechseln steht man wieder in der Variantenliste statt im Kachelraster.

**Hinweis zum Branch:** #110 steckt hier als Commit mit drin. Ich merge #110 separat; bitte danach rebasen. Die Punkte aus #110 (Stepper-Breite reservieren, veraltete Kommentare) gelten hier genauso.
```

## Verworfene Befunde

- #109 `backend/repository/produkt_repo/repo.go`: Der Tausch liest den Nachbarn ohne Sperre und schreibt beide Zeilen; zwei überlappende gleichzeitige Verschiebungen verlieren ein Update und hinterlassen doppelte `reihenfolge`-Werte, die dann dauerha
- #109 `docs/handbuch.md`: docs/handbuch.md was not updated, although it is the canonical architecture/invariants doc and the repo's own plan names it as an acceptance criterion for exactly this PR. Only docs/language.md got an
- #109 `backend/api/admin.go`: The three new endpoint paths are German verb-first (`/verschiebe-produkt`, `/verschiebe-variante`, `/sortiere-varianten`). Every existing route in the repo is either English-verb-first or German noun-
- #109 `backend/repository/produkt_repo/repo.go`: The new persistence-layer methods and helpers use German verbs, which contradicts the documented rule that persistence takes English verbs — and the PR itself declares the feature "reine Persistenz".
- #109 `frontend/src/admin/products/Produkt.ts`: The PR introduces a bare `Richtung` type and JSON key `richtung` for hoch/runter, colliding with the established `richtung` of the Geldtransit (einlage/entnahme). The repo's pattern is to context-pref
- #109 `frontend/src/admin/products/Produkt.ts`: `RichtungSchema` repeats the literals instead of deriving them from the `Richtung` const, and every call site passes raw `'hoch'`/`'runter'` strings, so the exported const is never actually used as a
- #109 `backend/sqlc/queries/produkte.sql`: The PR deletes all seven pre-existing blank separator lines between the queries in produkte.sql — unrelated to the feature — leaving it the only query file in the directory without blank-line separati
- #109 `frontend/src/admin/products/Produkt.ts`: The exported `Richtung` const object is dead — no call site uses it as a value — and `RichtungSchema` re-lists the literals instead of deriving from it, breaking the established pattern.
- #110 `frontend/src/service/components/Stepper.tsx`: The PR adds a new public Stepper prop with a rendering branch and ships zero tests for it. Stepper.test.tsx is untouched, and there is no ProductList test file at all, so the new hide-at-zero behaviou
- #110 `frontend/src/service/components/Stepper.tsx`: The new prop name `minusNurAbEins` is German while every sibling prop in the same interface is English, and it names a policy rather than the rendering it controls.
- #110 `frontend/src/service/components/table/ProductList.tsx`: The product group heading is now CSS-uppercased, which alters admin-configured product names on screen and can collapse two distinct names into the same rendering — the exact failure mode the PR's own
- #110 `frontend/src/service/components/table/ProductList.tsx`: The inlined price loses `font-semibold` and moves to `text-muted-foreground`, so the same variant price is now rendered two different ways in the app: full-contrast semibold in the admin chip, muted r
- #110 `frontend/src/service/components/Stepper.tsx`: The PR adds a mode that causes a layout shift but leaves the contract comment claiming the opposite standing right above it (AGENTS.md rule 18: a change that makes a statement false rewrites it in the
- #110 `frontend/src/service/components/table/ProductList.tsx`: The single commit has no body although it changes two files; AGENTS.md requires a bullet body for multi-file changes.
- #110 `frontend/src/service/components/Stepper.tsx`: The new `minusNurAbEins` branch is untested, although Stepper.test.tsx covers every other branch of the same component, and the PR ships no e2e test although the plan's Phase 4 acceptance criteria dem
- #111 `frontend/src/service/TablePage.test.tsx`: The `produktOeffnen` helper is copy-pasted verbatim, comment included, into four test files. AGENTS.md / handbook require a single source of truth instead of duplicated content.
- #111 `frontend/src/service/components/Stepper.tsx`: The rationale comment for the new `minusNurAbEins` prop describes a truncation risk that the PR's own `VariantRow` no longer has, so it documents a state that is not current (AGENTS.md rule 18).
- #111 `frontend/src/service/components/Stepper.tsx`: The pre-existing Stepper contract comment keeps asserting a no-layout-shift guarantee that the new `minusNurAbEins` mode removes; the new paragraph is appended beside it instead of qualifying it.
- #111 `frontend/src/service/TablePage.test.tsx`: The identical 9-line `produktOeffnen` helper (comment included) is copy-pasted into four test files although the repo has a shared test directory.
- #111 `frontend/src/service/components/table/ProductList.tsx`: Both commits change multiple files but carry an empty body, contrary to the AGENTS.md git workflow.
- #111 `frontend/src/service/components/table/ProductList.tsx`: The PR title carries review metadata that a squash merge would bake into the permanent main commit message.
- #111 `frontend/src/service/components/Stepper.tsx`: `minusNurAbEins` reintroduces exactly the layout shift that the Stepper's own (unchanged) doc comment says the fixed-width quantity exists to prevent: the first "+" tap shrinks the name column by 88 p
- #111 `frontend/src/service/components/table/ProductList.tsx`: A whole new navigation level, a new component (ProductTile) and a new shared Stepper prop ship with zero new tests; the existing tests were only patched so they keep passing.
- #111 `frontend/src/service/components/table/ProductList.tsx`: The accessible name of a selected tile still runs two numbers together — "…1 Variante2 gewählt" — the exact defect the comma at line 183 was added to avoid, now reproduced one element later.
- #111 `frontend/src/service/components/table/ProductList.tsx`: On the variant level the only back control scrolls out of view, while the category chip row it replaces is sticky.
