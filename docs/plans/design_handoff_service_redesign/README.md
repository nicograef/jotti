# Handoff: jotti Service-Bereich — Redesign

## Overview

Redesign des Service-Bereichs (`frontend/src/service`) des Kassensystems **jotti** (React 19, Vite, Tailwind CSS 4, shadcn/ui, TypeScript). Ziele laut Design-Review: Tempo unter Last (Vereinsfest, einhändig, Sonnenlicht), Fehlervermeidung (falscher Tisch, versehentliches Storno), Konsistenz zwischen Tischservice und Direktverkauf, klarere Hierarchie.

Betroffene Screens: **Meine Tische** (Dashboard), **Tisch-Detail** (Bestellen / Kassieren / Historie), **Direktverkauf**, plus die zugehörigen Drawer.

## About the Design Files

Die beiden `.dc.html`-Dateien in diesem Paket sind **Design-Referenzen in HTML** — Mockups, die Aussehen und Verhalten zeigen. Sie sind **kein Produktionscode** und werden nicht direkt übernommen. Aufgabe ist, die gezeigten Designs **im bestehenden jotti-Frontend nachzubauen** — mit den vorhandenen Mitteln: Tailwind-4-Utilities, den shadcn/ui-Komponenten unter `frontend/src/components/ui/`, den bestehenden Hooks (`use-mengen`, `use-action-submit`) und Backends. Keine neuen Bibliotheken nötig.

- `Jotti Service Review.dc.html` — Canvas mit Ist-Zustand (Frames 1a–1m, nur Referenz) und **Redesign (Frames 1n–1r — das ist die Ziel-UI)**. Jeder Redesign-Frame hat eine „Was ist neu"-Karte.
- `Implementierungsplan Redesign.dc.html` — 5-Phasen-Plan, jede Phase einzeln shippbar, mit konkreten Dateipfaden (unten wiederholt).

## Fidelity

**High-fidelity.** Die Redesign-Frames (1n–1r) verwenden exakt die bestehenden Design-Tokens aus `frontend/src/index.css` (olive/emerald-Theme, Inter, Radius 0.45rem). Farben, Größen, Abstände und Typografie sind pixelgenau gemeint — allerdings über die **vorhandenen CSS-Variablen** (`--primary`, `--muted-foreground`, …) umsetzen, nicht über hartkodierte oklch-Werte. Es werden **keine neuen Tokens** eingeführt.

## Implementierungsreihenfolge (5 Phasen, je 1 PR)

Phase 1 zuerst — alle weiteren bauen darauf auf. Danach sind Phase 2–5 unabhängig und parallelisierbar. Keine Backend-Änderung nötig (eine optionale Ausnahme in 4.4).

### Phase 1 — Fundament: Dock & Stepper (Referenz: Frame 1o)

1. **`ServiceDock.tsx` (neu)** — ersetzt die zwei gestapelten schwebenden Bodenleisten (StickyActionBar + fixierte TabsList) durch **ein** opakes Bottom-Dock: `border-t bg-background`, `padding: 12px 16px 16px`, innen vertikal: Aktionsbutton (56 px hoch, s. u.) und darunter die TabsList in voller Breite (`h-10`, Trigger `flex-1`). Safe-Area via `pb-[env(safe-area-inset-bottom)]`. Das Spezial-Padding `tabInhaltFreiraum` in `TablePage.tsx` entfällt; Content bekommt normales `padding-bottom` in Dock-Höhe.
2. **`Stepper.tsx` (neu)** — einheitlicher Mengen-Stepper überall: 44 px rund (`size-11 rounded-full`). **Plus dauerhaft primär** (bg `--primary`, Icon `--primary-foreground`). **Minus outline**; bei Menge 0: `border-dashed`, Icon 25 % Deckkraft, `disabled` (statt bisher 50 %-Icon-Opacity bei voll sichtbarem Button). Menge dazwischen: 17 px, bold, `tabular-nums`, 28 px breit zentriert. Ersetzt die Button-Paare in `ProductList.tsx`, `Zahlung.tsx` (PositionItem) und `PositionAuswahlListe.tsx` (bisher 32 px secondary).
3. **Saldo mit Label** — überall wo ein Tisch-Saldo steht: Label „OFFEN" (11 px, medium, uppercase, letter-spacing 0.04em, `--muted-foreground`) über dem Betrag (17–20 px, bold, `tabular-nums`). In `TablePage.tsx` Kopf entzerren: Titel „Tisch 12" 22 px semibold, Badge **unter** dem Titel (nicht daneben), Saldo-Block rechts.

