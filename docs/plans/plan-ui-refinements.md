# Plan: UI-Verfeinerungen (Dashboard, Produktstatistik, Autofokus, Tischsuche)

> Source PRD: `docs/prds/prd-ui-refinements.md`

## Goal

Mehrere zusammenhängende UI-Reibungspunkte beheben — mobil-first, für
ehrenamtliche Helfer unter Stress:

1. Admin-Übersicht: Stornierungen über die Produktstatistik ziehen; die
   Produktstatistik-Liste scrollbar deckeln.
2. Produkt-„Umsatz" auf dieselbe Ereignis-Grundlage wie „Ausgegeben" stellen
   (bestellbasiert), Beschreibungstext kürzen.
3. Automatisches Öffnen der Mobil-Tastatur beim Öffnen von Overlays zentral
   unterbinden; Admin-Dialoge scrollbar mit gepinnter Aktionsleiste machen.
4. Service-Hauptsuche über alle aktiven Tische statt nur Favoriten; zweite
   Suche im „Alle Tische"-Drawer entfernen.
5. Politur: Zählhilfe-Touch-Ziele, `formatEuro`-Konsistenz, Begriffe.

## Architectural decisions

Durable decisions across all phases:

- **Keine Schema-/Event-Änderungen.** Freeze-Disziplin gewahrt: nur ein
  Read-Query-Eingriff (`GetProduktStatistik`), keine Migration, keine
  Event-Format-Änderung, keine persistierten Daten berührt.
- **Backend als Single Source of Truth.** Die Umsatz-Definition ändert sich in
  der SQL-Query, nicht im Frontend. Frontend-DTOs/Typen (`ProduktStatistik`,
  `VarianteStatistik` mit `umsatzCents`/`ausgegebeneMenge`) bleiben unverändert
  — nur der gelieferte Zahlenwert ändert sich.
- **Fiskalische Größen bleiben kassenbasiert.** `GetUmsatzPositionszeilen`
  (Steuer-Aufschlüsselung, DSFinV-K, „Kassierter Umsatz"-KPI, Umsatz je
  Servicekraft) wird **nicht** angefasst. Nur die Produktstatistik
  (`GetProduktStatistik`) wechselt auf Bestellbasis.
- **Zentrale Overlay-Primitive.** Autofokus-Unterdrückung und
  Dialog-Scroll/Fußleiste werden in den geteilten `components/ui`-Primitiven
  gelöst, nicht je Aufrufstelle. Der Drawer ist bereits Vorbild
  (`max-h-[85dvh]` + einziger Scrollbereich `DrawerBody` + gepinnter Footer).
- **„Meine Tische" = Favoriten** bleibt bestehen; kein neues „zuletzt
  bearbeitet"-Konzept. Die Hauptsuche wechselt lediglich die Datenquelle beim
  Suchen auf „alle aktiven Tische".
- **Kanonische Begriffe** laut `docs/language.md`: Rolle = „Servicekraft";
  Person mit Zugang = „Benutzer"; Zugang/Credential = „Benutzername". Für die
  Helfer-Verwaltung wird die **Person** durchgängig „Helfer" genannt (Feld
  „Benutzername" bleibt).

## Inventory

**Dashboard / Reporting (Frontend):**
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Live-Übersicht;
  enthält den `VerkaufStatistik`-Block und darunter den Stornierungen-
  `Collapsible` (zu vertauschen). Nutzt `formatBediener`.
- `frontend/src/admin/reporting/VerkaufStatistik.tsx` — geteilte
  Produktstatistik-Tabelle; Beschreibungs-`<p>` und Listen-Wrapper
  (`<div className="overflow-hidden rounded-lg border">`) ohne Höhen-Cap.
- `frontend/src/admin/reporting/ReportingResults.tsx` — „Berichte & Export";
  rendert `VerkaufStatistik`, Stornierungen stehen dort bereits darüber.
- `frontend/src/admin/reporting/KassenberichtePage.tsx` — Seitenrahmen „Berichte
  & Export".

**Umsatz-Query (Backend):**
- `backend/sqlc/queries/reporting.sql — GetProduktStatistik` — der Umsatz-Term
  (`umsatz_cents`) und der WHERE-Event-Filter; Doc-Kommentar darüber.
