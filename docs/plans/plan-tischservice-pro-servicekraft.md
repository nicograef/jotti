# Plan: Tischservice pro Servicekraft (Zuordnung, Filter, Schichtende)

> Source PRD: [docs/prds/prd-tischservice-pro-servicekraft.md](../prds/prd-tischservice-pro-servicekraft.md)

## Goal

Jede offene Tisch-Position trägt den Besteller (die Servicekraft, die die
Bestellung aufgenommen hat). Darauf bauen vier Sichten auf:

- Eine persönliche "Alles erledigt!"-Aussage je Servicekraft über alle Tische
  der offenen Kassensitzung, an denen sie bestellt hat.
- "Eigene zuerst" beim Kassieren und Ausgabe-Bestätigen (fremde Positionen
  eingeklappt).
- Besteller-Name an jeder Bestellung beim Stornieren und Umbuchen (flach,
  nichts eingeklappt).
- Schichtende-Prüfung pro Servicekraft im Admin-Live-Dashboard.

Die Umsatzzuordnung (kassiert-basiert) und der Tisch-Gesamtzustand bleiben
unverändert. Der Direktverkauf ist nicht betroffen.

## Architectural decisions

Durchgängig gültige Entscheidungen:

- **Datenmodell Position:** Die Projektions-Struktur `Position` erhält
  `BestellerUserID int` und `BestellerName string` als reine
  Projektions-/Anzeige-Felder. Die Event-Form `PositionEventData` bleibt
  unverändert; der Besteller steht bereits im Event-Umschlag (`UserID`,
  `UserName`) und wird beim Anwenden des `bestellung-aufgenommen`-Events von
  dort übernommen.
- **Eingefrorener Name:** Der Besteller-Name ist der zum Bestellzeitpunkt
  eingefrorene Username aus dem Event-Umschlag (siehe abgeschlossenen Plan
  [plan-bediener-username.md](plan-bediener-username.md)). Spätere Umbenennungen
  ändern alte Positionen nicht. Kein neues Namensfeld im Journal.
- **Keine Migration:** Es gibt keine produktiven Instanzen. Sobald `Position`
  die Besteller-Felder trägt, serialisiert die JSONB-Projektion sie automatisch
  (Marshal nach Go-Feldnamen). Die Projektion wird per Replay neu aufgebaut.
- **Deep Module in der `kasse`-Domäne:** DB-freie reine Funktionen berechnen die
  offene Arbeit aus Tisch-Sessions. Zwei Ebenen: pro Tisch+Servicekraft (eigene
  ausstehende/unbezahlte Positionen, "für diese Person an diesem Tisch
  erledigt") und ein Rollup über mehrere Sessions (Liste offener Tische,
  Anzahlen, Gesamt-Kennzeichen). Eingaben sind Events bzw. Tisch-Sessions,
  Ausgaben sind die berechneten Sichten.
- **Erledigt-Definition:** "Erledigt" heißt keine eigenen ausstehenden UND keine
  eigenen unbezahlten Positionen. Die tischweite ausstehende Auszahlung
  (negativer Saldo) fließt nicht ein.
- **Tischkarte-Zählung:** "offen" auf der Tischkarte ist eine Zahl, die
  nicht-erledigte eigene Positionen als Vereinigung zählt (ausstehend ODER
  unbezahlt, je Position einmal), deckungsgleich mit der Erledigt-Aussage.
- **Keine neuen Routen:** Bestehende Endpunkte werden erweitert. Die
  Tisch-State-Abfragen erhalten die anfragende `userID` (aus dem JWT, wie schon
  `GetAktiveTischeMitFavoriten`). Die persönliche Übersicht erweitert
  `/get-eigene-uebersicht`, die Schichtende-Sicht erweitert das bestehende
  Live-Reporting.
- **Frontend-Gruppierung:** "eigene vs. fremde" Positionen werden client-seitig
  über `bestellerUserId` gebildet, analog zum bestehenden `isFromUser`-Muster
  der Historie. Der Tisch-Gesamtzustand wird weiterhin aus der vollständigen
  Liste dargestellt. Filter sind rein visuell, sperren nichts.

## Inventory

Domäne / Projektion:

- [backend/domain/kasse/bestellung.go:12-21](../../backend/domain/kasse/bestellung.go) —
  `Position` (Ziel der Besteller-Felder).
