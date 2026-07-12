# Plan: Service-Redesign (Design-Handoff)

> Source PRD: `docs/plans/design_handoff_service_redesign/README.md` (Handoff) plus `Jotti Service Review.dc.html` (Review-Canvas, Ziel-Frames 1n-1r) und `Implementierungsplan Redesign.dc.html` (5-Phasen-Vorschlag).

## Goal

Den Service-Bereich (`frontend/src/service`) nach den Redesign-Frames 1n-1r umbauen: Tempo unter Last (flache Variantenliste, Alle-auswählen, ein Bottom-Dock), Fehlervermeidung (Storno raus aus der Historien-Liste, Umbuchung ohne stillen Ziel-Tisch-Default), Konsistenz (ein Stepper-Stil, Direktverkauf folgt dem Tischservice-Muster) und klarere Hierarchie (Saldo mit Label, scanbare Historie). Zusätzlich, über den Handoff hinaus vom Nutzer beauftragt: ein optionales Benutzerkommentar für Umbuchungen (Phase 6) und die Behebung der Prüf-Drawer-Befunde 1h/1i (Phase 7).

Umsetzung mit vorhandenen Mitteln: Tailwind-4-Utilities, shadcn/ui-Komponenten unter `frontend/src/components/ui/`, bestehende Hooks und Backends. Keine neuen Bibliotheken, keine neuen Design-Tokens, Farben nur über CSS-Variablen (`--primary`, `--muted-foreground`, ...), nie hartkodierte oklch-Werte.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **Neue Komponenten**: `frontend/src/service/components/ServiceDock.tsx` (eine opake Bottom-Fläche für Aktionsbutton + TabsList, ersetzt die zwei schwebenden Leisten) und `frontend/src/service/components/Stepper.tsx` (einheitlicher 44-px-Mengen-Stepper). Beide liegen in `service/components/`, weil sie von Tischservice und Direktverkauf gemeinsam genutzt werden.
- **Dock-Komposition per Portal**: `ServiceDock` wird von der Seite (innerhalb von `<Tabs>`) gerendert und stellt neben der TabsList einen Aktions-Slot bereit. Der Aktionsbutton bleibt in den Drawer-Komponenten (er braucht deren Mengen-State und den Radix-`DrawerTrigger`-Kontext) und rendert per React-Portal in den Slot; React-Context bleibt über Portale erhalten. Die Slot-Mechanik (Context + `createPortal`) kapselt `ServiceDock`. Der Slot nimmt beliebigen Inhalt auf, nicht nur einen einzelnen Button — Phase 3 rendert darüber zusätzlich die Restbetrag-Zeile über dem Kassieren-Button.
- **Dock auf allen Viewports**: Das bisherige `isMobile`-Sonderlayout in `TablePage.tsx` (Tabs oben auf Desktop, fixiert auf Mobile) entfällt; das Dock gilt überall. Das Spezial-Padding `tabInhaltFreiraum` entfällt, Content bekommt normales `padding-bottom` in Dock-Höhe.
- **Umbenennung**: `StickyActionBar.tsx` wird zu `DockActionButton.tsx` (Positionierung raus, Button-Inhalt bleibt); Props (`label`, `anzahl`, `summeCents`, `disabled`) bleiben.
- **Neue Utility**: `formatRelativeTime()` in `frontend/src/lib/utils.ts` ("gerade eben" < 1 min, "vor X min" < 60, "vor X Std" < 6 h, sonst absolut `18:42` bzw. mit Datum). Die Anzeige aktualisiert sich nur bei Re-Render/Refetch (kein Ticker) — bewusst akzeptiert.
- **Hook-Erweiterung**: `useMengen` in `frontend/src/hooks/use-mengen.ts` bekommt zusätzlich `setAll(mengen: Record<K, number>)`.
- **Historie-Drawer-Datenfluss**: `HistorieStornierungDrawer` und `HistorieUmbuchungDrawer` erhalten den vollständigen Quell-Eintrag (`Quelle = Bestellung | Umbuchung`) als Prop statt nur `vorgangId` + `positionen`.
- **Event-Erweiterung statt neuer Version**: `bestellung-umgebucht:v1` behält sein `kommentar`-Feld unverändert (Richtungs-Autotext "Umbuchung auf/von Tisch X", immer gesetzt; konserviert den Tischnamen zum Buchungszeitpunkt). Das Benutzerkommentar kommt als zusätzliches optionales Feld `benutzerKommentar` (Max 100, JSON mit `omitempty`) additiv in die v1-Event-Data; Alt-Events parsen mit Leerstring, Events ohne Benutzerkommentar serialisieren byte-identisch wie heute. Das ist eine vom Nutzer ausdrücklich autorisierte Ausnahme von der Freeze-Regel "Änderungen additiv als neue Event-Version (`:vN`), nie in-place" (AGENTS.md; der Freeze gilt seit der Erstinstallation 2026-07-07, nicht erst ab v1.0) — die erste Abweichung überhaupt, alle 12 Event-Typen stehen auf `:v1`. Durch `omitempty` bleibt die Ausnahme maximal eng: Nur Events, die tatsächlich ein Benutzerkommentar tragen, weichen vom heutigen Format ab. Der Contract-Guard (`backend/domain/kasse/event_json_contract_test.go`) erzwingt neue Felder auf bestehenden Typen nicht (fehlender JSON-Key parst als Zero-Value, bestehende Fixtures bleiben grün) — die Fixture-Erweiterung um beide Fälle (mit/ohne Feld) ist der einzige Schutz und Pflichtteil von Phase 6.
- **API**: Keine neuen Endpunkte. Einzige Request-Änderung: `service/bestellung-umbuchen` bekommt ein optionales Feld `kommentar` (Phase 6). Das Historie-Response-DTO `umbuchung` bekommt zusätzlich `benutzerKommentar`. Alle Endpunkte bleiben POST-only.
- **Design-Tokens**: unverändert aus `frontend/src/index.css`; Beträge immer `tabular-nums`; Mindest-Trefferfläche interaktiver Elemente 44 px; Icons nur lucide-react.