- `backend/sqlc/dbgen/reporting.sql.go` — generiert; via `make sqlc`
  neu erzeugen, **nie** von Hand editieren.
- `backend/api/reporting/application/query.go — gruppiereProduktStatistik()` —
  summiert `umsatz_cents`/`ausgegebene_menge`; bleibt unverändert.
- `backend/repository/reporting_repo/repo_test.go`,
  `backend/api/reporting/application/query_test.go` — Prior-Art-Tests für die
  Produktstatistik.
- `backend/api/reporting/application/query_export_konsistenz_test.go` — prüft
  Steuer-Aufschlüsselung gegen DSFinV-K über `GetUmsatzPositionszeilen`;
  **nicht** betroffen (verifiziert).

**Overlay-Primitive (Frontend):**
- `frontend/src/components/ui/drawer.tsx` — Vorbild: `DrawerBody` als einziger
  Scrollbereich, `max-h-[85dvh]`, gepinnter Footer.
- `frontend/src/components/ui/dialog.tsx`, `frontend/src/components/ui/sheet.tsx`,
  `frontend/src/components/ui/alert-dialog.tsx` — Radix-Wrapper ohne
  `onOpenAutoFocus`-Override; `dialog.tsx`/`alert-dialog.tsx` ohne Höhen-Cap/
  internen Scroll.

**Service-Suche (Frontend):**
- `frontend/src/service/TableSelectionPage.tsx` — Hauptseite; `suche`-State
  filtert `useMeineTischeState()` (favoritengebunden).
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — „Alle Tische"-
  Drawer; eigener `suche`-State über `useAktiveTischeMitFavoriten()`.
- `frontend/src/service/table/hooks.ts — useMeineTischeState`,
  `useAktiveTischeMitFavoriten` — Datenquellen.
- `frontend/src/service/components/MeinTischCard.tsx` — Navigation zum Tisch.
- `frontend/src/service/TablePage.test.tsx`,
  `frontend/src/service/components/TischAuswahlDrawer.test.tsx` — Prior-Art-Tests.

**Politur (Frontend):**
- `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx` — Zähl-Eingaben `h-8 w-20`.
- `frontend/src/lib/utils.ts — formatEuro()` (geschütztes Leerzeichen),
  `formatCents()`.
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx`,
  `frontend/src/service/components/TischAuswahlDrawer.tsx` — nutzen
  `` `${formatCents(x)} €` `` mit normalem Leerzeichen.
- `frontend/src/admin/users/NewUserDialog.tsx`,
  `frontend/src/admin/users/EditUserDialog.tsx`,
  `frontend/src/admin/users/AdminUsersPage.tsx` — „Helfer"/„Benutzer" gemischt.
- `frontend/src/admin/reporting/utils.ts — formatBediener()` (+ Aufrufer
  `LiveReportingSection.tsx`, `StornoItem.tsx`, `StornoServicekraft.tsx`,
  `ReportingResults.tsx`) — internes Bezeichner-Naming „Bediener".
- `docs/language.md` — Begriffs-Quelle der Wahrheit.

## Resolved decisions

- **Umsatz-Grundlage:** Produkt-Umsatz nutzt dieselbe Ereignismenge/Gewichtung
  wie „Ausgegeben" (`bestellung-aufgenommen` +, `bestellung-korrigiert` −,
  `direktverkauf-getaetigt` +), gewichtet mit `einzelpreisCents × menge`. Die
  Event-Typen `zahlung-kassiert`, `stornierung-erteilt`, `direktverkauf-storniert`
  entfallen im Umsatz-Term und im WHERE-Filter der Query.
- **Umsatz-Reichweite:** beide Seiten (Übersicht **und** „Berichte & Export").
- **Autofokus:** zentral in allen Overlays (Drawer/Dialog/Sheet) unterdrückt;
  `AlertDialog` (Bestätigungen) bleibt button-fokussiert.
- **Tischsuche:** Hauptsuche über alle aktiven Tische, Treffer öffnet Tisch
  direkt; zweites Suchfeld im Drawer entfällt.
- **Begriff Person:** „Helfer" durchgängig in der Helfer-Verwaltung; „Benutzername"
  als Feld bleibt.
- **Beschreibungstext** der Produktstatistik: kurze Aussage, z. B. „Zahlen
  basieren auf Bestellungen".

## Open questions / Risks

- **Produkt-Umsatz ≠ Kassierter Umsatz bei Stornos.** Bewusst und in der PRD
  dokumentiert. Prior-Art-Tests, die für die Produktstatistik implizit
  Umsatz = kassiert annehmen, sind entsprechend auf die Bestellbasis
  anzupassen (siehe Phase 2). Der DSFinV-K/Steuer-Konsistenztest ist **nicht**
  betroffen (verifiziert, andere Query).
- **`formatBediener`-Umbenennung** ist ein rein internes Bezeichner-Refactoring
  (keine benutzer-sichtbare „Bediener"-Zeichenkette; Überschriften lauten bereits
  „Servicekraft"). Niedriges Risiko, aber mehrere Aufrufstellen.

---

## Phase 1: Dashboard umsortieren und Produktstatistik scrollbar

**User stories**: 1, 2

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx` — die zwei
  Geschwister-Blöcke (Produktstatistik-`<div>` und Stornierungen-`Collapsible`)
  stehen in einer `space-y-6`-Spalte; Reihenfolge tauschen (Stornierungen
  zuerst). Rein präsentational, kein geteilter State.