- [backend/domain/kasse/bestellung.go:30-57](../../backend/domain/kasse/bestellung.go) —
  `PositionEventData` (bleibt) + `toPositionenEventData`/`fromPositionenEventData`
  nutzen heute die **direkte Struct-Konvertierung** `PositionEventData(p)` /
  `Position(p)`. Diese bricht, sobald `Position` ein Feld mehr hat, und muss auf
  explizites Feld-Mapping umgestellt werden.
- [backend/domain/kasse/bestellung.go:77-85](../../backend/domain/kasse/bestellung.go) —
  Historien-`Bestellung` trägt `UserID`, aber **kein** `UserName` (für den
  Storno/Umbuchungs-Label in Phase 4 zu ergänzen).
- [backend/domain/kasse/tisch_session.go:30-37](../../backend/domain/kasse/tisch_session.go) —
  `ApplyEvent` für `bestellung-aufgenommen`: hier wird die Position mit
  `evt.UserID`/`evt.UserName` getagt.
- [backend/domain/kasse/tisch_session.go:127-160](../../backend/domain/kasse/tisch_session.go) —
  `accumulatePositionen`/`reduceByPosition` matchen über `PositionID` und
  erhalten den Eintrag (und damit das Besteller-Tag) bei Zahlung/Storno/Ausgabe.
- [backend/domain/kasse/tisch_session_test.go](../../backend/domain/kasse/tisch_session_test.go) —
  Prior Art für Besteller-Tagging-Tests.

Persistenz / Projektion:

- [backend/repository/kassenjournal_repo/repo.go:292-318](../../backend/repository/kassenjournal_repo/repo.go) —
  `upsertTischSessionState` marshalt Positionen nach JSONB (Go-Feldnamen, daher
  Besteller automatisch enthalten).
- [backend/repository/kassenjournal_repo/repo.go:349-378](../../backend/repository/kassenjournal_repo/repo.go) —
  `toTischSession` (Unmarshal) und
  [repo.go:516-609](../../backend/repository/kassenjournal_repo/repo.go) —
  `RebuildAllProjections` + `GetTischSessionsByKassensitzungNr` (bereits
  vorhanden, speist den Rollup).

Tisch-Sichten (Service):

- [backend/api/table/application/query.go:13-22](../../backend/api/table/application/query.go) —
  `TischStateView` (um Besteller je Position + "für mich erledigt" je Tisch zu
  ergänzen).
- [backend/api/table/application/query.go:67-185](../../backend/api/table/application/query.go) —
  `GetTischState`/`GetMeineTischeState` (erhalten die anfragende `userID`).
- [backend/api/table/http/query_handler.go:19-21](../../backend/api/table/http/query_handler.go) —
  Handler-Interface (userID aus dem JWT durchreichen, Muster:
  [query.go:110-131](../../backend/api/table/application/query.go) bei
  `GetAktiveTischeMitFavoriten`).
- [backend/api/table/application/query_test.go](../../backend/api/table/application/query_test.go) —
  Prior Art für Tisch-State-Tests.

Persönliche Übersicht (Service):

- [backend/domain/reporting/reporting.go:96-101](../../backend/domain/reporting/reporting.go) —
  `EigeneUebersicht` (um offene eigene Arbeit zu erweitern).
- [backend/api/reporting/application/query.go:115-132](../../backend/api/reporting/application/query.go) —
  `GetEigeneUebersicht` (lädt zusätzlich Tisch-Sessions und ruft den Rollup).
- [frontend/src/service/table/Tisch.ts:38-44](../../frontend/src/service/table/Tisch.ts) —
  `EigeneUebersichtSchema`, [Tisch.ts:18-26](../../frontend/src/service/table/Tisch.ts)
  `TischSessionSchema` (Besteller-Felder + "für mich erledigt").
- [frontend/src/service/table/Bestellung.ts:6-16](../../frontend/src/service/table/Bestellung.ts) —
  `PositionSchema` (Besteller-Felder).
- [frontend/src/service/components/EigeneUebersicht.tsx](../../frontend/src/service/components/EigeneUebersicht.tsx) —
  Karten der persönlichen Übersicht.

Kassieren / Ausgabe (Frontend):

- [frontend/src/service/components/table/Zahlung.tsx:34-104](../../frontend/src/service/components/table/Zahlung.tsx) —
  Positionsliste (in "Meine"/"Andere" zu trennen).
