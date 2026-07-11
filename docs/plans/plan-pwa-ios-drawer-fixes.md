# Plan: PWA/iOS-Drawer-Fixes (Praxistest 2026-07-09)

> Source PRD: n/a — abgeleitet aus `praxistest-2026-07-09-bugreport.md` (Fassung 2, auditiert)

## Goal

Die drei im Praxistest belegten Fehlerbilder beheben:

1. **Bug 1:** vaul-Drawer bricht in der iOS-Standalone-PWA (Flicker, versetzte
   `fixed`-Elemente, wirkungslose Taps) — Ursache: unkompensierter `scrollTo(0,0)`-Scroll-Lock
   von vaul 1.1.2 in `display-mode: standalone`; upstream offen (vaul#505), kein Update
   verfügbar.
2. **Bug 2:** Drawer-Inhalt ist nicht scrollbar; bei langer Positionsliste oder nach
   Trinkgeld-/Erhalten-Eingabe gerät der Primär-Button („Kassieren") unter die
   Viewport-Unterkante — Zahlung/Bestellung nicht abschließbar.
3. **Bug 3:** Während des Absendens fehlt wahrnehmbares Pending-Feedback über den
   Button-Spinner hinaus; der Drawer bleibt währenddessen schließbar.

## Architectural decisions

Durable decisions that apply across all phases:

- **Drawer-Engine**: `frontend/src/components/ui/drawer.tsx` wird in place auf die
  Radix-`Dialog`-Primitives aus dem `radix-ui`-Paket umgeschrieben (dasselbe Muster wie
  `frontend/src/components/ui/sheet.tsx`), als Bottom-Sheet gestylt. Die **exportierte API
  bleibt unverändert**: `Drawer`, `DrawerTrigger`, `DrawerClose`, `DrawerContent`,
  `DrawerHeader`, `DrawerFooter`, `DrawerTitle`, `DrawerDescription` — alle sechs Verwender
  behalten ihre Imports. Die Dependency `vaul` wird aus `frontend/package.json` entfernt.
  `sheet.tsx` bleibt unangetastet (generisches Seiten-Sheet, aktuell ohne Verwender).
- **Layout-Vertrag des Bottom-Sheets**: `DrawerContent` ist `fixed inset-x-0 bottom-0 flex
  flex-col` mit `max-h-[85dvh]` (dvh statt vh), `rounded-t-xl` und
  `pb-[env(safe-area-inset-bottom,0px)]`. Neu exportiert wird **`DrawerBody`** — der
  **einzige** Scrollbereich (`min-h-0 overflow-y-auto`). `DrawerHeader` und `DrawerFooter`
  sind direkte Flex-Kinder von `DrawerContent` und damit immer sichtbar. Innere Bereiche mit
  eigenen `max-h`-Caps (Receipt `max-h-[40dvh]`, Tisch-Liste `max-h-[60vh]`) verlieren diese —
  keine verschachtelten Scrollcontainer.
- **Dismissal-UX**: Kein Swipe-to-dismiss, kein Drag-Handle mehr (Radix Dialog kann kein
  Drag; ein Handle ohne Funktion wäre irreführend). Schließen über „Abbrechen"-Button,
  Backdrop-Tap und Escape. Während eines laufenden Submits (`loading`) ist der Drawer nicht
  schließbar.
- **Keine neuen Dependencies, kein Dependency-Update**: `radix-ui` ist bereits vorhanden;
  vaul entfällt ersatzlos. Ein allgemeines Frontend-Update und ein shadcn-Reinstall sind
  bewusst nicht Teil dieses Plans (behebt keinen der Bugs, Reinstall würde lokale
  Anpassungen überschreiben).
- **Verifikation ausschließlich automatisiert** (Vitest + Playwright); Bug 2 bekommt eine
  Viewport-Regression im `mobile-service`-Projekt der E2E-Suite.

## Inventory

- `frontend/src/components/ui/drawer.tsx — DrawerContent()` — vaul-Wrapper, wird ersetzt;
  enthält den obsoleten `data-vaul-no-drag`-Workaround und das Handle-Div.
- `frontend/src/components/ui/sheet.tsx — SheetContent()` — Vorlage für das Radix-Muster
  (`import { Dialog as SheetPrimitive } from "radix-ui"`, `side="bottom"`-Styling).
- Verwender (alle über die `Drawer*`-API):
  - `frontend/src/service/components/table/BestellungDrawer.tsx — BestellungDrawer()`
  - `frontend/src/service/components/table/ZahlungDrawer.tsx — ZahlungDrawer()`
  - `frontend/src/service/components/TischAuswahlDrawer.tsx — TischAuswahlDrawer()` —
    einziger Verwender der vaul-Prop `direction="bottom"` (entfällt); Tisch-Liste mit
    eigenem `overflow-y-auto max-h-[60vh]`.
  - `frontend/src/service/components/table/HistorieStornierungDrawer.tsx — HistorieStornierungDrawer()`
  - `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx — HistorieUmbuchungDrawer()`
  - `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx — DirektverkaufStornoDrawer()`
- `frontend/src/service/components/table/Receipt.tsx — Receipt()` — interner Scrollbereich
  `max-h-[40dvh]` (entfällt zugunsten `DrawerBody`).
- `frontend/src/service/components/PositionAuswahlListe.tsx — PositionAuswahlListe()` —
  Positionsliste der Storno-/Umbuchungs-Drawer (wandert in `DrawerBody`).
- `frontend/src/service/components/table/StickyActionBar.tsx — StickyActionBar()` —
  `DrawerTrigger asChild`-Muster (Trigger klont das Wurzel-`div`); muss mit Radix-Trigger
  identisch funktionieren.
- `frontend/src/hooks/use-action-submit.ts — useActionSubmit()` — liefert `loading` für
  Bug 3.
- Unit-Tests, die Drawer öffnen/bedienen: `frontend/src/service/components/table/Bestellung.test.tsx`,
  `Zahlung.test.tsx`, `TischHistorie.test.tsx`, `frontend/src/service/components/TischAuswahlDrawer.test.tsx`,
  `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx`,
  `frontend/src/service/TablePage.test.tsx`.
- E2E (Playwright, `e2e/`): Projekt `mobile-service` (Hochkant-Viewport) deckt alle
  Service-Flows ab — `bestellen-kassieren.spec.ts`, `tischservice-teilzahlung.mobile.spec.ts`,
  `kassieren-fehlerpfade.spec.ts`, `stornierung-serviceleitung.mobile.spec.ts`,
  `umbuchung.mobile.spec.ts`, `direktverkauf-storno.mobile.spec.ts`.
- Referenz: `praxistest-2026-07-09-bugreport.md` — Fehlerbilder, Mechanismen, Quellen.

## Resolved decisions

- Alle 6 Drawer wechseln auf die Radix-Basis; vaul wird entfernt (Nutzerentscheidung).
  Begründung: vaul#505 offen ohne Fix, 1.1.2 ist die letzte Version, Prop-Tuning wäre
  hypothesenbasiert und nur am Gerät verifizierbar; ein Drawer-System statt zwei.
- In-place-Rewrite von `drawer.tsx` statt Umbau aller Verwender auf `Sheet`-Imports:
  minimaler Diff, die Domänen-Komponentennamen (`…Drawer`) bleiben stimmig.
- Kein Dependency-Update, kein shadcn-Reinstall (Nutzerentscheidung; behebt keinen Bug).
- Nur automatisierte Verifikation, kein manuelles Geräteprotokoll im Plan
  (Nutzerentscheidung; Restrisiko dokumentiert unter „Open questions / Risks").
- Bug 3 wird minimal gelöst (Drawer während Submit nicht schließbar + gedimmter Inhalt),
  kein neues Loading-Overlay-Framework — Produkt-Konservatismus.

## Open questions / Risks

- **Restrisiko Bug 1:** `display-mode: standalone` + WebKit-Verhalten ist mit Playwright
  nicht nachstellbar. Die Automatisierung sichert Funktionsäquivalenz; dass der Flicker in
  der installierten PWA tatsächlich verschwunden ist, bestätigt erst der nächste Praxistest
  auf einem echten iPhone (shadcn-ui#8507 berichtet den Radix-Weg als funktionierend).
- **jsdom/Radix-Testverhalten:** Radix Dialog rendert per Portal und setzt
  `aria-hidden`/Fokus-Traps; bestehende Unit-Tests, die vaul-Verhalten implizit
  voraussetzen, können Anpassungen brauchen (Queries auf `role="dialog"` statt
  vaul-DOM-Struktur).
- **UX-Änderung:** Swipe-to-dismiss und Drag-Handle entfallen. Im Praxistest wurde beides
  nicht als Wert benannt (im Gegenteil: der Tap-als-Drag-Workaround existierte nur wegen
  vaul); falls Praxis-Feedback das vermisst, ist das ein separates Thema.

---

## Phase 1: Drawer-Engine auf Radix umstellen (Bug 1)

### Context

- `frontend/src/components/ui/drawer.tsx — DrawerContent()` — wird komplett neu geschrieben.
- `frontend/src/components/ui/sheet.tsx — SheetContent()` — Muster für Radix-Dialog-Wrapper
  mit Portal, Overlay und Slide-Animationen (`data-[side=bottom]`-Varianten).
- `frontend/src/service/components/TischAuswahlDrawer.tsx — TischAuswahlDrawer()` —
  `direction="bottom"`-Prop entfernen.
- `frontend/src/service/components/table/StickyActionBar.tsx — StickyActionBar()` —
  Trigger-Kompatibilität (`asChild`-Klonung von Ref + Click-Handler) prüfen.

### What to build

`drawer.tsx` neu auf `radix-ui`-`Dialog` implementieren: gleiche exportierte API, Bottom-
Sheet-Styling gemäß Layout-Vertrag (fixed unten, `max-h-[85dvh]`, `rounded-t-xl`,
Safe-Area-Padding, Slide-in/out von unten, Overlay mit Blur wie bisher). Handle-Div und
`data-vaul-no-drag`-Workaround entfallen ersatzlos. `vaul` aus `frontend/package.json` und
dem Lockfile entfernen. Alle sechs Verwender kompilieren unverändert (bis auf die entfernte
`direction`-Prop) und funktionieren: öffnen über Trigger bzw. `open`-Prop, schließen über
Abbrechen/Backdrop/Escape. Unit-Tests an Radix-Semantik anpassen (Dialog-Rolle, Portal).

Damit ist der fehlerhafte vaul-Scroll-Lock (`scrollTo(0,0)`) vollständig aus der App;
Hintergrund-Scroll-Sperre übernimmt Radix (`react-remove-scroll`), das laut shadcn-ui#8507
in iOS-Standalone korrekt arbeitet.

### Acceptance criteria

- [x] `grep -r "vaul" frontend/src frontend/package.json` liefert keine Treffer mehr.
- [x] Alle sechs Drawer öffnen und schließen in den bestehenden Unit-Tests über die
      unveränderte `Drawer*`-API (Dialog-Rolle sichtbar, Overlay vorhanden).
- [x] `StickyActionBar` als `DrawerTrigger asChild` öffnet den Drawer; bei leerer Auswahl
      verhindert der bestehende `onOpenChange`-Guard das Öffnen (bestehende Tests grün).
- [x] `make check` grün; komplette E2E-Suite (`desktop-admin` + `mobile-service`) grün —
      insbesondere `bestellen-kassieren.spec.ts`, `stornierung-serviceleitung.mobile.spec.ts`,
      `umbuchung.mobile.spec.ts`, `direktverkauf-storno.mobile.spec.ts`.

---

## Phase 2: Scroll-Layout für Bestellung und Zahlung (Bug 2, kritische Flows)

### Context

- `frontend/src/components/ui/drawer.tsx` — neues Export `DrawerBody` (einziger
  Scrollbereich, `min-h-0 overflow-y-auto`).
- `frontend/src/service/components/table/ZahlungDrawer.tsx — ZahlungDrawer()` — Auslöser (a)
  lange Liste und (b) dynamische Rückgeld-/Trinkgeld-Zeilen.
- `frontend/src/service/components/table/BestellungDrawer.tsx — BestellungDrawer()` —
  gleiche Struktur.
- `frontend/src/service/components/table/Receipt.tsx — Receipt()` — eigener
  `max-h-[40dvh]`-Cap entfällt.

### What to build

`DrawerBody` in `drawer.tsx` ergänzen. `BestellungDrawer` und `ZahlungDrawer` auf die
Struktur `DrawerHeader` → `DrawerBody` (Receipt + Beträge/Eingaben + Kommentar) →
`DrawerFooter` (Submit + Abbrechen) umbauen; der bisherige `mx-auto w-full max-w-sm`-
Block-Wrapper entfällt zugunsten von Breiten-Zentrierung innerhalb der drei Bereiche.
`Receipt` verliert seinen internen Scroll-Cap (scrollt jetzt als Teil von `DrawerBody`).
Ergebnis: Bei beliebig langer Positionsliste und nach Eingabe von Trinkgeld/Erhalten
(inkl. eingeblendeter Rückgeld-/Trinkgeld-Zeilen und Hinweistext) bleiben „Kassieren" bzw.
„Bestellung aufnehmen" und „Abbrechen" immer sichtbar; der Mittelteil scrollt.

Neue E2E-Regression im `mobile-service`-Projekt: Bestellung mit vielen Positionen (≥ 9)
aufnehmen, Kassieren-Drawer öffnen, Trinkgeld + Erhalten eingeben, dann per
`toBeInViewport()` prüfen, dass der Kassieren-Button sichtbar ist, und die Zahlung
erfolgreich abschließen.

### Acceptance criteria

- [x] Im mobilen E2E-Viewport ist der Kassieren-Button bei ≥ 9 Positionen **und** gefüllten
      Trinkgeld-/Erhalten-Feldern sichtbar (`toBeInViewport`) und die Zahlung wird
      erfolgreich verbucht (neue Spec grün).
- [x] Gleiches Kriterium für „Bestellung aufnehmen" bei einer Bestellung mit ≥ 9 Positionen.
- [x] Rückgeld-/Trinkgeld-Zeilen und Hinweistext erscheinen weiterhin korrekt (bestehende
      Unit-Tests `Zahlung.test.tsx` grün, ggf. angepasst an neue Struktur).
- [x] Kein Element im Drawer außer `DrawerBody` ist vertikal scrollbar (kein verschachteltes
      Scrollen; `Receipt` ohne eigenen `max-h`-Cap).
- [x] `make check` und E2E-Suite grün.

---

## Phase 3: Übrige Drawer auf das Scroll-Layout heben (Bug 2, Restflächen)

### Context

- `frontend/src/service/components/TischAuswahlDrawer.tsx — TischAuswahlDrawer()` —
  Suchfeld + potenziell lange Tisch-Liste (eigener `max-h-[60vh]`-Cap entfällt).
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx — HistorieStornierungDrawer()`
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx — HistorieUmbuchungDrawer()`
- `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx — DirektverkaufStornoDrawer()`
- `frontend/src/service/components/PositionAuswahlListe.tsx — PositionAuswahlListe()` —
  Liste wandert in `DrawerBody`.

### What to build

Die vier verbleibenden Drawer auf dieselbe Struktur umstellen: `DrawerHeader` →
`DrawerBody` (Suchfeld/Listen/Kommentar) → `DrawerFooter` (sofern vorhanden). Beim
`TischAuswahlDrawer` bleibt das Suchfeld oberhalb des Scrollbereichs sichtbar (Teil des
Headers oder als nicht scrollendes Element vor `DrawerBody`); der `max-h-[60vh]`-Cap der
Tisch-Liste entfällt. Bei Storno-/Umbuchungs-Drawern bleiben „Stornierung erteilen" /
Umbuchen-Button und „Abbrechen" bei beliebig vielen Positionen sichtbar.

### Acceptance criteria

- [x] Alle vier Drawer folgen der Struktur Header → `DrawerBody` → Footer; keine eigenen
      `max-h`-Caps in Listen mehr (`grep -rn "max-h-\[" frontend/src/service` zeigt keine
      Treffer in Drawer-Inhalten).
- [x] Storno- und Umbuchungs-Footer sind im mobilen E2E-Viewport bei vielen Positionen
      sichtbar und funktionsfähig (bestehende Specs `stornierung-serviceleitung.mobile`,
      `umbuchung.mobile`, `direktverkauf-storno.mobile` grün; wo keine lange Liste
      abgedeckt ist, Unit-Test auf Strukturebene ergänzen).
- [x] Tischsuche im `TischAuswahlDrawer` funktioniert unverändert (bestehender Unit-Test
      grün); das Suchfeld scrollt nicht mit der Liste.
- [x] `make check` und E2E-Suite grün.

---

## Phase 4: Pending-Zustand beim Absenden (Bug 3)

### Context

- `frontend/src/hooks/use-action-submit.ts — useActionSubmit()` — `loading`-Quelle.
- `frontend/src/components/ui/drawer.tsx — DrawerContent()` — Schließ-Verhalten (Backdrop,
  Escape) muss während Submit unterbindbar sein.
- `frontend/src/service/components/table/BestellungDrawer.tsx`, `ZahlungDrawer.tsx`,
  `HistorieStornierungDrawer.tsx`, `HistorieUmbuchungDrawer.tsx`,
  `DirektverkaufStornoDrawer.tsx` — alle Submit-Drawer.

### What to build

Während `loading === true`: Drawer nicht schließbar (Backdrop-Tap, Escape und
`onOpenChange(false)` werden ignoriert; die Buttons sind bereits `disabled`) und der
Drawer-Inhalt wird sichtbar als „in Arbeit" markiert (Dimmen/`pointer-events-none` des
Bodys; der Button-Spinner bleibt die primäre Anzeige). Umsetzung als kleine, einheitliche
Erweiterung — z. B. eine `pending`-Prop auf `DrawerContent` — statt Copy-Paste-Logik in
jedem Drawer. Kein neues Overlay-Framework, kein globaler State.

### Acceptance criteria

- [x] Während eines laufenden Submits schließt weder Escape noch Backdrop-Tap den Drawer;
      nach Erfolg schließt er wie bisher automatisch (Unit-Test je Verhalten).
- [x] Der Pending-Zustand ist visuell erkennbar (gedimmter Body + Spinner im Button) —
      Unit-Test auf Strukturebene.
- [x] Fehlerfall unverändert: Toast erscheint, Drawer bleibt offen und wieder bedienbar
      (bestehende Tests grün).
- [x] `make check` und E2E-Suite grün.
