# Plan: Zweispaltiges Service-Layout für Tablet und größer

> Source PRD: `docs/prds/prd-service-split-screen-tablet.md`

## Goal

Der Service-Bereich (`/service/*`) wird ab Viewport-Breite `lg` (1024px)
responsiv zweispaltig: links die Auswahl, rechts eine dauerhaft sichtbare
Abschluss-Spalte, die exakt den heutigen Drawer-Inhalt trägt (Belegvorschau,
Erhalten/Rückgeld/Kommentar, primärer Aktionsbutton). Unterhalb `lg` bleibt
alles unverändert (einspaltig, fixierte Dock-Leiste, Bottom-Sheet-Drawer). Die
zeilengebundenen Korrekturvorgänge (Stornieren, Umbuchen) bleiben Overlays,
erscheinen ab `lg` aber als mittig zentrierter Dialog statt als Bottom-Sheet.

Reiner Präsentations-Layer: kein Backend, keine Endpunkte, keine Event-Formate,
keine Validierung, Idempotenz oder Steuer-/TSE-Logik ändern sich. Damit
freeze-konform (nur Frontend).

## Architectural decisions

Durable decisions that apply across all phases:

- **Kein Backend/Schema/Event/API-Vertrag ändert sich.** Ausschließlich
  `frontend/src/service/**` ist betroffen. Keine neue Migration, kein neues
  Event, keine geänderte Schema-Nutzlast.
- **Breakpoint `lg` (1024px), gesteuert über den bestehenden `useIsMobile`-Hook**
  (`frontend/src/hooks/use-mobile.ts`, `MOBILE_BREAKPOINT = 1024`), konsistent mit
  ADR 07 (`docs/adrs/07_desktop-breakpoint-lg.md`). Für die drei
  Zusammenstellungs-Flächen entscheidet `useIsMobile()` (JS-Breakpoint), welcher
  Container gerendert wird — nicht reine CSS-Sichtbarkeit —, damit der
  Abschluss-Inhalt **genau einmal** mountet (eine State-Quelle, ein
  Idempotenz-Schlüssel).
- **Eine Quelle je Abschluss-Inhalt.** Der heutige Drawer-Body samt Footer jeder
  der drei Flächen wird als darstellungsneutrale Abschluss-Inhaltskomponente
  extrahiert (Belegvorschau, Eingaben, Betragsberechnung, primärer Button,
  Retry-/Fehlerverhalten). Diese eine Komponente wird sowohl im
  Bottom-Sheet-Drawer (schmal) als auch in der festen Abschluss-Spalte (breit)
  gerendert; die beiden Darstellungen unterscheiden sich nur im umschließenden
  Container. Es entsteht kein zweites Layout- oder Drawer-System.
- **Abschluss-Spalte nur für die drei Zusammenstellungs-Flächen:**
  Direktverkauf-Verkaufen, Tisch-Bestellen, Tisch-Kassieren. Historie,
  Tischauswahl und alle reinen Listen behalten ihre vorhandenen
  Rasterlayouts über volle Breite und bekommen keine Abschluss-Spalte.
- **Ein Drawer-System mit responsiver Präsentation.** Das eine Drawer-Primitive
  (`frontend/src/components/ui/drawer.tsx`, ADR 03) erhält für die
  Korrekturvorgänge eine responsive Präsentation: Bottom-Sheet unterhalb `lg`,
  mittig zentrierter Modal-Dialog ab `lg`. Der bindende Bottom-Sheet-Vertrag aus
  ADR 03 (85dvh, Safe-Area, ein Scrollbereich, kein Drag-Handle) gilt unverändert
  für Handy-Breiten. Kein Fork der Primitive.