- [frontend/src/service/components/table/Ausgabe.tsx](../../frontend/src/service/components/table/Ausgabe.tsx) —
  analog.
- [frontend/src/service/components/table/Zahlung.test.tsx](../../frontend/src/service/components/table/Zahlung.test.tsx) —
  Prior Art.

Stornieren / Umbuchen (Frontend):

- [frontend/src/service/components/table/TischHistorie.tsx:99-181](../../frontend/src/service/components/table/TischHistorie.tsx) —
  leitet heute `isFromUser={userId === item.userId}` ab;
  [TischHistorie.tsx:238-253,375-405](../../frontend/src/service/components/table/TischHistorie.tsx) —
  rein farblicher Eigen-Rand (`border-primary`) + "Du am …" (durch Besteller-Name
  zu ersetzen).
- [frontend/src/service/components/table/HistorieStornierungDrawer.tsx](../../frontend/src/service/components/table/HistorieStornierungDrawer.tsx),
  [HistorieUmbuchungDrawer.tsx](../../frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx) —
  Einstiegspunkte.

Schichtende (Admin Live-Dashboard):

- [backend/domain/reporting/reporting.go:9-16,84-94](../../backend/domain/reporting/reporting.go) —
  `UmsatzServicekraft`, `LiveReportingData` (um offene Arbeit pro Servicekraft
  zu erweitern).
- [backend/api/reporting/application/query.go:134-157](../../backend/api/reporting/application/query.go) —
  `GetLiveReporting` (zusätzlich Tisch-Sessions laden, Rollup, Merge per
  user_id).
- [backend/api/reporting/application/query_test.go](../../backend/api/reporting/application/query_test.go) —
  Prior Art.
- [frontend/src/admin/reporting/LiveReportingSection.tsx](../../frontend/src/admin/reporting/LiveReportingSection.tsx) —
  "Servicekräfte"-Tab.

Querschnitt / Regression:

- [backend/domain/kasse/direktverkauf_events.go](../../backend/domain/kasse/direktverkauf_events.go),
  [backend/domain/kasse/event_json_contract_test.go](../../backend/domain/kasse/event_json_contract_test.go) —
  Direktverkauf nutzt dieselbe `kasse`-Domäne; die umgestellten Konverter müssen
  weiter round-trippen.
- [docs/language.md](../../docs/language.md) — Begriff "Besteller" aufnehmen.

## Resolved decisions

- Besteller-Felder nur in `Position` (Projektion), nicht in `PositionEventData`;
  Name = eingefrorener Username aus dem Event-Umschlag.
- Konverter `to/fromPositionenEventData` werden von direkter Struct-Konvertierung
  auf explizites Feld-Mapping umgestellt (Besteller wird beim Schreiben ins Event
  weggelassen, beim Lesen nicht gesetzt).
- Projektion ohne Migration per Replay neu aufgebaut.
- Erledigt = keine eigenen ausstehenden UND keine eigenen unbezahlten
  Positionen; tischweite Auszahlung zählt nicht.
- Tischkarte zeigt **eine** Zahl: nicht-erledigte eigene Positionen als
  Vereinigung (ausstehend ∪ unbezahlt), plus Gesamtzahl offen.
- Keine neuen Routen; bestehende Endpunkte (`GetTischState`,
  `GetMeineTischeState`, `/get-eigene-uebersicht`, Live-Reporting) werden
  erweitert. Tisch-State-Abfragen erhalten die anfragende userID aus dem JWT.
- "eigene vs. fremde" client-seitig über `bestellerUserId`; Gesamtzustand aus der
  vollständigen Liste. Filter sind rein visuell.
- Phasenschnitt: 5 Phasen, Besteller-Tag als Fundament zuerst.

## Open questions / Risks

- **Rollup-Last:** Die persönliche Übersicht und das Live-Dashboard laden alle
  Tisch-Sessions der offenen Kassensitzung und aggregieren in Go. Für 5–50
  Tische unkritisch (PRD); keine JSONB-Aggregation in SQL.
- **Schichtübergabe-Semantik:** Sobald eine Kollegin eigene Positionen ausgibt
  oder kassiert, gilt für den Besteller "erledigt". Bewusst so gewählt; in den
  Deep-Module-Tests explizit abzudecken.

---

## Phase 1: Besteller-Tag durch Projektion bis Tisch-State