## Inventory

Verifiziert gegen den Code (Stand: main, 2026-07-12):

- `frontend/src/service/TablePage.tsx` — `TablePage()`, `tabInhaltFreiraum`-Konstante, `isMobile`-Verzweigung, Tabs `order`/`payment`/`history`, `tabsLocked = stateLoading || historieLoading`, Kopf mit Titel + Badge inline + Saldo ohne Label.
- `frontend/src/service/components/table/StickyActionBar.tsx` — Aktionsleiste, als `DrawerTrigger` genutzt in `BestellungDrawer.tsx` und `ZahlungDrawer.tsx` (Guard bei leerer Auswahl liegt in deren `onOpenChange`).
- `frontend/src/service/components/table/ProductList.tsx` — `ProductList`, `ProductListSkeleton`, internes `VariantItem` (48-px-Buttonpaar), `expandedProducts`-State, Chevrons, Zähl-Badge; iteriert `KategorieOrder` aus `frontend/src/service/product/Produkt.ts` (`KategorieOrder`, `KategorieLabels`). Genutzt von `Bestellung.tsx` und `Direktverkauf.tsx`.
- `frontend/src/service/components/table/Zahlung.tsx` — internes `PositionItem`, Formulierung "noch X unbezahlt", `alleAnzeigen`-Toggle für Fremd-Positionen, Trennung eigene/fremde über `position.bestellerUserId === AuthSingleton.userId`, `useMengen` gedeckelt über `unbezahlteMengen`; erhält `tisch` (inkl. `saldoCents`) bereits als Prop.
- `frontend/src/service/components/PositionAuswahlListe.tsx` — `PositionAuswahlListe`, Typ `AuswahlPosition`; 32-px-secondary-Buttonpaar; genutzt von `HistorieStornierungDrawer`, `HistorieUmbuchungDrawer`, `DirektverkaufStornoDrawer`.
- `frontend/src/hooks/use-mengen.ts` — `useMengen(max?)` mit `mengen`, `add`, `remove`, `reset`; kein `setAll`.
- `frontend/src/service/components/table/TischHistorie.tsx` — `HistoryItem` mit Aktions-Buttons (X, ArrowRightLeft, Eye), `detailView()`, Titel mit `vorgangId.slice(0, 8)`, Gating über `AuthSingleton.canCancel`/`canRebook` und `stornierbarePositionen`/`umbuchbarePositionen`.
- Historien-Datenmodell (`frontend/src/service/table/`): `Bestellung` (`aufgenommenAm`, `gesamtPreisCents`), `Zahlung` (`kassiertAm`, `gesamtZahlungCents`), `Stornierung` (`storniertAm`, `gesamtStornierungCents`, `barRueckgabe`), `Umbuchung` (`umgebuchtAm`, `gesamtCents`, `quellTischId`, `zielTischId`). Kein einheitliches Timestamp-/Betragsfeld; Aufrufer verzweigen über `art`. Alle vier Typen tragen `userName`; `Umbuchung` trägt zusätzlich `kommentar` (Autotext), und das `umbuchung`-DTO liefert ihn heute schon aus — Phase 4 (Titel, "· Name") baut genau darauf und bleibt deshalb rein Frontend.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx` — stiller Fallback `zielTische[0].id`, Submit-Button `variant="secondary"`; Props: `backend`, `tisch`, `vorgangId`, `positionen`, `onClose` + Erfolgs-Callback — vom Quell-Eintrag kommen nur `vorgangId` + `positionen` an.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` — Pflichtkommentar min. 3 Zeichen; gleiche Prop-Struktur, vom Quell-Eintrag nur `vorgangId` + `positionen`.
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx` — sticky Zahlungskarte (`sticky top-14`, nur Gesamt-Zeile, kein `Receipt`), Fehlercode-Mapping `kasse_nicht_geoeffnet`/`produkt_not_found`, `verkaufId` per `crypto.randomUUID()`.
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` — Drucken- und Storno-Buttons in der Liste; `DirektverkaufStornoDrawer` erhält bereits den vollen `verkauf`-Eintrag (Vorbild für den Drawer-Datenfluss in Phase 4).
- `frontend/src/service/TableSelectionPage.tsx` — nutzt `useMeineTischeState()` (liefert `TischSession[]`) und `useEigeneUebersicht()`.
- `frontend/src/service/components/EigeneUebersicht.tsx` — `EigeneUebersichtKarten` (zwei Statistik-Karten) + internes `EigeneArbeitSicht` ("Deine offenen Tische" / "Alles erledigt!").
- `frontend/src/service/components/MeinTischCard.tsx` — `countOffenePositionen(state)` liefert `anzahlOffen`/`anzahlEigeneOffen` aus `unbezahltePositionen`.
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — Favoriten-Stern `w-7` (~28 px), nur Namensfilter, keine Sortierung; Favoriten serverseitig (`favoritHinzufuegen`/`favoritEntfernen`); Datentyp `AktiverTischMitFavorit` (`id`, `name`, `saldoCents`, `istFavorit`).
- `frontend/src/service/components/table/ZahlungDrawer.tsx` — Receipt, Feld "inklusive Trinkgeld" (vor "Erhalten"), Rückgeld-/Trinkgeld-Zeilen aus `calculateZahlungsbetraege` (`drawerUtils.ts`), `KommentarField`, Reset bei Close; Vorbild für den neuen `DirektverkaufDrawer`.
- `frontend/src/lib/utils.ts` — `formatCents`, `parseCents`, `formatPositionName`, `cn`; kein Datums-/Relativzeit-Formatter (Datumsanzeige heute ad hoc per `toLocaleString('de-DE')`).
- Backend Umbuchung: `backend/domain/kasse/tisch_session_events.go` — `BestellungUmgebuchtV1Data` (mit `kommentar`-JSON-Feld), `NewBestellungUmgebuchtEvents(...)` mit `quellKommentar`/`zielKommentar`; `backend/api/kasse/tischgeschaeft/application/command.go` — `Command.BestellungUmbuchen(...)`, `buildUmbuchungKommentar()`; `backend/api/kasse/tischgeschaeft/http/command_handler.go` — `bestellungUmbuchenRequest`/`bestellungUmbuchenSchema`; `backend/api/kasse/tischgeschaeft/http/query_handler.go` — DTO `umbuchung`, `toUmbuchung()`; `backend/api/kasse/tischgeschaeft/application/query.go` — `Query` mit `TischRepo`, `GetTischHistorie()`.
- DSFinV-K: `backend/api/fiskal/dsfinvk/mapper.go` — Case `EventTypeBestellungUmgebuchtV1` schreibt `data.Kommentar` als `notiz`; `tischnamen`-Map ist dort bereits vorhanden.
- Frontend Umbuchung: `frontend/src/service/table/Umbuchung.ts` — `BestellungUmbuchenSchema` (ohne `kommentar`); `frontend/src/service/table/TischBackend.ts` — `bestellungUmbuchen()`.

