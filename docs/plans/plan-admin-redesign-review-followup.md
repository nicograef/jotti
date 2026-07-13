# Plan: Admin-UI-Redesign — Architektur-Review-Nachbereitung

> Source PRD: n/a (aus Architektur-Review, Task-Beschreibung)

## Ziel

Drei Nachbesserungen aus dem Architektur-Review des Admin-UI-Redesigns umsetzen:

1. Die schwach begründete **Produkt-Verkaufs-Lesekante** (Stammdaten liest
   `kassenjournal`) vollständig entfernen — sie schützt keine Korrektheit
   (Soft-Delete + Fat Events machen das Löschen eines verkauften Produkts
   folgenlos) und verschmutzt die Context Map mit einer zweiten Kasse→Stammdaten-
   Rückkante.
2. Eine **Defense-in-Depth-Lücke** schließen: Bestellungen und Direktverkäufe
   akzeptieren serverseitig deaktivierte (`inactive`) Produkte/Varianten.
3. Die **`docs/`-Dokumentation** gegen den Status-quo-Code korrigieren
   (verifizierte Findings).

Nach Umsetzung hat die Context Map wieder genau **eine** bewusste, begründete
Kasse→Stammdaten-Rückkante (Tisch-Saldo), und sie ist dokumentiert.

## Architektonische Entscheidungen (durabel, phasenübergreifend)

- **Tisch-Saldo-Kante bleibt.** Stammdaten liest weiter `tisch_sessions` +
  `kassensitzungen` für Saldo-Anzeige und Lösch-/Deaktivier-Schutz. Begründung:
  nur aktive Tische sind kassier-/stornier-/umbuchbar
  (`backend/api/kasse/tischgeschaeft/application/command.go — loadTischState()`
  erzwingt `tisch.ActiveStatus`); ein deaktivierter Tisch mit offenem Saldo würde
  Geld stranden. Die Query liest nur Projektionsspalten (kein Event-Contract),
  bleibt bewusst im `tisch_repo` (Präzedenz: `GetActiveTables` liest
  `tisch_sessions` bereits). **Keine** Verlagerung nach `kassensitzungen_repo`
  (Over-Engineering, gegen Produkt-Konservatismus).
- **Produkt-Retirement-Pfad bleibt Varianten-Deaktivierung.** `Produkt` hat kein
  `Deactivate()` (nur `Delete()`); ein Produkt wird aus dem Bestell-Menü genommen,
  indem alle Varianten deaktiviert werden (`GetAktiveProdukte` blendet Produkte
  ohne aktive Variante per `INNER JOIN` aus). `DeleteProdukt` wird nach Phase 1
  wieder reiner Soft-Delete (`status = 'deleted'`).
- **Fat Events sind Single Source of Truth für historische Produktdaten.**
  Reporting (`reporting.sql`), Beleg-Druck (`beleg/.../kassenbeleg_command.go`)
  und DSFinV-K-Export (`fiskal/dsfinvk/mapper.go`) lesen Produktdaten aus den
  eingefrorenen Positionen der Events, nie aus der `produkte`-Tabelle. Deshalb ist
  das Löschen eines verkauften Produkts folgenlos.
- **API bleibt POST-only, additiv.** Keine Migration nötig (nur Query-Entfernung
  in Phase 1). Frontend + Backend werden gemeinsam ausgeliefert (Freeze-Disziplin:
  API-Formate dürfen sich ändern).

## Inventar (relevante bestehende Symbole)

**Phase 1 (Produkt-Kante):**
- `backend/api/stammdaten/produkt/application/query.go — Query.GetAllProducts()`,
  `ProduktMitVerkauf`, `produktQueryRepo.GetProduktIDsMitVerkaeufen()`
- `backend/api/stammdaten/produkt/application/command.go — Command.DeleteProdukt()`,
  `produktRepo.ProduktHatVerkaeufe()`
- `backend/api/stammdaten/produkt/application/errors.go — ErrProduktHatVerkaeufe`
- `backend/api/stammdaten/produkt/http/query_handler.go — getAllProductsResponse`,
  `toProdukteMitVerkauf()`, `toProdukte()`, `produkt.HatVerkaeufe` (DTO-Feld),
  `GetAllProductsHandler()`
- `backend/api/stammdaten/produkt/http/command_handler.go — DeleteProduktHandler()`
  (Fehler-Map-Eintrag `ErrProduktHatVerkaeufe: "produkt_hat_verkaeufe"`)