- **Idempotenz-Schlüssel-Lebenszyklus (nur Direktverkauf + Bestellen).** Der
  Schlüssel (`verkaufId` bzw. `bestellungId`) wird in der festen Spalte je
  logischem Vorgang erzeugt (neuer Schlüssel, wenn eine Zusammenstellung aus dem
  Leerzustand beginnt; erneut nach jedem erfolgreichen Abschluss beim
  Zurücksetzen). Ein Retry desselben Vorgangs behält seinen Schlüssel.
- **Tisch-Kassieren trägt keinen Client-Idempotenz-Schlüssel** (siehe Resolved
  decisions): `ZahlungKassierenSchema` hat keinen Schlüssel; die Idempotenz ist
  zustandsbasiert (bereits bezahlte Positionen → `position_nicht_bezahlbar`).
  Es wird **kein** neuer Schlüssel eingeführt (wäre eine Backend-Änderung,
  out of scope). Der `useActionSubmit`-Loading-Guard verhindert den Doppel-Submit
  in der festen Spalte genauso wie heute im Drawer.
- **Neues ADR 08.** Die Entscheidung „Service-Bereich ist responsiv zweispaltig
  ab `lg`, mit responsiver Drawer-Präsentation für Korrekturvorgänge" wird als
  ADR `docs/adrs/08_service-split-screen.md` festgehalten (Nygard-Format), analog
  zu ADR 03, und in `docs/adrs/README.md` eingetragen.

## Inventory

Bestehende Dateien, Muster und Abhängigkeiten:

- `frontend/src/service/ServiceLayout.tsx` — `ServiceLayout()` — gemeinsame
  Service-Shell (Header + `<Outlet />`); trägt heute kein Layout-Splitting.
- `frontend/src/service/components/ServiceDock.tsx` — `ServiceDock()`,
  `DockActionSlot()`, `DockSlotContext` — fixierte Bodenfläche (Aktions-Slot per
  Portal + TabsList). Reine Handy-Darstellung ab diesem PRD.
- `frontend/src/service/components/table/DockActionButton.tsx` —
  `DockActionButton()` — primärer Aktionsbutton, rendert per `DockActionSlot` ins
  Dock. In der festen Spalte trägt die Spalte den Button selbst.
- `frontend/src/service/DirektverkaufPage.tsx` — `DirektverkaufPage()` — Tab-Host
  „Verkaufen"/„Historie", hostet `ServiceDock` und `ErfolgsPop`.
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx` —
  `Direktverkauf()` — Fläche 1 (Auswahl `ProductList` + `DirektverkaufDrawer`).
- `frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx` —
  `DirektverkaufDrawer()` — heutiger Abschluss-Inhalt Fläche 1 (Receipt,
  Erhalten/Rückgeld, Kommentar, „Verkauf abschließen", `verkaufId`-Lebenszyklus).
- `frontend/src/service/TablePage.tsx` — `TablePage()` — Tab-Host
  „Bestellen"/„Kassieren"/„Historie", hostet `ServiceDock` und `ErfolgsPop`,
  `dockFreiraum`-Idiom für den unteren Freiraum.
- `frontend/src/service/components/table/Bestellung.tsx` — `Bestellung()` —
  Fläche 2 (Auswahl + `BestellungDrawer`).
- `frontend/src/service/components/table/BestellungDrawer.tsx` —
  `BestellungDrawer()` — heutiger Abschluss-Inhalt Fläche 2
  (Receipt, Kommentar, Gesamt, „Bestellung aufnehmen", `bestellungId`).
- `frontend/src/service/components/table/Zahlung.tsx` — `Zahlung()` — Fläche 3
  (Auswahl offener Positionen + `ZahlungDrawer`).
- `frontend/src/service/components/table/ZahlungDrawer.tsx` — `ZahlungDrawer()` —
  heutiger Abschluss-Inhalt Fläche 3 (Receipt, Erhalten/Zielbetrag/Rückgeld/
  Trinkgeld, Kommentar, „Kassieren"). Nutzt `DockActionSlot` zusätzlich für die
  Restbetrag-Zeile.
- `frontend/src/service/table/Zahlung.ts` — `ZahlungKassierenSchema` — Nutzlast
  ohne Idempotenz-Schlüssel (Beleg für „kein `zahlungId`").
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`,
  `HistorieUmbuchungDrawer.tsx`,
  `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx` —
  die drei Korrekturvorgänge, alle auf dem Drawer-Primitive.