Tests anpassen: `StickyActionBar.test.tsx`, `TablePage.test.tsx`, `PositionAuswahlListe.test.tsx`.

### Phase 2 — Bestellen: flache Variantenliste (Frame 1o)

1. **`ProductList.tsx` umbauen** — Aufklapp-Interaktion entfällt komplett (`expandedProducts`-State, Chevrons, Zähl-Badge löschen). Neu: Produktname als **Gruppenkopf** (13 px, semibold, `--muted-foreground`, margin-bottom 6 px), darunter jede Variante als sofort tippbare Zeile: `border rounded-lg` (8 px), `padding: 10px 14px`, links Variantenname (15 px medium) + Preis (14 px bold, `tabular-nums`, 10 px Abstand), rechts der Stepper aus Phase 1. Zeilen mit Menge > 0: `border-primary/50 bg-primary/[0.04]`. `Direktverkauf.tsx` nutzt dieselbe Liste → profitiert automatisch.
2. **Sticky Kategorie-Chips** — Zeile unter dem Tisch-Kopf, `sticky`, `border-b`, `padding: 6px 16px 12px`: Pills 36 px hoch, `rounded-full`, `padding 0 16px`, 14 px medium. Aktiv: bg `--foreground`, Text `--background`. Inaktiv: `border`. Empfehlung: Chips **filtern** die Liste (kein Scroll-Spy). Nur zeigen, wenn mehr als eine Kategorie Produkte hat. Reihenfolge wie `KategorieOrder`.

Tests: `Bestellung.test.tsx`, `Direktverkauf.test.tsx`.

### Phase 3 — Kassieren: Alles auswählen & Restbetrag (Frame 1p)

1. **„Alle auswählen"** — `use-mengen.ts` um `setAll(maxMengen: Record<Id, number>)` erweitern. In `Zahlung.tsx` oben ein voll breiter Button (44 px, `border-primary/50 bg-primary/5 text-primary` mit CircleCheck-Icon): „Alle 5 Positionen auswählen · 46,50 €". Wählt **nur eigene** Positionen voll aus (Fremd-Positionen bleiben bewusste Einzelaktion). Zweiter Tap leert die Auswahl.
2. **Positive Formulierung** — Positionszeilen: Titel 15 px medium; Unterzeile bei Auswahl: „2 von 2 ausgewählt · 9,00 €" (13 px, medium, emerald `--primary`-Abdunklung), ohne Auswahl: „1 unbezahlt · 3,50 €" (muted). Die invertierte Formulierung „noch X unbezahlt" entfällt. Ausgewählte Zeilen: `border-primary/60 bg-primary/5`.
3. **„Von anderen" sichtbar** — statt Toggle-Button: eingeklappte Gruppe mit Kopf „VON ANDEREN · 2" (12 px uppercase muted) + Chevron und einer Summen-/Namenszeile (13 px muted). Aufklappen zeigt die Zeilen mit Steppern und `von {bestellerName}`.
4. **Restbetrag im Dock** — über dem Kassieren-Button: „Nach dieser Zahlung noch offen" links (13 px muted), Betrag rechts (semibold, `tabular-nums`) = `saldoCents − auswahlSumme`.

Tests: `Zahlung.test.tsx`, `ZahlungDrawer.test.tsx`.

### Phase 4 — Historie & Fehlervermeidung (Frames 1q, 1k)