- `backend/repository/produkt_repo/repo.go — GetProduktIDsMitVerkaeufen()`,
  `ProduktHatVerkaeufe()`; `backend/repository/produkt_repo/mock.go` (dieselben)
- `backend/sqlc/queries/produkte.sql — GetProduktIDsMitVerkaeufen`, `ProduktHatVerkaeufe`
- `backend/repository/produkt_repo/sales_test.go` (ganze Datei)
- `frontend/src/admin/products/Produkt.ts` (`hatVerkaeufe` im Zod-Schema)
- `frontend/src/admin/products/ProductItem.tsx` (`hatVerkaeufe`-Branch im „···"-Menü)
- `frontend/src/admin/products/NewProductDialog.tsx` (`hatVerkaeufe: false`)
- `frontend/src/admin/products/Products.tsx` (`HinweisKarte`-Text)
- `frontend/src/lib/errorMessages.ts` (Key `produkt_hat_verkaeufe`)
- `frontend/src/admin/products/Products.test.tsx`

**Phase 2 (Status-Guard):**
- `backend/api/kasse/tischgeschaeft/application/command.go — BestellungAufnehmen()`
  (Batch-Fetch + Positions-Anreicherung), `errors.go`
- `backend/api/kasse/direktverkauf/application/command.go — Direktverkauf-Command`
  (Batch-Fetch), `backend/api/kasse/direktverkauf/application/errors.go`
- `backend/repository/produkt_repo/batch.go — GetVariantsByIDs()`,
  `GetProductsByIDs()` (filtern `status != 'deleted'`, lassen `inactive` durch)
- `backend/domain/produkt — Variante.Status`, `Produkt.Status` (`ActiveStatus`)
- `frontend/src/lib/errorMessages.ts` (neuer Fehler-Key)

**Phase 3 (Doku):**
- `docs/handbuch.md` §2.2 (Context Map), §3.9 (Kassenbestand), §4.6 (Bondruck),
  §7.1/§7.2 (Read Models + Endpunkte)
- `docs/language.md` (`bon_art`-Enum; Glossar; Kassenbestand-JSON-Keys)
- `docs/prds/prd-admin-redesign.md` **(löschen)** — Redesign abgeschlossen (PR #84),
  Aussage „Produkte mit Verkäufen sind nur deaktivierbar" ist zudem obsolet.
- `docs/prds/design_handoff_admin_redesign/` **(löschen)** — Handoff-Ordner (3
  `.dc.html` + `README.md`), reines Übergabe-Artefakt.
- `docs/prds/prd-windows-nativ-ohne-docker.md` — **bleibt** (nicht redesign-bezogen).
- `docs/README.md` (Index, falls Abschnittsverweise berührt — aktuell nicht indexiert)
- Belege für die Doku: `backend/domain/kasse/kassensitzung.go — Kassenbestand`,
  `Geldtransit`; `backend/domain/reporting/reporting.go — Metadaten`,
  `AbgeschlosseneSitzung`; `backend/domain/betreiber/betreiber.go — ElsterGemeldetAm`;
  `backend/api/admin.go` (`/get-abgeschlossene-kassensitzungen`, `/testbon-drucken`);
  `database/migrations/04_testbon_bonart.up.sql`

## Aufgelöste Entscheidungen

- **Produkt-Verkaufs-Kante:** entfernen (nicht nach Kasse verschieben). Bestätigt
  vom Nutzer.
- **Tisch-Saldo-Kante:** unverändert im `tisch_repo` belassen und dokumentieren.
- **Reihenfolge:** Phase 1 → Phase 3. Phase 3 (§2.2, §7) darf die Produkt-Kante
  nicht dokumentieren, weil Phase 1 sie entfernt. Phase 2 ist unabhängig.
- **Phase 2 Fehlerwahl (entschieden):** eigener Fehler `ErrVarianteNichtAktiv` je
  Kontext-`errors.go`, um „nicht gefunden" (deleted) von „deaktiviert" (inactive)
  zu trennen — sauberere Fehlermeldung als Reuse von `ErrProduktNotFound`.
- **Phase 3 Redesign-Doku (entschieden):** die Redesign-PRD und den Handoff-Ordner
  **löschen** statt zu korrigieren (Redesign abgeschlossen, PR #84; Git-Historie
  bewahrt sie). Damit entfällt Finding #6 (Wortlaut-Korrektur) — die Aussage
  verschwindet mit der Datei.

## Open Questions / Risiken

- **Risiko Phase 1:** e2e-Test `e2e/tests/admin-produkte-verwalten.spec.ts` könnte
  den Lösch-Guard prüfen (Assertion „Löschen disabled bei Verkäufen"). Muss beim
  Entfernen angepasst werden — beim Bearbeiten gezielt suchen.

---

## Phase 1: Produkt-Verkaufs-Kante entfernen

### Kontext

- `backend/api/stammdaten/produkt/application/query.go — Query.GetAllProducts()` —
  gibt aktuell `[]ProduktMitVerkauf` zurück; die Journal-Projektion
  `GetProduktIDsMitVerkaeufen` reichert `HatVerkaeufe` an. Beides entfällt.
- `backend/api/stammdaten/produkt/application/command.go — Command.DeleteProdukt()` —
  ruft den `ProduktHatVerkaeufe`-Guard vor dem Soft-Delete. Guard entfällt.
- `backend/sqlc/queries/produkte.sql` — enthält die beiden `kassenjournal`-Queries,
  die die Kante physisch realisieren. Entfernen → `make sqlc`.
- `frontend/src/admin/products/ProductItem.tsx` — der `hatVerkaeufe`-Branch
  disabled „Löschen…". Ohne Guard ist Löschen immer aktiv.

### Was zu bauen ist

Die Kasse→Stammdaten-Produkt-Kante End-to-End entfernen: SQL-Queries →
Repository (+ Mock) → Application-Query/-Command (+ Interfaces + Fehler) →
HTTP-DTO/-Fehler-Mapping → Frontend-Schema/-UI/-Fehlertext → Tests.

`GetAllProducts` liefert wieder das reine Domain-Produkt (`[]produkt.Produkt`,
via `toProdukte()`); die Produktliste zeigt keinen Verkaufsstatus mehr.
`DeleteProdukt` ist wieder reiner Soft-Delete für **alle** Produkte. Der
Retirement-Pfad (Varianten deaktivieren) und „Alle Varianten deaktivieren" im
„···"-Menü bleiben unverändert. Der `HinweisKarte`-Text in `Products.tsx` wird
auf die Varianten-Deaktivierung reduziert (der „nur für Produkte ohne Verkäufe"-
Zusatz entfällt).

### Akzeptanzkriterien

- [ ] `GetProduktIDsMitVerkaeufen` und `ProduktHatVerkaeufe` sind aus
      `produkte.sql`, `produkt_repo` (repo + mock) und den beiden Application-
      Interfaces entfernt; `make sqlc` neu generiert; kein `sqlc/dbgen/`-Handedit.
- [ ] `ProduktMitVerkauf`, `ErrProduktHatVerkaeufe`, das DTO-Feld `hatVerkaeufe`
      und der Fehler-Map-Eintrag `produkt_hat_verkaeufe` existieren nicht mehr
      (Backend + Frontend-Schema + `errorMessages.ts`).
- [ ] `DeleteProdukt` löscht ein Produkt **mit** Verkäufen erfolgreich
      (Soft-Delete `status = 'deleted'`); Reporting-, Beleg- und DSFinV-K-Pfade
      bleiben unberührt (lesen aus Fat Events).
- [ ] `ProductItem` zeigt „Löschen…" immer aktiv; „Alle Varianten deaktivieren"
      unverändert vorhanden.
- [ ] `sales_test.go` gelöscht; `command_test.go`, `query_handler_test.go`,
      `Products.test.tsx` und ggf. `admin-produkte-verwalten.spec.ts` angepasst.
- [ ] `make check` und `make lint` grün.

---

## Phase 2: Bestellung/Direktverkauf auf deaktivierte Produkte/Varianten ablehnen

### Kontext

- `backend/repository/produkt_repo/batch.go — GetVariantsByIDs()`,
  `GetProductsByIDs()` — filtern `status != 'deleted'`, lassen also `inactive`
  durch. Beide Verkaufspfade reichern Positionen darüber an.
- `backend/api/kasse/tischgeschaeft/application/command.go — BestellungAufnehmen()`
  und `backend/api/kasse/direktverkauf/application/command.go` — prüfen nach dem
  Batch-Fetch nur Existenz, nicht `Status == ActiveStatus`.

### Was zu bauen ist

In beiden Command-Sites nach dem Batch-Fetch prüfen, dass jede referenzierte
Variante **und** ihr Produkt `produkt.ActiveStatus` haben; andernfalls die
Operation mit einem eigenen Fehler `ErrVarianteNichtAktiv` (je Kontext-`errors.go`)
ablehnen — getrennt von `ErrProduktNotFound` (das für gelöschte/nicht existente
IDs bleibt). Gemeinsame Prüf-Logik ggf. als kleine Hilfsfunktion. HTTP-Fehler-
Mapping in beiden HTTP-Command-Handlern ergänzen; passenden Key in
`frontend/src/lib/errorMessages.ts` hinzufügen (deutsche Meldung).

Reine serverseitige Härtung (Defense-in-Depth); das Frontend-Menü
(`GetAktiveProdukte`) zeigt deaktivierte Varianten ohnehin nicht — normales
UI-Verhalten ändert sich nicht.

### Akzeptanzkriterien

- [ ] Ein direkter POST mit einer `inactive` Varianten- oder Produkt-ID an
      `BestellungAufnehmen` **und** an den Direktverkauf-Endpunkt wird mit
      `ErrVarianteNichtAktiv` (bzw. dessen HTTP-Mapping) abgelehnt.
- [ ] Eine `deleted` ID liefert weiterhin `ErrProduktNotFound` (unverändert).
- [ ] Aktive Varianten/Produkte funktionieren wie bisher; Menü-Verhalten und
      bestehende Bestell-/Direktverkauf-Tests unverändert grün.
- [ ] Neue Tests für beide Pfade decken den `inactive`-Abweisfall ab.
- [ ] `frontend/src/lib/errorMessages.ts` enthält den neuen Fehler-Key mit
      deutscher, servicetauglicher Meldung.
- [ ] `make check` und `make lint` grün.

---

## Phase 3: Dokumentation gegen Status-quo-Code korrigieren

### Kontext

Verifizierte Findings aus dem Doku-Review. Reihenfolge nach Phase 1 (die
Context-Map- und Read-Model-Einträge dürfen die entfernte Produkt-Kante nicht
enthalten).

### Was zu bauen ist

Die folgenden Doku-Stellen korrigieren; nach jeder Änderung `grep` auf
Dead-Links und `docs/README.md`-Index prüfen.

1. **`docs/language.md` — `bon_art`-Enum (WRONG):** um `'testbon'` ergänzen
   (`('arbeitsbon' | 'kassenbeleg' | 'testbon')`), passend zu
   `database/migrations/04_testbon_bonart.up.sql`.
2. **`docs/handbuch.md` §2.2 — Context Map (STALE):** die bewusste Kasse→Stammdaten-
   **Tisch-Saldo**-Lesekante als eigene Beziehungszeile aufnehmen (Stammdaten liest
   `tisch_sessions`/`kassensitzungen` für Saldo-Anzeige und Lösch-/Deaktivier-
   Schutz; read-only). Die Aussage „keine Cross-Context-Kommunikation nötig" so
   präzisieren, dass sie nur noch für Reporting gilt. **Die Produkt-Kante nicht
   eintragen** (in Phase 1 entfernt).
3. **`docs/handbuch.md` §7 — Read Models (MISSING):** Read-Model „Tischliste mit
   offenem Saldo" (`GetAllTische → TischMitSaldo`) und den Endpunkt
   `POST /admin/get-abgeschlossene-kassensitzungen`
   (`GetAbgeschlosseneKassensitzungenHandler`) ergänzen. **Kein** Produkt-
   `hatVerkaeufe`-Read-Model dokumentieren.
4. **`docs/handbuch.md` — `get-abrechnung`-Inhalt (MISSING):** `metadaten`
   (`EroeffnetAm`, `AbgeschlossenAm`, `AbgeschlossenVon`, `KassensturzDifferenzCents`)
   und die Abgeschlossene-Sitzungsliste (`AbgeschlosseneSitzung`) ergänzen, passend
   zu `backend/domain/reporting/reporting.go`.
5. **`docs/handbuch.md` §3.9 + `docs/language.md` — Kassenbestand (MISSING):** die
   Vier-Komponenten-Aufschlüsselung dokumentieren
   (`Anfangsbestand + Bareinnahmen + Einlagen − Entnahmen = Soll`, aus
   `backend/domain/kasse/kassensitzung.go — Kassenbestand`) und das Geldtransit-
   Listen-Read-Model erwähnen; in `language.md` die JSON-Keys
   `anfangsbestandCents`, `bareinnahmenCents`, `einlagenCents`, `entnahmenCents`
   ergänzen.
6. **Redesign-Doku löschen (Aufräumen):** `docs/prds/prd-admin-redesign.md` und den
   gesamten Ordner `docs/prds/design_handoff_admin_redesign/` (3 `.dc.html` +
   `README.md`) entfernen — Redesign ist abgeschlossen (PR #84), reine Übergabe-
   Artefakte, Git-Historie bewahrt sie. Damit ist auch die obsolete Aussage
   „Produkte mit Verkäufen sind nur deaktivierbar" weg (keine separate Korrektur
   nötig). `docs/prds/prd-windows-nativ-ohne-docker.md` bleibt. Vor dem Löschen
   `grep` auf Referenzen (aktuell keine) und `docs/README.md`-Index prüfen.
7. **`docs/language.md` — Glossar (MISSING):** Einträge für `Testbon`, `Zählhilfe`
   (`frontend/src/admin/kasse/zaehlhilfe.ts`) und `elster_gemeldet_am` /
   `ElsterGemeldetAm` (`backend/domain/betreiber/betreiber.go`) ergänzen.

`docs/compliance.md` §7.4 (ELSTER „gemeldet am") ist bereits korrekt — **nicht
ändern**.

### Akzeptanzkriterien

- [ ] Keine Doku-Aussage widerspricht nach Phase 1+2 dem Code (Context Map,
      Read Models, `bon_art`-Enum, Kassenbestand, Reporting-Metadaten,
      Produkt-Status-Semantik).
- [ ] `docs/handbuch.md` §2.2 nennt genau eine Kasse→Stammdaten-Rückkante
      (Tisch-Saldo); die Produkt-Kante ist nirgends dokumentiert.
- [ ] `docs/language.md` Glossar/Enums/JSON-Keys decken Testbon, Zählhilfe,
      `elster_gemeldet_am` und die Kassenbestand-Aufschlüsselung ab.
- [ ] `docs/prds/prd-admin-redesign.md` und `docs/prds/design_handoff_admin_redesign/`
      sind gelöscht; `docs/prds/prd-windows-nativ-ohne-docker.md` bleibt.
- [ ] `grep -r` auf umbenannte/entfernte Begriffe zeigt keine Dead-Links;
      `docs/README.md`-Index konsistent.
- [ ] `docs/compliance.md` unverändert.

---

## Verifikation (gesamt)

- `make check` (Unit + Lint ohne DB-Integration) nach Phase 1 und Phase 2.
- `make lint` nach Code-Änderungen; `make sqlc` nach Query-Entfernung (Phase 1).
- Betroffene e2e-Specs sichten/anpassen: `e2e/tests/admin-produkte-verwalten.spec.ts`.
- Nach Phase 3: Dead-Link-`grep`, `docs/README.md`-Index.

## Vorgeschlagene Commit-Messages (Conventional Commits, englisch, keine KI-Attribution)

**Phase 1:**
```
refactor(produkt): drop kassenjournal read for product deletion

Remove the Stammdaten→Kasse sales projection (GetProduktIDsMitVerkaeufen /
ProduktHatVerkaeufe) and the delete guard. Soft-delete plus fat events make
deleting a sold product harmless (reporting, receipts and DSFinV-K read frozen
event data, never the produkte table). Product retirement stays via variant
deactivation. Removes the weak Kasse→Stammdaten context edge.
```

**Phase 2:**
```
fix(kasse): reject orders and direct sales for inactive products

BestellungAufnehmen and Direktverkauf enriched positions from batch fetches that
only filter deleted, so a deactivated variant could still be booked via a direct
POST. Validate active status in both commands and return a dedicated error.
```

**Phase 3:**
```
docs: align handbuch/language with code and drop stale redesign docs

Document the deliberate Kasse→Stammdaten tisch-saldo read edge, the admin
read models (tisch saldo list, abgeschlossene Kassensitzungen), reporting
metadata and the Kassenbestand breakdown. Add testbon to the bon_art enum and
glossary terms (Testbon, Zählhilfe, elster_gemeldet_am). Remove the completed
admin-redesign PRD and design handoff folder.
```