- `frontend/src/components/ui/drawer.tsx` — Drawer-Primitive auf Radix Dialog;
  `DrawerContent`-Klassen tragen den ADR-03-Layout-Vertrag.
- `frontend/src/hooks/use-mobile.ts` — `useIsMobile()`, `MOBILE_BREAKPOINT = 1024`.
- `frontend/src/hooks/use-action-submit.ts` — `useActionSubmit()` — Submit-Guard
  (`loading`), Fehler-Mapping. Verhalten bleibt in allen Flächen identisch.
- Prior-Art-Tests: `DirektverkaufDrawer`-, `ZahlungDrawer`-, `Bestellung`-,
  `Zahlung`-Tests (Vitest + Testing Library) prüfen Texte, Beträge, Callback-
  Reihenfolge und Backend-Nutzlasten — nicht Container/CSS.

## Resolved decisions

- **Kein `zahlungId` für Tisch-Kassieren.** Die PRD-Formulierung „`verkaufId` bzw.
  `zahlungId` … beim Öffnen des Drawers neu erzeugt" trifft für Kassieren nicht
  zu: `ZahlungKassierenSchema` (`frontend/src/service/table/Zahlung.ts`) trägt
  keinen Schlüssel, `ZahlungDrawer` erzeugt keinen. Idempotenz ist dort
  zustandsbasiert. Ein neuer Schlüssel wäre eine Backend-Änderung und damit
  out of scope. Der Idempotenz-Lebenszyklus dieses Plans gilt nur für
  Direktverkauf (`verkaufId`) und Bestellen (`bestellungId`).
- **Container-Wahl per `useIsMobile()`, nicht per CSS-Sichtbarkeit,** damit der
  extrahierte Abschluss-Inhalt genau einmal mountet und nicht in zwei parallel
  gemounteten Bäumen (Drawer + Spalte) doppelten State und doppelte
  Idempotenz-Schlüssel führt.
- **Responsive Drawer-Präsentation im Primitive, kein Fork** (Phase 4): Die
  Bottom-Sheet-versus-zentrierter-Dialog-Umschaltung wird als responsive Variante
  im `DrawerContent` umgesetzt; der ADR-03-Handy-Vertrag bleibt unter `lg`
  wörtlich erhalten.
- **ADR 08 wird bereits in Phase 1 geschrieben** und beschreibt das vollständige
  Zielbild (Abschluss-Spalte für die drei Flächen + Dialog-Präsentation der
  Korrekturvorgänge), auch wenn die Umsetzung über die Phasen verteilt landet —
  ADRs werden nie umgeschrieben (`docs/adrs/README.md`).

## Open questions / Risks

- **Sicherheitskritische Fläche wird umgebaut.** Tisch-Kassieren und
  Direktverkauf sind die kassen- und geldrelevanten Oberflächen. Die
  darstellungsneutrale Extraktion muss verhaltensgleich zum heutigen Drawer sein
  (Referenz-Verhalten laut PRD). Absicherung: bestehende Verhaltenstests ziehen
  unverändert auf die extrahierten Komponenten um.
- **Zweispaltige Darstellung und Breakpoint-Verhalten sind manuelle Abnahme**
  (schmal, Querformat-Tablet, Laptop; Hell/Dunkel), analog zur Spektral-Abnahme.
  Automatisierte Tests prüfen Verhalten, nicht welcher Container rendert.

---

## Phase 1: Responsive Shell + Direktverkauf-Verkaufen zweispaltig (Tracer Bullet)

**User stories**: 1, 4, 6

### Context