## Resolved decisions

Aus der Klärungsrunde mit dem Nutzer (2026-07-11):

- Umbuchungs-Kommentar wird umgesetzt (Phase 6). Prüfauftrag des Nutzers beantwortet: keine neue Event-Version nötig. Entscheidung aus der Nachfrage (2026-07-12): Der Richtungs-Autotext bleibt immer gesetzt im bestehenden `kommentar`-Feld; das Benutzerkommentar kommt als zusätzliches optionales Feld `benutzerKommentar` additiv in die v1-Event-Data. Kein String-Mischen, kein Parsen bei der Anzeige, das Journal bleibt selbstbeschreibend. Bewusste Inkonsistenz: bei den übrigen Events ist `kommentar` der Benutzertext; hier macht der Feldname `benutzerKommentar` den Unterschied explizit.
- Die Prüf-Drawer-Befunde 1h/1i werden als Zusatzphase (Phase 7) umgesetzt, nach Ermessen ohne Ziel-Design. Die Mengen-Korrektur im Prüf-Drawer (Befund 1h) bleibt bewusst draußen (eigenständige Interaktionsänderung ohne Design-Referenz).
- Der 5-Phasen-Schnitt des Handoffs wird 1:1 übernommen; Phase 6 und 7 kommen dazu.
- Kategorie-Chips filtern die Liste (Empfehlung des Handoffs), kein Scroll-Spy.
- Dock auf allen Viewports statt Mobile-Sonderfall (vereinfacht `TablePage`, konsistent mit Mobile-first-Zielgruppe).
- Sortierung im `TischAuswahlDrawer` erfolgt clientseitig: reine Darstellungsreihenfolge bereits vollständiger Daten, keine Filterung/Aggregation (Regel "Backend ist Single Source of Truth für Daten-Filterung" unberührt). Gleiche Begründung deckt die Tisch-Suche auf "Meine Tische" (Phase 5): clientseitige Live-Eingrenzung bereits vollständig geladener eigener Tische (Präzedenz: der bestehende Namensfilter im `TischAuswahlDrawer`).

## Open questions / Risks

- Die Portal-Slot-Mechanik des `ServiceDock` ist die strukturell heikelste Stelle: Radix-`DrawerTrigger` muss über das Portal hinweg funktionieren (React-Context tut das, in Phase 1 explizit testen, auch das Drawer-Open-Verhalten).
- E2E-Tests (`e2e/`) matchen teils auf UI-Struktur und Texte; pro Phase gegenprüfen und anpassen.
- Bewusst offen gelassen (Kandidaten für ein Folge-Review, nicht Teil dieses Plans): Mengen-Korrektur im Prüf-Drawer (Befund 1h), Positions-Zusammenfassung in Direktverkauf-Historien-Titeln (Befund 1f), Validierungshinweis-Timing im Storno-Drawer (Befund 1j), Vereinheitlichung der Vorauswahl-Muster (Umbuchung startet mit voller Menge, Kassieren/Storno leer — Konsistenz-Kernbefund, Befund 1k), Erklärung des negativen Saldos im Alle-Tische-Drawer (Befund 1g), Dark-Mode-Destructive-Kontrast (separat getrackt).

## Querschnitt (gilt für jede Phase)

- `make check` grün; die je Phase genannten Tests angepasst; betroffene E2E-Flows geprüft.
- Dark Mode mitprüfen (Befund 1l/1m: Listen-Items ggf. `bg-card` statt nur Rahmen).
- A11y: bestehende `aria-label` übernehmen, neue interaktive Elemente (Stepper, Chips, Alle-auswählen, Dock) beschriften.
- Loading/Empty/Error-Zustände (Skeletons, EmptyStates, `LadefehlerAlert`) 1:1 auf die neuen Layouts übertragen; Toasts unverändert.
- Abnahme: App neben dem Review-Canvas öffnen und gegen die Frames vergleichen (Screenshots headless per Playwright/Chromium).

---

## Phase 1: Fundament — Dock, Stepper, Saldo-Label

Referenz: Frame 1o. Muss zuerst kommen; Phase 2-5 bauen darauf auf.

### Context