**User stories**: 20, 21 (Teil), 22 (Regression), 6 (Backend-Anteil)

### Context

- [backend/domain/kasse/bestellung.go:12-57](../../backend/domain/kasse/bestellung.go) —
  `Position`, `PositionEventData`, Konverter.
- [backend/domain/kasse/tisch_session.go:30-37,127-160](../../backend/domain/kasse/tisch_session.go) —
  `ApplyEvent` + accumulate/reduce.
- [backend/repository/kassenjournal_repo/repo.go:292-378,516-609](../../backend/repository/kassenjournal_repo/repo.go) —
  Projektion serialisieren/lesen, Rebuild.
- [backend/api/table/application/query.go:13-185](../../backend/api/table/application/query.go),
  [backend/api/table/http/query_handler.go:19-21](../../backend/api/table/http/query_handler.go) —
  Tisch-State + userID-Durchreichung.
- [frontend/src/service/table/Bestellung.ts:6-16](../../frontend/src/service/table/Bestellung.ts),
  [frontend/src/service/table/Tisch.ts:18-26](../../frontend/src/service/table/Tisch.ts) —
  Schemas.
- [backend/domain/kasse/event_json_contract_test.go](../../backend/domain/kasse/event_json_contract_test.go) —
  Event-Form unverändert verifizieren.

### What to build

Der Besteller fließt vom Bestell-Event bis in die Tisch-State-Antwort. `Position`
trägt `BestellerUserID`/`BestellerName`; beim Anwenden eines
`bestellung-aufgenommen`-Events tagt die Projektion jede neue Position mit dem
Akteur aus dem Event-Umschlag, und Zahlung/Storno/Ausgabe erhalten das Tag über
die Positions-ID. Die Konverter zwischen `Position` und `PositionEventData`
bilden die Felder explizit ab; die Event-Form und ihr JSON-Contract bleiben
unverändert (auch für den Direktverkauf). Die Projektion wird per Replay neu
aufgebaut. `GetTischState`/`GetMeineTischeState` erhalten die anfragende userID
aus dem JWT und liefern je offene Position `bestellerUserId`/`bestellerName`
sowie je Tisch ein Kennzeichen "für mich erledigt" (über die per-Tisch-Funktion
des Deep Module). Die Frontend-Schemas akzeptieren die neuen Felder; die UI
nutzt sie hier noch nicht (reines Plumbing).

### Acceptance criteria

- [x] Nach einer Bestellung tragen die offenen Positionen der Tisch-Session den
      korrekten Besteller (UserID + eingefrorener Username); mehrere Bestellungen
      verschiedener Servicekräfte an einem Tisch bleiben korrekt zugeordnet.
- [x] Zahlung, Stornierung und Ausgabe erhalten das Besteller-Tag der
      betroffenen Positionen.
- [x] Der JSON-Contract der Event-Form (`PositionEventData`) ist unverändert;
      Direktverkauf-Events round-trippen weiter (bestehende Contract-Tests grün).
- [x] `GetTischState`/`GetMeineTischeState` liefern je Position
      `bestellerUserId`/`bestellerName` und je Tisch ein korrektes
      "für mich erledigt" für die anfragende Servicekraft.
- [x] Die Projektion ist per Replay neu aufgebaut; Frontend-Schemas validieren
      die erweiterte Antwort; alle Backend-Tests grün.

---

## Phase 2: Persönliche Erledigt-Sicht (Servicekraft, alle Tische)

**User stories**: 1, 2, 3, 4, 5, 6

### Context

- [backend/domain/reporting/reporting.go:96-101](../../backend/domain/reporting/reporting.go) —
  `EigeneUebersicht`.
- [backend/api/reporting/application/query.go:115-132](../../backend/api/reporting/application/query.go) —
  `GetEigeneUebersicht` (Rollup über Tisch-Sessions ergänzen).
- [backend/repository/kassenjournal_repo/repo.go:593-609](../../backend/repository/kassenjournal_repo/repo.go) —
  `GetTischSessionsByKassensitzungNr` als Rollup-Eingabe.
- [frontend/src/service/components/EigeneUebersicht.tsx](../../frontend/src/service/components/EigeneUebersicht.tsx),
  [frontend/src/service/TableSelectionPage.tsx](../../frontend/src/service/TableSelectionPage.tsx) —
  Übersicht + Tischkarten.