- `frontend/src/admin/reporting/VerkaufStatistik.tsx` — Listen-Wrapper erhält
  Höhen-Cap + `overflow-y-auto`, sodass die Liste innerhalb ihres Rahmens
  scrollt. Wirkt geteilt auch auf „Berichte & Export".
- `frontend/src/admin/reporting/ReportingResults.tsx` — hat Stornierungen
  bereits über `VerkaufStatistik`; profitiert nur vom Scroll-Cap, keine
  Umsortierung nötig.

### What to build

Auf der Live-Übersicht erscheint der Stornierungen-Block über der
Produktstatistik. Die Produktstatistik-Tabelle bekommt eine sinnvolle maximale
Höhe mit eigenem vertikalem Scroll, damit sie bei vielen Produkten die Seite
nicht überlängt — auf Übersicht und „Berichte & Export".

### Acceptance criteria

- [x] Auf der Admin-Übersicht steht der Stornierungen-Abschnitt oberhalb von
      „Verkäufe pro Produkt".
- [x] Die Produktstatistik-Liste scrollt innerhalb eines gedeckelten Bereichs
      statt die Seite unbegrenzt zu verlängern — auf beiden Seiten.
- [x] Auf „Berichte & Export" bleibt die bestehende Reihenfolge (Stornierungen
      vor Produktstatistik) erhalten.
- [x] `make lint` ist grün; keine Änderung an gelieferten Zahlen.

---

## Phase 2: Produkt-Umsatz auf Bestellbasis + Beschreibungstext

**User stories**: 3, 4

### Context

- `backend/sqlc/queries/reporting.sql — GetProduktStatistik` — Umsatz-Term auf
  dieselbe Ereignismenge/Gewichtung wie den Ausgegeben-Term umstellen; die drei
  kassenbasierten Event-Typen aus Umsatz-`CASE` und WHERE-Filter entfernen;
  Doc-Kommentar (Grundlage + Cross-Reference auf `GetUmsatzPositionszeilen`)
  aktualisieren.
- `make sqlc` regeneriert `backend/sqlc/dbgen/reporting.sql.go` (nicht manuell
  editieren).
- `backend/api/reporting/application/query.go — gruppiereProduktStatistik()` —
  unverändert (summiert nur die gelieferten Felder).