- `frontend/src/service/TablePage.tsx — TablePage()` — Kopf entzerren, Tabs ins Dock, `tabInhaltFreiraum` und `isMobile`-Verzweigung entfernen.
- `frontend/src/service/components/table/StickyActionBar.tsx` — wird zu `DockActionButton.tsx`, rendert per Portal in den Dock-Slot.
- `frontend/src/service/components/table/BestellungDrawer.tsx`, `ZahlungDrawer.tsx` — nutzen den Button als `DrawerTrigger`; Guard-Logik bleibt.
- `frontend/src/service/components/table/ProductList.tsx — VariantItem`, `Zahlung.tsx — PositionItem`, `frontend/src/service/components/PositionAuswahlListe.tsx` — bekommen den neuen Stepper.
- `frontend/src/service/components/MeinTischCard.tsx` — Saldo-Label auch hier (vollständiger Karten-Umbau erst in Phase 5).

### What to build

`ServiceDock.tsx` (neu): ein opakes Bottom-Dock (`border-t bg-background`, Padding 12px 16px 16px, Safe-Area via `pb-[env(safe-area-inset-bottom)]`), innen vertikal Aktions-Slot (Button 56 px hoch) und darunter die TabsList in voller Breite (`h-10`, Trigger `flex-1`). `TablePage` rendert das Dock innerhalb von `<Tabs>`; der Tabs-Lock-Hinweis bleibt. Content-Bottom-Padding in Dock-Höhe ersetzt `tabInhaltFreiraum`.

`DockActionButton.tsx` (aus `StickyActionBar.tsx`): eigene Fixierung raus, rendert per Portal in den Dock-Slot. Button-Anatomie: links Zähl-Pill (`bg-primary-foreground/20 rounded-full px-2 text-sm font-semibold tabular-nums`) + Label, rechts Summe bold `tabular-nums`.

`Stepper.tsx` (neu): 44 px rund (`size-11 rounded-full`). Plus dauerhaft primär (bg `--primary`, Icon `--primary-foreground`), Minus outline; bei Menge 0 `border-dashed`, Icon 25 % Deckkraft, `disabled`. Menge dazwischen 17 px bold `tabular-nums`, 28 px breit zentriert. Kein Layout-Shift beim Zustandswechsel. Ersetzt die Buttonpaare in `ProductList.tsx`, `Zahlung.tsx` und `PositionAuswahlListe.tsx`.

Saldo mit Label: Label "OFFEN" (11 px, medium, uppercase, letter-spacing 0.04em, `--muted-foreground`) über dem Betrag (17-20 px, bold, `tabular-nums`). `TablePage`-Kopf entzerren: Titel 22 px semibold, Badge unter dem Titel, Saldo-Block rechts. Gleiches Label am Saldo in `MeinTischCard`.

### Acceptance criteria

- [x] Auf Tisch-Detail existiert genau eine Bodenfläche (Dock) mit Aktionsbutton über der TabsList; keine zwei übereinander schwebenden Leisten mehr, keine `tabInhaltFreiraum`-Konstante, keine `isMobile`-Verzweigung in `TablePage`.
- [x] Drawer öffnen weiterhin über den Dock-Button (Portal + `DrawerTrigger` verifiziert), Guard bei leerer Auswahl greift unverändert.
- [x] Letzte Listenzeile ist über dem Dock vollständig sichtbar und antippbar (Content-Padding stimmt, auch mit Safe-Area-Inset).
- [x] Alle drei Stepper-Einsatzorte (`ProductList`, `Zahlung`, `PositionAuswahlListe`) nutzen `Stepper` mit identischem Verhalten; Minus bei Menge 0 ist `disabled` und gestrichelt.
- [x] Tisch-Kopf zeigt Titel, Badge darunter, rechts Saldo mit "OFFEN"-Label; `MeinTischCard`-Saldo trägt dasselbe Label.
- [x] Tests angepasst: `StickyActionBar.test.tsx` (umbenannt zu `DockActionButton.test.tsx`), `TablePage.test.tsx`, `PositionAuswahlListe.test.tsx`; dazu `Bestellung.test.tsx`, `Zahlung.test.tsx`, `Direktverkauf.test.tsx` (der Stepper ersetzt die Buttonpaare schon in dieser Phase) und `MeinTischCard.test.tsx` (Saldo-Label); `make check` grün.

---

## Phase 2: Bestellen — flache Variantenliste und Kategorie-Chips

Referenz: Frame 1o. Nach Phase 1; unabhängig von 3-7.

### Context

- `frontend/src/service/components/table/ProductList.tsx — ProductList, ProductListSkeleton` — kompletter Umbau; `Direktverkauf.tsx` nutzt dieselbe Liste und profitiert automatisch.
- `frontend/src/service/product/Produkt.ts — KategorieOrder, KategorieLabels` — Chip-Reihenfolge und -Beschriftung.

### What to build

Aufklapp-Interaktion entfällt komplett (`expandedProducts`-State, Chevrons, Zähl-Badge löschen). Produktname wird Gruppenkopf (13 px, semibold, `--muted-foreground`, margin-bottom 6 px); darunter jede Variante als sofort tippbare Zeile: `border rounded-lg`, Padding 10px 14px, links Variantenname (15 px medium) + Preis (14 px bold, `tabular-nums`), rechts der Stepper aus Phase 1. Zeilen mit Menge > 0: `border-primary/50 bg-primary/[0.04]`.

Sticky Kategorie-Chips als Zeile über der Liste (`sticky` unter dem App-Header, `border-b`, Padding 6px 16px 12px): Pills 36 px hoch, `rounded-full`, 14 px medium; aktiv bg `--foreground` / Text `--background`, inaktiv `border`. Chips filtern die Liste (lokaler State `aktiveKategorie`, Default: erste Kategorie mit Produkten, Reihenfolge `KategorieOrder`). Nur zeigen, wenn mehr als eine Kategorie Produkte hat. `ProductListSkeleton` an das neue Layout angleichen.