- [frontend/src/service/table/Tisch.ts:38-44](../../frontend/src/service/table/Tisch.ts) —
  `EigeneUebersichtSchema`.

### What to build

Die Servicekraft sieht ihre eigene "Alles erledigt!"-Aussage über alle Tische
der offenen Kassensitzung, an denen sie bestellt hat (nicht nur Favoriten). Das
Rollup-Modul (DB-frei, in `kasse`) aggregiert die Tisch-Sessions der offenen
Kassensitzung pro Servicekraft zu: Liste offener Tische mit Anzahl eigener
ausstehender und unbezahlter Positionen sowie Gesamt-Kennzeichen "alles
erledigt". Der Service-Endpunkt `/get-eigene-uebersicht` wird um diese offene
eigene Arbeit erweitert (gefiltert auf die anfragende Servicekraft). Das Frontend
zeigt die persönliche Erledigt-Aussage und die Liste der eigenen offenen Tische
in der Übersicht; jede Tischkarte zeigt "X offen, davon Y von dir" (eine Zahl,
Vereinigung ausstehend ∪ unbezahlt); die Tisch-Detailseite zeigt das
per-Tisch-"für mich erledigt" aus Phase 1, getrennt vom Gesamtzustand.

### Acceptance criteria

- [x] Das Rollup berechnet pro Servicekraft die offenen Tische, die Anzahlen
      (ausstehend/unbezahlt) und "alles erledigt"; Schichtübergabe (fremde
      Ausgabe/Kassierung schließt eigene Arbeit ab) ist abgedeckt.
- [x] `/get-eigene-uebersicht` liefert die offene eigene Arbeit zusätzlich zum
      bestehenden Inhalt, bezogen auf alle Tische der offenen Kassensitzung.
- [x] Die Übersicht zeigt "Alles erledigt!" bzw. die Liste der eigenen offenen
      Tische.
- [x] Jede Tischkarte zeigt die Anzahl offener Positionen und davon eigene als
      eine Zahl; die Detailseite zeigt "für mich erledigt" getrennt vom
      Gesamtzustand.
- [x] Tests für Rollup (mehrere Tische/Servicekräfte) und für die erweiterte
      Übersicht grün.

---

## Phase 3: Eigene Positionen zuerst (Kassieren + Ausgabe)

**User stories**: 7, 8, 9, 10, 11

### Context

- [frontend/src/service/components/table/Zahlung.tsx:34-104](../../frontend/src/service/components/table/Zahlung.tsx) —
  Positionsliste Kassieren.
- [frontend/src/service/components/table/Ausgabe.tsx](../../frontend/src/service/components/table/Ausgabe.tsx) —
  Positionsliste Ausgabe.
- [frontend/src/service/components/table/Zahlung.test.tsx](../../frontend/src/service/components/table/Zahlung.test.tsx) —
  Prior Art.
- [frontend/src/lib/Auth.ts](../../frontend/src/lib/Auth.ts) —
  `AuthSingleton.userId` für die Eigen-Erkennung.

### What to build

Beim Kassieren und beim Ausgabe-Bestätigen werden die Positionen client-seitig
über `bestellerUserId` in "Meine" (oben, ausgeklappt) und "Andere" (eingeklappt,
über "Alle anzeigen" erreichbar) getrennt. Fremde Positionen zeigen den Namen der
bestellenden Servicekraft. Die eigentliche Buchung bleibt unverändert; es gibt
keine Sperre, die Trennung ist rein visuell. So landet der Umsatz in der Praxis
bei der richtigen Person, ohne neue Umsatzzuordnung.

### Acceptance criteria

- [x] Kassieren und Ausgabe zeigen eigene Positionen oben und ausgeklappt,
      fremde eingeklappt hinter "Alle anzeigen".
- [x] Fremde Positionen sind mit dem Besteller-Namen beschriftet.
- [x] "Alle anzeigen" blendet fremde Positionen ein; Buchung funktioniert für
      eigene wie fremde Positionen unverändert.
- [x] Tests: eigene vs. fremde getrennt, fremde erst nach "Alle anzeigen"
      sichtbar.

---

## Phase 4: Besteller-Name in Stornieren/Umbuchen-Historie

**User stories**: 12, 13, 14

### Context

- [backend/domain/kasse/bestellung.go:77-85](../../backend/domain/kasse/bestellung.go) —
  Historien-`Bestellung` (um `UserName` zu ergänzen).