- `frontend/src/service/ServiceLayout.tsx — ServiceLayout()` — Ort für die
  gemeinsame responsive Shell-Entscheidung, falls die Zweispaltigkeit dort
  zentral getragen wird; alternativ pro Tab-Fläche.
- `frontend/src/service/DirektverkaufPage.tsx — DirektverkaufPage()` — Tab-Host,
  entscheidet je Breakpoint über Dock+Drawer (schmal) versus Zweispalt (breit).
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx — Direktverkauf()`
  — Auswahl links; bindet heute den Drawer ein.
- `frontend/src/service/components/direktverkauf/DirektverkaufDrawer.tsx —
  DirektverkaufDrawer()` — Quelle des zu extrahierenden Abschluss-Inhalts.
- `frontend/src/service/components/ServiceDock.tsx — ServiceDock()`,
  `frontend/src/service/components/table/DockActionButton.tsx — DockActionButton()`
  — bleiben die Handy-Darstellung; die feste Spalte trägt den Button selbst.
- `frontend/src/hooks/use-mobile.ts — useIsMobile()` — Breakpoint-Quelle.

### What to build

Die tragende Scheibe: die responsive Service-Shell **und** die erste Fläche
(Direktverkauf-Verkaufen) end-to-end.

1. Den Abschluss-Inhalt von `DirektverkaufDrawer` als eine
   darstellungsneutrale Inhaltskomponente herausziehen (Receipt, Erhalten/
   Rückgeld, Kommentar, primärer Button, `verkaufId`-Lebenszyklus, Submit-/
   Fehler-/Retry-Verhalten). Diese Komponente kennt keinen Container.
2. Die Fläche rendert breakpoint-abhängig: unter `lg` weiterhin
   `ServiceDock` + Bottom-Sheet-Drawer mit dem extrahierten Inhalt; ab `lg` ein
   zweispaltiges Layout (Auswahl links, feste Abschluss-Spalte rechts mit
   demselben extrahierten Inhalt und dem primären Button in der Spalte). Beide
   Spalten scrollen unabhängig; die Reiter liegen ab `lg` oben im Inhaltsbereich,
   die fixierte Dock-Leiste entfällt auf dieser Breite.
3. Idempotenz-Schlüssel-Lebenszyklus in der festen Spalte: `verkaufId` je
   logischem Vorgang (neu aus dem Leerzustand heraus, neu nach erfolgreichem
   Abschluss beim Reset), stabil über Retry.
4. Leerzustand der Abschluss-Spalte: Hinweistext (etwa „Produkte auswählen") und
   deaktivierter Aktionsbutton, gleichbedeutend mit dem heute deaktivierten
   Dock-Button.
5. Erfolgs-Pop, Fehler-Toasts und Betragsberechnung bleiben identisch; der
   nachgelagerte Refetch läuft weiter beim Schließen des Pops.
6. ADR `docs/adrs/08_service-split-screen.md` schreiben (vollständiges Zielbild)
   und in `docs/adrs/README.md` eintragen.

### Acceptance criteria

- [x] Ab 1024px zeigt Direktverkauf-Verkaufen Produkte links und den Warenkorb
      (Receipt, Erhalten/Rückgeld, Kommentar, „Verkauf abschließen") rechts
      dauerhaft nebeneinander; kein Bottom-Sheet-Drawer und keine fixierte
      Dock-Leiste erscheinen auf dieser Breite. _(Sichtabnahme manuell)_
- [x] Unterhalb 1024px ist Direktverkauf-Verkaufen unverändert (einspaltig,
      fixierte Dock-Leiste, Bottom-Sheet-Drawer).
- [x] Der Abschluss-Inhalt stammt aus **einer** Komponente, die in beiden
      Containern gerendert wird (kein dupliziertes Markup, keine zweite
      State-Quelle).
- [x] Leerzustand der Abschluss-Spalte: Hinweistext plus deaktivierter Button.
- [x] `verkaufId` bleibt über einen Retry stabil und wechselt bei einem neuen
      logischen Vorgang (aus Leerzustand bzw. nach erfolgreichem Abschluss).
- [x] Beim Abschluss wird `direktverkaufTaetigen` genau einmal mit der
      erwarteten Nutzlast aufgerufen; Erfolgs-Pop und Betragsberechnung
      verhalten sich wie zuvor.
- [x] Die bestehenden Direktverkauf-Verhaltenstests laufen (ggf. auf die
      extrahierte Komponente umgezogen) grün; keine Snapshot-Tests.
- [x] `docs/adrs/08_service-split-screen.md` existiert, ist in
      `docs/adrs/README.md` verlinkt, und `make check` ist grün.

---

## Phase 2: Tisch-Bestellen zweispaltig

**User stories**: 2, 6

### Context

- `frontend/src/service/TablePage.tsx — TablePage()` — Tab-Host; Reiter
  „Bestellen"/„Kassieren"/„Historie".
- `frontend/src/service/components/table/Bestellung.tsx — Bestellung()` — Auswahl
  links.
- `frontend/src/service/components/table/BestellungDrawer.tsx — BestellungDrawer()`
  — Quelle des zu extrahierenden Abschluss-Inhalts (`bestellungId`).
- Shell-Muster und extrahierter-Inhalt-Ansatz aus Phase 1.

### What to build

Tisch-Bestellen auf dasselbe responsive Muster heben: Abschluss-Inhalt von
`BestellungDrawer` darstellungsneutral extrahieren (Receipt, Kommentar, Gesamt,
„Bestellung aufnehmen", `bestellungId`-Lebenszyklus). Unter `lg` weiterhin
Dock+Drawer; ab `lg` Produktliste links, Bestellübersicht rechts mit dem Button
in der Spalte. Leerzustand und Idempotenz-Lebenszyklus analog Phase 1.

### Acceptance criteria

- [ ] Ab 1024px zeigt der Reiter „Bestellen" die Produktliste links und die
      entstehende Bestellung rechts dauerhaft nebeneinander; kein Drawer, keine
      fixierte Dock-Leiste auf dieser Breite.
- [ ] Unterhalb 1024px ist der Reiter „Bestellen" unverändert.
- [ ] Abschluss-Inhalt aus einer Komponente in beiden Containern; Leerzustand mit
      Hinweistext plus deaktiviertem Button.
- [ ] `bestellungId` bleibt über Retry stabil und wechselt bei neuem logischem
      Vorgang.
- [ ] Beim Abschluss wird `bestellungAufnehmen` genau einmal mit der erwarteten
      Nutzlast aufgerufen; Erfolgs-Pop verhält sich wie zuvor.
- [ ] Bestehende Bestellen-Verhaltenstests grün; `make check` grün.

---

## Phase 3: Tisch-Kassieren zweispaltig

**User stories**: 3, 6

### Context

- `frontend/src/service/TablePage.tsx — TablePage()` — Tab-Host.
- `frontend/src/service/components/table/Zahlung.tsx — Zahlung()` — Auswahl
  offener Positionen links (mit „Alle auswählen", Fremdpositionen-Aufklapper).
- `frontend/src/service/components/table/ZahlungDrawer.tsx — ZahlungDrawer()` —
  Quelle des Abschluss-Inhalts; nutzt zusätzlich `DockActionSlot` für die
  Restbetrag-Zeile.
- `frontend/src/service/table/Zahlung.ts — ZahlungKassierenSchema` — Beleg für
  „kein Client-Idempotenz-Schlüssel".

### What to build

Tisch-Kassieren auf das responsive Muster heben: Abschluss-Inhalt von
`ZahlungDrawer` darstellungsneutral extrahieren (Receipt, Erhalten, Zielbetrag
inkl. Trinkgeld, Rückgeld, Trinkgeld-Hinweis, Kommentar, „Kassieren"). Unter `lg`
weiterhin Dock+Drawer inklusive der Restbetrag-Zeile im Dock-Slot; ab `lg` offene
Positionen links, Zahlungsübersicht rechts mit dem Button in der Spalte. Die
Restbetrag-Zeile (`restNachZahlungCents`) erscheint ab `lg` in der Spalte statt im
Dock-Slot. **Kein** neuer Idempotenz-Schlüssel — die feste Spalte setzt bei
erfolgreicher Zahlung nur den lokalen Eingabe-State zurück; der
`useActionSubmit`-Loading-Guard verhindert den Doppel-Submit wie heute.

### Acceptance criteria

- [ ] Ab 1024px zeigt der Reiter „Kassieren" die offenen Positionen links und die
      Zahlungsübersicht (Erhalten, Zielbetrag/Trinkgeld, Rückgeld, Restbetrag)
      rechts dauerhaft nebeneinander; kein Drawer, keine fixierte Dock-Leiste.
- [ ] Die Restbetrag-Zeile ist ab `lg` in der Abschluss-Spalte sichtbar (nicht im
      Dock-Slot).
- [ ] Unterhalb 1024px ist der Reiter „Kassieren" unverändert, inklusive der
      Restbetrag-Zeile im Dock-Slot.
- [ ] Abschluss-Inhalt aus einer Komponente in beiden Containern; Leerzustand mit
      Hinweistext plus deaktiviertem Button.
- [ ] Rückgeld- und Trinkgeld-Berechnung sind in der breiten Ansicht identisch zur
      schmalen.
- [ ] Beim Abschluss wird `zahlungKassieren` genau einmal mit der erwarteten
      Nutzlast (ohne zusätzlichen Schlüssel) aufgerufen; ein Doppelklick löst
      keinen zweiten Aufruf aus.
- [ ] Bestehende Kassieren-Verhaltenstests grün; `make check` grün.

---

## Phase 4: Korrekturvorgänge als zentrierter Dialog ab `lg`

**User stories**: 5

### Context

- `frontend/src/components/ui/drawer.tsx — DrawerContent()` — trägt den
  ADR-03-Layout-Vertrag; erhält die responsive Präsentation.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`,
  `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx`,
  `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx` —
  die drei Korrekturvorgänge auf dem Drawer-Primitive.