### Acceptance criteria

- [x] "1 Bier 0,5 l" kostet genau einen Tap (Plus auf der Variantenzeile); kein Aufklapp-Schritt mehr, `expandedProducts` existiert nicht mehr.
- [x] Kategorie-Chips filtern die Liste, erscheinen nur bei mehr als einer belegten Kategorie und bleiben beim Scrollen sichtbar.
- [x] Direktverkauf zeigt dieselbe flache Liste ohne eigene Anpassung.
- [x] Zeilen mit Menge > 0 sind visuell markiert; Skeleton passt zum neuen Layout.
- [x] Tests angepasst: `Bestellung.test.tsx`, `Direktverkauf.test.tsx`; `make check` grün.

---

## Phase 3: Kassieren — Alle auswählen und Restbetrag

Referenz: Frame 1p. Nach Phase 1; unabhängig von 2, 4-7.

### Context

- `frontend/src/hooks/use-mengen.ts — useMengen` — bekommt `setAll`.
- `frontend/src/service/components/table/Zahlung.tsx — Zahlung, PositionItem` — Auswahl-UI, Formulierungen, Fremd-Positionen-Gruppe.
- `frontend/src/service/components/table/ZahlungDrawer.tsx` — Dock-Aktionsbereich (Restbetrag-Zeile + Kassieren-Button).

### What to build

`useMengen` um `setAll(mengen: Record<K, number>)` erweitern (setzt die Auswahl exakt auf die übergebenen Mengen). In `Zahlung` oben ein voll breiter Button (44 px, `border-primary/50 bg-primary/5 text-primary`, CircleCheck-Icon): "Alle 5 Positionen auswählen · 46,50 €". Er wählt nur eigene Positionen voll aus (`setAll` mit den unbezahlten Mengen der eigenen Positionen); Fremd-Positionen bleiben bewusste Einzelaktion. Sind alle eigenen Positionen voll ausgewählt, leert ein zweiter Tap die gesamte Auswahl (`reset`).

Positive Formulierung in den Positionszeilen: Titel 15 px medium; Unterzeile bei Auswahl "2 von 2 ausgewählt · 9,00 €" (13 px, medium, `--primary`-Abdunklung), ohne Auswahl "1 unbezahlt · 3,50 €" (muted). "noch X unbezahlt" entfällt. Ausgewählte Zeilen: `border-primary/60 bg-primary/5`.

Fremd-Positionen statt Toggle-Button als eingeklappte Gruppe: Kopf "VON ANDEREN · 2" (12 px uppercase muted) + Chevron, darunter eingeklappt eine Summen-/Namenszeile (13 px muted); aufgeklappt die Zeilen mit Steppern und "von {bestellerName}".

Restbetrag im Dock über dem Kassieren-Button: links "Nach dieser Zahlung noch offen" (13 px muted), rechts der Betrag (semibold, `tabular-nums`) als `saldoCents − auswahlSumme`. Die abgeleiteten Werte `auswahlSumme` und `restNachZahlung` entstehen in `Zahlung` (`tisch.saldoCents` liegt dort als Prop vor) und gehen als Props an `ZahlungDrawer`, der die Zeile zusammen mit dem `DockActionButton` durch das Portal in den Dock-Slot rendert (der Slot nimmt beliebigen Inhalt auf, siehe Architectural decisions).

### Acceptance criteria

- [x] Ein Tap auf "Alle X Positionen auswählen" wählt alle eigenen unbezahlten Positionen voll aus; Button-Text zeigt Anzahl und Summe; zweiter Tap leert die Auswahl.
- [x] Fremd-Positionen erscheinen als sichtbare eingeklappte Gruppe mit Summe; aufgeklappt einzeln wählbar mit Besteller-Name.
- [x] Positionszeilen formulieren positiv ("X von Y ausgewählt") und markieren Auswahl farblich; die Formulierung "noch X unbezahlt" kommt nicht mehr vor.
- [x] Über dem Kassieren-Button steht der Restbetrag nach dieser Zahlung und rechnet live mit.
- [x] Tests angepasst: `Zahlung.test.tsx`, `ZahlungDrawer.test.tsx`; neue Fälle für `setAll` (über die bestehenden Komponententests); `make check` grün.

---

## Phase 4: Historie — scanbare Zeilen, Storno raus aus der Liste

Referenz: Frames 1q, 1k. Nach Phase 1; unabhängig von 2, 3, 5, 7. Rein Frontend.

### Context

- `frontend/src/service/components/table/TischHistorie.tsx — HistoryItem, detailView(), Details` — Zeilen-Anatomie, Detail-Drawer, Aktions-Verlagerung.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`, `HistorieUmbuchungDrawer.tsx` — bekommen den vollen Quell-Eintrag als Prop; Umbuchung ohne stillen Default.
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` — gleiches Zeilen-/Detail-Muster; `DirektverkaufStornoDrawer` als Datenfluss-Vorbild.
- `frontend/src/lib/utils.ts` — neue Utility `formatRelativeTime()`.

### What to build

Historien-Zeilen verlieren alle Aktions-Buttons (X, ArrowRightLeft, Eye); die ganze Zeile ist die einzige Geste (Tap öffnet den Detail-Drawer), Chevron rechts als Affordance. Zeilen-Anatomie: links Typ-Icon in 40-px-Kreis (Bestellung = Plus auf `--primary`/10, Zahlung = Banknote auf `--muted`, Umbuchung = ArrowRightLeft auf `--muted`, Stornierung/Warenrücknahme = RotateCcw auf `--destructive`/10). Titel = Typname (15 px medium); für Umbuchungszeilen dient `item.kommentar` als Titel (enthält je Event-Seite exakt "Umbuchung von/auf Tisch X"; das bleibt auch nach Phase 6 so, das Benutzerkommentar ergänzt dann nur die Unterzeile). Betrag rechts 15 px bold `tabular-nums`, Plus-Beträge emerald, Storno-Minus rot, Zahlung neutral. Unterzeile: relative Zeit + Name + Kommentar in Anführungszeichen.