- `frontend/src/admin/reporting/VerkaufStatistik.tsx` — Beschreibungs-`<p>` auf
  kurze Aussage kürzen (z. B. „Zahlen basieren auf Bestellungen").
- `backend/repository/reporting_repo/repo_test.go`,
  `backend/api/reporting/application/query_test.go` — Produktstatistik-Tests auf
  die neue Bestellbasis anpassen/ergänzen.

### What to build

Der Produkt-Umsatz je Variante ist der Euro-Wert (zu Bestellzeit-Preisen) genau
der Portionen, die „Ausgegeben" zählt. Nachträgliche Stornierungen und
kassierte Zahlungen fließen nicht mehr in den Produkt-Umsatz ein. Die
Tabellenbeschreibung wird auf eine knappe Aussage reduziert. „Kassierter
Umsatz"-KPI, Umsatz je Servicekraft und Steuer-/DSFinV-K-Größen bleiben
kassenbasiert.

### Acceptance criteria

- [x] `GetProduktStatistik` liefert Umsatz je Variante = Σ
      `einzelpreisCents × menge` über `bestellung-aufgenommen` (+),
      `bestellung-korrigiert` (−), `direktverkauf-getaetigt` (+).
- [x] Eine nachträgliche kassenwirksame Stornierung reduziert den
      Produkt-Umsatz **nicht** (Test belegt das).
- [x] Die separate „Kassierter Umsatz"-KPI/Steueraufschlüsselung bleibt
      kassenbasiert unverändert (DSFinV-K-Konsistenztest weiterhin grün).
- [x] Der Beschreibungstext über der Tabelle ist auf eine kurze Aussage gekürzt.
- [x] `make sqlc` ausgeführt, `dbgen` regeneriert; `make verify` grün.

---

## Phase 3: Autofokus zentral unterdrücken + Admin-Dialoge scrollbar

**User stories**: 5, 6, 9

### Context

- `frontend/src/components/ui/dialog.tsx`,
  `frontend/src/components/ui/drawer.tsx`,
  `frontend/src/components/ui/sheet.tsx` — Radix `onOpenAutoFocus` beim Öffnen
  verhindern (kein Auto-Fokus auf das erste Feld), ohne Fokus-Trap/Escape/
  Fokusrückgabe zu brechen.
- `frontend/src/components/ui/dialog.tsx`,
  `frontend/src/components/ui/alert-dialog.tsx` — Höhen-Cap (analog Drawer
  `max-h-[85dvh]`) + intern scrollender Inhaltsbereich + gepinnte Fußleiste, so
  dass die Aktionsschaltflächen bei geöffneter Tastatur/viel Inhalt erreichbar
  bleiben. `drawer.tsx` (`DrawerBody`-Muster) als Vorbild.
- `frontend/src/service/components/table/BestellungAbschluss.tsx` — bestes
  Beispiel für den Autofokus-Test (erstes Feld ist das optionale Kommentarfeld).

### What to build

Kein Overlay öffnet beim Erscheinen automatisch die Mobil-Tastatur; der Nutzer
tippt selbst das gewünschte Feld an. Admin-`Dialog`/`AlertDialog` bekommen die
Drawer-Behandlung (Höhen-Cap, interner Scroll, gepinnte Fußleiste), damit die
primäre Aktion nie hinter der Tastatur oder unterhalb des Inhalts verschwindet.

### Acceptance criteria

- [x] Beim Öffnen von Drawer/Dialog/Sheet erhält **kein** Eingabefeld den Fokus
      (Test am Bestell-Drawer belegt: `document.activeElement` ist kein Input).
- [x] `AlertDialog`-Bestätigungen bleiben unverändert (button-fokussiert).
- [x] Admin-Formulardialoge (u. a. Zählhilfe, Kassenabschluss, Produkt-/Tisch-/
      Benutzer-Dialoge) haben einen Höhen-Cap mit internem Scroll und gepinnter
      Fußleiste; die Absende-Schaltfläche bleibt erreichbar.
- [x] Tastaturbedienung (Tab-Fokus-Trap, Escape, Fokusrückgabe) funktioniert
      weiter; `make lint` grün.

---

## Phase 4: Tischsuche über alle Tische vereinheitlichen

**User stories**: 7, 8

### Context

- `frontend/src/service/TableSelectionPage.tsx` — bei leerem Suchfeld weiterhin
  Favoriten anzeigen; sobald gesucht wird, über alle aktiven Tische filtern
  (Datenquelle `useAktiveTischeMitFavoriten()` nutzen) und einen Treffer direkt
  öffnen. Vorhandene „In allen Tischen suchen"-Notlösung entfällt damit.
- `frontend/src/service/table/hooks.ts — useMeineTischeState`,
  `useAktiveTischeMitFavoriten` — favoriten- vs. alle-aktiven-Datenquelle.
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — eigenes Suchfeld
  (`suche`-State + `Input`) entfernen; Drawer bleibt zum Durchblättern/
  Favorisieren.
- `frontend/src/service/components/MeinTischCard.tsx` — Navigationsmuster zum
  Tisch als Vorbild für Treffer-Navigation.
- `frontend/src/service/TablePage.test.tsx`,
  `frontend/src/service/components/TischAuswahlDrawer.test.tsx` — Prior-Art.

### What to build

Die Suche auf der Service-Hauptseite findet jeden aktiven Tisch (nicht nur
Favoriten) und öffnet ihn per Treffer direkt. Der „Alle Tische"-Drawer bleibt
für das Durchblättern und Favorisieren, hat aber kein eigenes Suchfeld mehr.

### Acceptance criteria

- [x] Ein Suchbegriff auf der Hauptseite, der auf einen **nicht favorisierten**
      aktiven Tisch passt, zeigt diesen als Treffer (Test belegt das).
- [x] Ein Treffer öffnet den Tisch direkt.
- [x] Bei leerem Suchfeld zeigt die Hauptseite unverändert die Favoriten
      („Meine Tische"), gruppiert in „Noch offen"/„Erledigt".
- [x] Der „Alle Tische"-Drawer hat kein zweites Suchfeld mehr, bleibt aber zum
      Durchblättern/Favorisieren funktionsfähig.
- [x] `make lint` grün; bestehende Service-Tabellentests grün.

---

## Phase 5: Politur — Touch-Ziele, Währungsformat, Begriffe

**User stories**: 10, 11, 12

### Context

- `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx` — Zähl-Eingaben von `h-8` auf
  ein mobiltaugliches Touch-Maß (mind. ~44 px Höhe) vergrößern, Layout kompakt
  halten.
- `frontend/src/lib/utils.ts — formatEuro()` — vorhandene Hilfe mit geschütztem
  Leerzeichen; in Admin-Code statt `` `${formatCents(x)} €` `` verwenden
  (`frontend/src/admin/kasse/KasseAbschliessenSection.tsx`,
  `frontend/src/service/components/TischAuswahlDrawer.tsx`).
- `frontend/src/admin/users/NewUserDialog.tsx`,
  `frontend/src/admin/users/EditUserDialog.tsx`,
  `frontend/src/admin/users/AdminUsersPage.tsx` — Person durchgängig „Helfer"
  (Trigger, Titel, Schaltfläche, Toast); „Benutzername"-Feld bleibt.
- `frontend/src/admin/reporting/utils.ts — formatBediener()` (+ Aufrufer) —
  internes Bezeichner-Refactoring zur Domänenkonsistenz („Servicekraft"); keine
  benutzer-sichtbare Änderung.
- `docs/language.md` — bei benutzer-sichtbaren Begriffsänderungen Ist/Soll
  angleichen.

### What to build

Kleinere Konsistenz- und Ergonomie-Politur: größere Zählhilfe-Eingaben, überall
`formatEuro` für Beträge (kein umbrechendes „€"), und einheitliche Begriffe
(Person = „Helfer" in der Verwaltung; interne `formatBediener`-Benennung an
„Servicekraft" angleichen).

### Acceptance criteria

- [x] Zählhilfe-Eingaben erfüllen ein mobiltaugliches Touch-Maß (≥ ~44 px Höhe).
- [x] Admin-seitige Beträge nutzen `formatEuro`; „€" bricht nicht mehr allein um.
- [x] Die Helfer-Verwaltung nennt die Person durchgängig „Helfer" (Trigger,
      Titel, Schaltfläche, Toast); das Zugangsfeld heißt weiter „Benutzername".
- [x] `formatBediener` ist zur Domänenkonsistenz umbenannt; keine
      benutzer-sichtbare „Bediener"-Zeichenkette verbleibt.
- [x] `docs/language.md` spiegelt geänderte benutzer-sichtbare Begriffe;
      `make lint` grün.