1. **Storno raus aus der Liste** (`TischHistorie.tsx`) — HistoryItem verliert alle Aktions-Buttons (X, ⇄, Auge). Ganze Zeile = Tap → Detail-Drawer, Chevron rechts als Affordance. Im **Detail-Footer**: Button-Zeile „Umbuchen" (outline, mit ⇄-Icon) + „Stornieren…" (outline, `border-destructive/40 text-destructive`, Ellipse signalisiert Folgeschritt), darunter „Schließen". Rollen-Gating (`canCancel`/`canRebook`) und Bedingungen (stornierbare/umbuchbare Positionen > 0) unverändert. Gleiches Muster in `DirektverkaufHistorie.tsx` (Drucken bleibt als Aktion im Detail).
2. **Scanbare Zeilen** — links Typ-Icon in 40-px-Kreis: Bestellung = Plus (bg `--primary`/10, Icon emerald-dunkel), Zahlung = Banknote (bg `--muted`, Icon muted), Umbuchung = ArrowRightLeft (bg `--muted`), Storno/Warenrücknahme = RotateCcw (bg `--destructive`/10, Icon `--destructive`). Titel = Typname (15 px medium; Umbuchung mit Quelle: „Umbuchung von Tisch 7"). Betrag rechts: 15 px bold `tabular-nums`, **+Beträge emerald, −Storno rot**, Zahlung neutral. Unterzeile: relative Zeit + Name + Kommentar in Anführungszeichen.
3. **`formatRelativeTime()`** neu in `lib/utils.ts` — „gerade eben" (< 1 min), „vor X min" (< 60), „vor X Std" (< 6 h), sonst absolut `18:42` bzw. mit Datum. Voller Timestamp bleibt im Detail-Drawer.
4. **Menschenlesbare Drawer-Titel** — „Bestellung · vor 32 min · Nico" statt `vorgangId.slice(0, 8)`; Detail-Unterzeile „Tisch 12 · 11.7.2026, 18:42". Betrifft `detailView()` in `TischHistorie.tsx`, `HistorieStornierungDrawer.tsx`, `HistorieUmbuchungDrawer.tsx` (Quell-Eintrag als Prop durchreichen statt nur `vorgangId`).
5. **Umbuchung entschärfen** (`HistorieUmbuchungDrawer.tsx`) — Ziel-Tisch **ohne stillen Default**: Placeholder „Ziel-Tisch wählen…", Submit disabled bis zur expliziten Wahl (den Fallback auf `zielTische[0].id` entfernen). Primärbutton von `secondary` auf `default` (primär). *Optional, nur mit Backend-Änderung:* Kommentarfeld wie beim Storno.

Tests: `TischHistorie.test.tsx`, `HistorieStornierungDrawer.test.tsx`, `HistorieUmbuchungDrawer.test.tsx`, `DirektverkaufHistorie.test.tsx`.

### Phase 5 — Dashboard & Direktverkauf (Frames 1n, 1r)

1. **`TableSelectionPage.tsx` / `EigeneUebersicht.tsx`** — die zwei Statistik-Karten werden **eine einzeilige Stat-Card** (bg `--muted`-nah, `rounded-xl`, `padding 12px 16px`, zwei Spalten mit 1-px-Trenner): Label 12 px muted, Wert 16 px bold + Betrag 13 px muted daneben. Die Karte „Deine offenen Tische" **entfällt ersatzlos** (Duplikat der Tischkarten).
2. **`MeinTischCard.tsx`** — Karte: `rounded-xl`, Ring wie Card, `padding 16px`, horizontal: Status-Punkt 10 px (rot = eigene offene, amber = nur fremde offene, grün = erledigt), Mitte: Tischname 16 px semibold + Unterzeile „5 offen · 3 von dir" (13 px muted), rechts: OFFEN-Label + Betrag, Chevron. Erledigte Tische: Unterzeile „Alles bezahlt" grün, Karte `opacity-75`, unter Abschnittskopf „ERLEDIGT · 1". Gruppierung: „NOCH OFFEN · 3" zuerst.
3. **Suche auf der Seite** — Suchfeld (44 px, Search-Icon, Placeholder „Tisch suchen — Name oder Nummer") direkt unter der Stat-Card; filtert die eigenen Tische und (bei 0 Treffern) verlinkt in den Alle-Tische-Drawer. „Alle Tische"-Button als **fixe Fußzeile** (48 px, outline, Table-Icon).
4. **`TischAuswahlDrawer.tsx`** — Favoriten-Stern auf 44 px Trefferfläche (`size-11`), klar getrennt von der Zeilen-Navigation; Sortierung: Favoriten → offener Saldo → Name.
5. **`Direktverkauf.tsx` / `DirektverkaufPage.tsx`** — die sticky Zahlungskarte entfällt. Produkte beginnen direkt unter den Kategorie-Chips. Neues Dock (Phase 1) mit „2 · Kassieren · 9,00 €" + Tabs (Verkaufen/Historie) unten. Neuer **`DirektverkaufDrawer`** nach dem Muster von `ZahlungDrawer.tsx`: Receipt, Feld „Erhalten" (EuroInput), Rückgeld-Zeile, Kommentar (optional), Buttons „Verkauf abschließen" / „Abbrechen". Fehlercode-Mapping (`kasse_nicht_geoeffnet`, `produkt_not_found`) zieht unverändert um; `verkaufId`-Logik (UUID pro Vorgang) bleibt.

Tests: `MeinTischCard.test.tsx`, `TischAuswahlDrawer.test.tsx`, `Direktverkauf.test.tsx`.

## Interactions & Behavior

- **Dock-Aktionsbutton** = DrawerTrigger (bestehendes Muster beibehalten: Guard in `onOpenChange` bei leerer Auswahl). Button 56 px: links Zähl-Pill (`bg-primary-foreground/20 rounded-full px-2 text-sm font-semibold tabular-nums`) + Label, rechts Summe bold `tabular-nums`.
- **Stepper**: Plus immer aktiv (bis `maxMenge` wo vorhanden), Minus disabled bei 0 (gestrichelt). Kein Layout-Shift beim Zustandswechsel.
- **Kategorie-Chips**: Filter-State lokal (`useState<Kategorie>`), Default = erste Kategorie mit Produkten.
- **Historie-Zeile**: einzige Geste = Tap → Detail. Destruktive Aktionen ausschließlich im Detail-Footer, Storno weiterhin mit Pflichtkommentar (min. 3 Zeichen) im Folgedrawer.
- **Loading/Empty/Error**: bestehende Skeletons, EmptyStates und `LadefehlerAlert` 1:1 auf die neuen Layouts übertragen; Tabs-Lock während `stateLoading || historieLoading` bleibt.
- Toasts (sonner) unverändert: „Bestellung wurde aufgenommen.", „Zahlung erfolgreich.", „Verkauf abgeschlossen.", „Bestellung umgebucht."

## State Management

Kein neues State-Konzept. Erweiterungen:
- `use-mengen.ts`: zusätzlich `setAll(mengen: Record<Id, number>)` (Phase 3).
- `ProductList`: `expandedProducts` entfällt; neu `aktiveKategorie` (Phase 2).
- `Zahlung`: abgeleitete Werte `auswahlSumme`, `restNachZahlung` (Phase 3).
- `HistorieUmbuchungDrawer`: `zielTischId` initial `null`, kein Fallback (Phase 4).
- Alle Backends (`TischBackend`, `DirektverkaufBackend`), react-query-Keys und Event-Payloads unverändert.

## Design Tokens

Unverändert aus `frontend/src/index.css` — im Code über Tailwind-Klassen/CSS-Variablen referenzieren:
- `--primary` emerald-700 `oklch(0.508 0.118 165.612)` (dark: emerald-500 `oklch(0.696 0.17 162.48)`)
- `--destructive` red-600 `oklch(0.577 0.245 27.325)` (dark: red-400)
- `--muted-foreground` olive-600 `oklch(0.53 0.031 107.3)`, `--border` olive-200 `oklch(0.93 0.007 106.5)`, `--input` olive-500
- Grün „erledigt": Tailwind green-600/700 (wie bisher in `MeinTischCard`/`EigeneUebersicht`)
- Radius: `--radius: 0.45rem`; Karten `rounded-xl`, Zeilen/Buttons `rounded-lg`/`rounded-md`, Pills/Stepper `rounded-full`
- Font: Inter Variable (bereits eingebunden); Beträge immer `tabular-nums`
- Mindest-Trefferfläche interaktiver Elemente: 44 px

## Assets

Nur **lucide-react** (bereits Dependency): ChevronLeft/Right/Down, User, Plus, Minus, X, Search, Table, CircleCheck, Banknote, ArrowRightLeft, RotateCcw, Printer. Keine Bilder, keine neuen Assets.

## Files

- `Jotti Service Review.dc.html` — Review-Canvas; **Frames 1n–1r = Ziel-Design**, Frames 1a–1m = Ist-Zustand + Befunde (Begründungen), Karten „Was ist neu" unter jedem Redesign-Frame.
- `Implementierungsplan Redesign.dc.html` — der 5-Phasen-Plan als Dokument.

## Vorgehen mit Claude Code

1. Dieses Paket ins jotti-Repo legen (z. B. `docs/design_handoff_service_redesign/`).
2. Pro Phase eine Session/PR: `Lies docs/design_handoff_service_redesign/README.md und setze Phase 1 um. Halte dich an bestehende Muster in frontend/src.`
3. Abnahme je Phase: App neben dem Review-Canvas öffnen und gegen Frames 1n–1r vergleichen; Tests der jeweiligen Phase grün.