`formatRelativeTime()` in `lib/utils.ts` mit Unit-Tests in `utils.test.ts`; voller Timestamp bleibt im Detail-Drawer.

Detail-Drawer: menschenlesbarer Titel "Bestellung · vor 32 min · Nico" statt `vorgangId.slice(0, 8)`; Unterzeile "Tisch 12 · 11.7.2026, 18:42". Im Footer Button-Zeile "Umbuchen" (outline, ArrowRightLeft-Icon) + "Stornieren…" (outline, `border-destructive/40 text-destructive`), darunter "Schließen". Rollen-Gating (`canCancel`/`canRebook`) und Bedingungen (stornierbare/umbuchbare Positionen > 0) unverändert. `HistorieStornierungDrawer` und `HistorieUmbuchungDrawer` erhalten den vollen Quell-Eintrag als Prop und titeln ebenso menschenlesbar.

`HistorieUmbuchungDrawer` entschärfen: Ziel-Tisch ohne stillen Default (Placeholder "Ziel-Tisch wählen…", `zielTischId` initial `null`, Fallback auf `zielTische[0].id` entfernen), Submit disabled bis zur expliziten Wahl, Primärbutton von `secondary` auf `default`.

`DirektverkaufHistorie` folgt demselben Muster: Zeile = Tap → Detail-Drawer, Drucken und "Stornieren…" als Aktionen im Detail-Footer (Gating unverändert), gleiche Zeilen-Anatomie und relative Zeit.

### Acceptance criteria

- [x] Historien-Zeilen (Tisch und Direktverkauf) haben keine Inline-Aktions-Buttons mehr; Tap öffnet das Detail; Umbuchen/Stornieren/Drucken liegen ausschließlich im Detail-Footer mit unverändertem Gating.
- [x] Zeilen zeigen Typ-Icon im farbigen Kreis, farbcodierten Betrag und relative Zeit; Kommentare erscheinen in Anführungszeichen in der Unterzeile.
- [x] Kein UUID-Fragment mehr in Drawer-Titeln; Detail und Folge-Drawer titeln "Typ · relative Zeit · Name".
- [x] Umbuchung: ohne explizite Ziel-Tisch-Wahl ist Submit disabled; es gibt keinen vorbelegten Ziel-Tisch; Primärbutton ist primär.
- [x] `formatRelativeTime` deckt die Grenzen (< 1 min, < 60 min, < 6 h, absolut) per Unit-Test ab.
- [x] Tests angepasst: `TischHistorie.test.tsx`, `HistorieStornierungDrawer.test.tsx`, `HistorieUmbuchungDrawer.test.tsx`, `DirektverkaufHistorie.test.tsx`, `utils.test.ts`; `make check` grün.

---

## Phase 5: Dashboard und Direktverkauf

Referenz: Frames 1n, 1r. Nach Phase 1 (Dock) und sinnvoll nach Phase 2 (Chips/Liste im Direktverkauf); unabhängig von 3, 4, 6, 7.

### Context