- [backend/domain/kasse/historie.go:37-74](../../backend/domain/kasse/historie.go) —
  `GetHistorieFromEvents`/`buildBestellungFromEvent` (Username aus dem Umschlag
  übernehmen).
- [frontend/src/service/components/table/TischHistorie.tsx:99-405](../../frontend/src/service/components/table/TischHistorie.tsx) —
  Eigen-Rand + "Du am …" durch Besteller-Name ersetzen.
- [frontend/src/service/components/table/HistorieStornierungDrawer.tsx](../../frontend/src/service/components/table/HistorieStornierungDrawer.tsx),
  [HistorieUmbuchungDrawer.tsx](../../frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx) —
  Einstiegspunkte.

### What to build

Stornieren und Umbuchen erfolgen weiterhin zentral über die flache
Bestell-Historie. Die Historie liefert je Bestell-Eintrag den eingefrorenen
Username des Bestellers (Backend ergänzt `UserName` an der Historien-`Bestellung`
aus dem Event-Umschlag). Im Frontend trägt jeder Historien-Eintrag sichtbar den
Namen der bestellenden Servicekraft statt des bisherigen rein farblichen
Eigen-Rands. Nichts wird eingeklappt, damit die Serviceleitung fremde
Bestellungen ohne zusätzlichen Klick findet.

### Acceptance criteria

- [ ] Die Tisch-Historie liefert je Bestellung den eingefrorenen Besteller-Namen.
- [ ] Stornieren und Umbuchen zeigen die vollständige Historie flach (nichts
      eingeklappt), jeder Eintrag mit Besteller-Namen beschriftet.
- [ ] Tests: flache Darstellung und Besteller-Name je Eintrag.

---

## Phase 5: Schichtende-Prüfung im Admin Live-Dashboard

**User stories**: 15, 16, 17, 18, 19

### Context

- [backend/domain/reporting/reporting.go:9-16,84-94](../../backend/domain/reporting/reporting.go) —
  `UmsatzServicekraft`, `LiveReportingData`.
- [backend/api/reporting/application/query.go:134-157](../../backend/api/reporting/application/query.go) —
  `GetLiveReporting` (Tisch-Sessions laden, Rollup, Merge per user_id).
- [backend/repository/kassenjournal_repo/repo.go:593-609](../../backend/repository/kassenjournal_repo/repo.go) —
  `GetTischSessionsByKassensitzungNr`.
- [frontend/src/admin/reporting/LiveReportingSection.tsx](../../frontend/src/admin/reporting/LiveReportingSection.tsx) —
  "Servicekräfte"-Tab.
- [backend/api/reporting/application/query_test.go](../../backend/api/reporting/application/query_test.go) —
  Prior Art.

### What to build

`LiveReportingData` erhält die offene Arbeit pro Servicekraft aus dem
Rollup-Modul über die Tisch-Sessions der offenen Kassensitzung. Der bestehende
kassierte Umsatz pro Servicekraft und die offene Arbeit werden über die
Benutzer-ID zusammengeführt, sodass auch Personen mit offener Arbeit, aber ohne
kassierten Umsatz erscheinen. Der "Servicekräfte"-Tab zeigt pro Servicekraft
weiterhin den Umsatz und zusätzlich die offene eigene Arbeit (offene Tische mit
Anzahl ausstehender/unbezahlter eigener Positionen) oder einen "fertig"-Hinweis.

### Acceptance criteria

- [ ] `LiveReportingData` enthält die offene Arbeit pro Servicekraft, per user_id
      mit dem kassierten Umsatz zusammengeführt.
- [ ] Eine Servicekraft mit offener Arbeit, aber ohne kassierten Umsatz
      erscheint dennoch.
- [ ] Der "Servicekräfte"-Tab zeigt je Person den Umsatz und die offene Arbeit
      bzw. "fertig".
- [ ] Tests: Merge per user_id, offene Arbeit korrekt aus den Sessions, "fertig"
      bei keiner offenen eigenen Arbeit.

---

## Abschluss

- [ ] Begriff "Besteller" / "bestellende Servicekraft" in
      [docs/language.md](../../docs/language.md) aufgenommen.
- [ ] Out of Scope bestätigt: keine Übernahme/Neuzuordnung, keine
      besteller-basierte Umsatzzuordnung, Direktverkauf unberührt, historische
      Tagesabrechnung nicht erweitert.