- `docs/adrs/03_drawer-radix-statt-vaul.md` — bindender Handy-Vertrag.

### What to build

Dem Drawer-Primitive eine responsive Präsentation geben: unter `lg` unverändert
Bottom-Sheet (85dvh, Safe-Area, ein Scrollbereich, kein Drag-Handle — ADR 03),
ab `lg` mittig zentrierter Modal-Dialog. Umsetzung als responsive Variante im
`DrawerContent` (kein Fork, kein zweites System). Betroffen sind die drei
Korrekturvorgänge Storno (Tisch), Umbuchung (Tisch) und Storno (Direktverkauf).
Fachliches Verhalten (Validierung, Idempotenz, Fehler) bleibt unberührt.

### Acceptance criteria

- [ ] Ab 1024px erscheinen Stornieren und Umbuchen (Tisch) sowie der
      Direktverkauf-Storno als mittig zentrierter Dialog; unterhalb 1024px
      unverändert als Bottom-Sheet.
- [ ] Der ADR-03-Handy-Vertrag (85dvh, Safe-Area, ein Scrollbereich, kein
      Drag-Handle) gilt unter `lg` wörtlich weiter; die installierte iOS-PWA
      (nur Handy-Breiten) ist nicht betroffen.
- [ ] Es bleibt ein Drawer-System (kein Fork der Primitive, kein Parallelsystem).
- [ ] Auf großen Bildschirmen fährt in keiner Service-Fläche mehr ein Sheet vom
      unteren Rand herein.
- [ ] Fachliches Verhalten der Korrekturvorgänge unverändert; bestehende Tests
      grün; `make check` grün.