- `frontend/src/service/TableSelectionPage.tsx`, `frontend/src/service/components/EigeneUebersicht.tsx — EigeneUebersichtKarten, EigeneArbeitSicht` — Stat-Card-Konsolidierung, Suche, Fußzeile.
- `frontend/src/service/components/MeinTischCard.tsx — countOffenePositionen()` — Karten-Umbau, Status-Punkt, Gruppierung.
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — Stern-Trefferfläche, Sortierung.
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx`, `frontend/src/service/DirektverkaufPage.tsx` — Zahlungskarte raus, Dock rein, neuer Drawer.
- `frontend/src/service/components/table/ZahlungDrawer.tsx` — Muster für den neuen `DirektverkaufDrawer` (ohne Trinkgeld-Feld).

### What to build

Meine Tische: Die zwei Statistik-Karten werden eine einzeilige Stat-Card (bg nahe `--muted`, `rounded-xl`, Padding 12px 16px, zwei Spalten mit 1-px-Trenner; Label 12 px muted, Wert 16 px bold, Betrag 13 px muted daneben). Die Karte "Deine offenen Tische" (`EigeneArbeitSicht`) entfällt ersatzlos. Suchfeld (44 px, Search-Icon, Placeholder "Tisch suchen — Name oder Nummer") direkt unter der Stat-Card; filtert die eigenen Tische clientseitig, bei 0 Treffern Verweis in den Alle-Tische-Drawer. "Alle Tische"-Button als fixe Fußzeile (48 px, outline, Table-Icon; gleiche Dock-Fläche wie Phase 1, ohne Tabs).

`MeinTischCard`: `rounded-xl`, Ring wie Card, Padding 16px, horizontal: Status-Punkt 10 px (rot = eigene offene, amber = nur fremde offene, grün = erledigt), Mitte Tischname 16 px semibold + Unterzeile "5 offen · 3 von dir" (13 px muted), rechts OFFEN-Label + Betrag (aus Phase 1) und Chevron. Erledigte Tische (keine unbezahlten Positionen): Unterzeile "Alles bezahlt" grün, Karte `opacity-75`. Gruppierung mit Abschnittsköpfen "NOCH OFFEN · 3" zuerst, dann "ERLEDIGT · 1".

`TischAuswahlDrawer`: Favoriten-Stern auf 44 px Trefferfläche (`size-11`), klar getrennt von der Zeilen-Navigation; Sortierung Favoriten → offener Saldo (absteigend) → Name.

Direktverkauf: Die sticky Zahlungskarte entfällt; Produkte beginnen direkt unter den Kategorie-Chips. `DirektverkaufPage` bekommt das Dock aus Phase 1 mit Aktionsbutton ("2 · Kassieren · 9,00 €") und den Tabs Verkaufen/Historie unten. Neuer `DirektverkaufDrawer` (`frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx`) nach dem Muster von `ZahlungDrawer`, ohne Trinkgeld-Feld: Receipt, Feld "Erhalten" (EuroInput), Rückgeld-Zeile, Kommentar (optional), Buttons "Verkauf abschließen" / "Abbrechen". Fehlercode-Mapping (`kasse_nicht_geoeffnet`, `produkt_not_found`) zieht unverändert um; `verkaufId` wird beim Öffnen des Drawers erneuert (Muster `BestellungDrawer`: neue ID pro Drawer-Öffnung). Das ändert den Erneuerungs-Trigger bewusst — heute bei Mount und nach Erfolg, künftig pro Prüf-Anlauf; der serverseitige Idempotenz-Mechanismus bleibt unverändert.

### Acceptance criteria

- [ ] Meine Tische zeigt eine einzeilige Stat-Card, darunter Suche, darunter gruppierte Tischkarten (Noch offen zuerst); die Duplikat-Karte existiert nicht mehr; "Alle Tische" liegt fix in der Fußzeile.
- [ ] `MeinTischCard` zeigt Status-Punkt, Unterzeile, OFFEN-Label mit Betrag und Chevron; erledigte Tische sind gedimmt unter eigenem Abschnittskopf.
- [ ] Suche filtert live; bei 0 Treffern führt ein Verweis in den Alle-Tische-Drawer.
- [ ] Favoriten-Stern hat 44 px Trefferfläche; Drawer-Sortierung ist Favoriten → offener Saldo → Name.
- [ ] Direktverkauf: Produkte ab dem ersten Viewport sichtbar; Kassieren läuft über Dock-Button + `DirektverkaufDrawer` mit Receipt, Erhalten, Rückgeld, Kommentar; Verkauf und Fehlerfälle funktionieren wie zuvor.
- [ ] Tests angepasst: `MeinTischCard.test.tsx`, `TischAuswahlDrawer.test.tsx`, `Direktverkauf.test.tsx`; `make check` grün.

---

## Phase 6: Umbuchung mit Benutzerkommentar (Backend + Frontend)

Über den Handoff hinaus vom Nutzer beauftragt (Klärungsrunde). Nach Phase 4 (baut auf dem umgebauten `HistorieUmbuchungDrawer` und der neuen Zeilen-Anatomie auf).

### Context

- `backend/domain/kasse/tisch_session_events.go — BestellungUmgebuchtV1Data, bestellungUmgebuchtV1DataSchema, NewBestellungUmgebuchtEvents(), buildUmbuchungFromEvent()` — neues Event-Feld, Durchreichung in den Historien-Eintrag.
- `backend/domain/kasse/umbuchung.go — Umbuchung, umbuchungSchema` — Domain-Struct zieht nach.
- `backend/domain/kasse/event_json_contract_test.go` — Contract-Guard um das additive Feld erweitern; Alt-Fixture ohne Feld muss weiter parsen.
- `backend/api/kasse/tischgeschaeft/application/command.go — Command.BestellungUmbuchen(), buildUmbuchungKommentar()` — neuer Parameter; Autotext-Erzeugung bleibt unverändert.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go — bestellungUmbuchenRequest, bestellungUmbuchenSchema` — Request-Feld.
- `backend/api/kasse/tischgeschaeft/http/query_handler.go — umbuchung, toUmbuchung()` — DTO-Feld.
- `backend/api/fiskal/dsfinvk/mapper.go — Case EventTypeBestellungUmgebuchtV1` — Notiz-Komposition.
- `frontend/src/service/table/Umbuchung.ts — BestellungUmbuchenSchema, Umbuchung`, `frontend/src/service/table/TischBackend.ts — bestellungUmbuchen()`, `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx`, `TischHistorie.tsx` — Frontend-Seite.

### What to build

Durchgängiger optionaler Kommentar für Umbuchungen, konsistent zum Storno, aber ohne Pflicht. Der Richtungs-Autotext bleibt unangetastet: `kommentar` trägt weiterhin immer "Umbuchung auf/von Tisch X" (konserviert den Tischnamen zum Buchungszeitpunkt), das Benutzerkommentar kommt als eigenes Feld dazu.

Event: `BestellungUmgebuchtV1Data` bekommt additiv `benutzerKommentar` (JSON `benutzerKommentar,omitempty`, zog `Max(100)` ohne `Required`; Alt-Events parsen mit Leerstring, Events ohne Benutzerkommentar serialisieren byte-identisch wie heute). `NewBestellungUmgebuchtEvents` erhält einen zusätzlichen Parameter `benutzerKommentar` (gleicher Text für beide Event-Seiten); `quellKommentar`/`zielKommentar` bleiben. `buildUmbuchungFromEvent` und das Domain-Struct `Umbuchung` reichen das Feld in den Historien-Eintrag durch. Contract-Guard-Fixtures um beide Fälle (mit/ohne Feld) erweitern und das neue Feld explizit pinnen — der Guard erzwingt das nicht von selbst (fehlender Key parst als Zero-Value, bestehende Fixtures blieben auch ohne Erweiterung grün), die Erweiterung ist der einzige Schutz.

Command und HTTP: `bestellungUmbuchenRequest` und `bestellungUmbuchenSchema` bekommen ein optionales `kommentar` (Max 100); `Command.BestellungUmbuchen` erhält den Parameter und schreibt ihn als `benutzerKommentar` ins Event. `buildUmbuchungKommentar` und die Autotext-Logik bleiben unverändert.

DSFinV-K: Die Notiz des Umbuchungs-Bons wird `data.Kommentar`, ergänzt um `data.BenutzerKommentar`, falls vorhanden (einfache Verkettung mit "; "). Alt-Events exportieren dadurch byte-identisch wie heute. Feldlänge ist geklärt: `BON_NOTIZ` erlaubt 255 Zeichen (DSFinV-K 2.4, `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/`); der Autotext ist auf 100 Runen gekappt (`buildUmbuchungKommentar`), das Benutzerkommentar auf 100 — mit Separator maximal 202 Zeichen, keine Kürzung nötig.

Query und Frontend: Das `umbuchung`-DTO und der Frontend-Typ `Umbuchung` bekommen `benutzerKommentar`. `BestellungUmbuchenSchema` und `bestellungUmbuchen()` reichen `kommentar` durch; `HistorieUmbuchungDrawer` bekommt ein optionales `KommentarField` (wie Bestellung/Zahlung). In `TischHistorie` bleibt der Zeilentitel `item.kommentar` (Autotext, aus Phase 4); das Benutzerkommentar erscheint in Anführungszeichen in der Unterzeile und im Detail-Drawer.

### Acceptance criteria

- [ ] Eine Umbuchung mit Kommentar trägt in beiden Event-Seiten unverändert den Richtungs-Autotext in `kommentar` und den Benutzertext in `benutzerKommentar`; ohne Kommentar fehlt der Key im Event-JSON (`omitempty`), das Event ist byte-identisch zum heutigen Format.
- [ ] Alt-Events ohne das Feld parsen weiter (Replay und Historie); der Contract-Guard deckt beide Fälle ab.
- [ ] Der DSFinV-K-Export trägt für Umbuchungs-Bons Autotext plus Benutzerkommentar in der Notiz ("; "-Verkettung, maximal 202 von 255 erlaubten Zeichen); Alt-Events exportieren unverändert.
- [ ] Historien-Zeilen titeln Umbuchungen weiter mit dem Autotext; das Benutzerkommentar erscheint in Anführungszeichen in Unterzeile und Detail.
- [ ] Der Umbuchungs-Drawer bietet ein optionales Kommentarfeld; Backend- und Frontend-Validierung begrenzen auf 100 Zeichen.
- [ ] Tests angepasst: Backend Event-/Command-/Handler-/Mapper-Tests, `event_json_contract_test.go`, `HistorieUmbuchungDrawer.test.tsx`, `TischHistorie.test.tsx`; `make verify` grün (Integrationstests wegen Backend-Änderung).

---

## Phase 7: Prüf-Drawer-Feinschliff (Befunde 1h/1i)

Über den Handoff hinaus vom Nutzer beauftragt; kein Ziel-Design, Umsetzung nach Ermessen entlang der Befunde. Logisch unabhängig von allen anderen Phasen, überschneidet sich aber auf Dateiebene mit Phase 3 (`ZahlungDrawer.tsx`) und Phase 5 (`DirektverkaufDrawer.tsx`) — nicht parallel zu diesen beiden in getrennten Worktrees umsetzen, sondern danach oder koordiniert.

### Context

- `frontend/src/service/components/table/BestellungDrawer.tsx` — Titel/Description (Befund 1h).
- `frontend/src/service/components/table/ZahlungDrawer.tsx` — Feld-Reihenfolge, Trinkgeld-Label, Rückgeld-Prominenz (Befund 1i).
- `frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx` — sofern Phase 5 schon gelandet ist, gleiche Kopf- und Rückgeld-Behandlung dort.

### What to build

Tisch-Bestätigung prominent: In `BestellungDrawer` und `ZahlungDrawer` wird der Tischname das dominante Kopf-Element (Vorgangstyp als kleine muted Zeile darüber, Tischname darunter groß, ca. 22 px semibold). Die Standard-Prosa-Description ("Überprüfe deine … vor dem Absenden.") entfällt; die a11y-Beschreibung übernimmt der knappe Kopf (DrawerDescription mit dem Tischnamen bzw. der Vorgangszeile, kein Prosa-Satz).

Feld-Reihenfolge in `ZahlungDrawer`: Standardfall zuerst, also "Erhalten" vor dem Trinkgeld-Feld. Das Trinkgeld-Feld bekommt ein verständliches Label statt "inklusive Trinkgeld" (Zielbetrag-Semantik klar benennen, z. B. "Zahlbetrag inkl. Trinkgeld", Wortlaut bei Umsetzung gegen `docs/language.md` prüfen) und einen kurzen Hilfetext, wofür es dient.

Rückgeld prominent: Die Rückgeld-Zeile wird der größte Betrag im Sheet (ca. 20 px bold, `tabular-nums`), da sie laut vorgelesen wird. Gleiches Muster im `DirektverkaufDrawer`, falls Phase 5 bereits gelandet ist (sonst dort bei der Erstellung berücksichtigen).

Nicht enthalten (bewusst, siehe Resolved decisions): Mengen-Korrektur im Prüf-Drawer.

### Acceptance criteria

- [ ] In beiden Prüf-Drawern ist der Tischname das größte Kopf-Element; die generische Prosa-Description existiert nicht mehr; keine a11y-Warnungen (Description-Slot sinnvoll belegt).
- [ ] Im Zahlungs-Drawer steht "Erhalten" vor dem Trinkgeld-Feld; das Trinkgeld-Feld erklärt sich über Label/Hilfetext.
- [ ] Die Rückgeld-Zeile ist der visuell größte Betrag im Sheet; Live-Berechnung unverändert.
- [ ] Bestell- und Kassier-Flow funktionieren unverändert (Guard, Reset bei Close, Idempotenz der Bestellung).
- [ ] Tests angepasst: `BestellungDrawer.test.tsx`, `ZahlungDrawer.test.tsx`; `make check` grün.
