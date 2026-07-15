# Plan: Behebung der Multi-Experten-Review-Befunde

> Source PRD: n/a (aus Review der Commits 68d7916..713b247, 2026-07-14)

## Goal

Alle bestätigten Befunde des Multi-Experten-Reviews sowie die dabei
verifizierten, bereits vor dem Review-Bereich bestehenden Layout-Defekte
beheben. Der Review ergab: das Spektral-Redesign und die Refactor-Welle sind
im Kern gesund (Backend, Event-Contracts und DSFinV-K-Format unangetastet,
Frontend-Suite grün) — die offenen Punkte sind fünf Low-Findings im Frontend,
zwei Vertragswerk-Findings (ein Medium, ein Low) und vier pre-existing
Tablet-/Phone-Layout-Defekte im Admin-Bereich. Alle Änderungen sind
Korrektheits-, Robustheits- und Konsistenzfixes; kein neues Feature, keine
Änderung an persistierten Daten, kein Eingriff in die Freeze-Disziplin.

## Architectural decisions

Durable decisions, die für alle Phasen gelten:

- **Erfolgs-Pop bleibt.** Das Vollbild-Overlay wird nicht durch Toasts ersetzt;
  es wird gehärtet (Alt-Browser-Fallback, Screenreader). Der bewusste
  Redesign-Entwurf bleibt erhalten. **Wichtig:** Das Overlay bleibt bewusst
  klick-blockierend (kein `pointer-events-none`) — der nachgelagerte Refetch
  (`onDismiss` → `reload` in `TablePage`) ist an das Schließen gekoppelt, sodass
  das Overlay die nächste Interaktion so lange hält, bis die frischen Daten
  geladen sind. Ein Tap schließt weiterhin früher (`onClick={onDismiss}`). Der
  ursprünglich geplante Tap-Durchgriff (`pointer-events-none`) würde erlauben,
  auf noch nicht aktualisierten Bildschirmzustand zu tippen (Bestellen →
  Kassieren rennt der Projektion davon) und wird deshalb nicht umgesetzt.
- **Alt-Browser-Fallback-Muster.** Neue dekorative `color-mix()`-Flächen
  bekommen eine solide bzw. rgba-Vorab-Deklaration, sodass bei fehlendem
  `color-mix()`-Support (iOS Safari 16.0–16.1) ein sichtbarer Ersatz bleibt.
  Backdrop-Blur wird — wie bei den bestehenden Overlays (`drawer`, `dialog`,
  `alert-dialog`, `sheet`) — hinter `supports-[backdrop-filter]` gegatet.
- **Glow-Clipping isoliert.** Dekorative Glows werden in einer eigenen,
  absolut positionierten `aria-hidden`-Ebene (`inset-0 overflow-hidden -z-10`)
  geclippt; der Inhalts-/Scroll-Container behält normalen Überlauf, damit hohe
  Karten scrollbar bleiben.
- **Admin-Chrome ab `md`.** Im Band 768–1023px besitzt die feste Desktop-Sidebar
  das Chrome; der mobile Hamburger-Header wird ab `md` ausgeblendet
  (`md:hidden` statt `lg:hidden`), passend zum Sichtbarkeitsbreakpoint der
  Sidebar (`md:block`/`md:flex`). Folge: die effektive Inhaltsbreite bei 768px
  beträgt ~512px — die Overflow-Fixes in Phase 4 müssen auf diese Breite zielen.
- **Geld-Formatierung mit geschütztem Leerzeichen.** Betrag und `€` werden nie
  durch ein normales Leerzeichen getrennt (Umbruchgefahr in Kacheln/Tabellen).
  Neuer Helper `formatEuro(cents)` in `frontend/src/lib/utils.ts` liefert
  `` `${formatCents(cents)} €` `` (NBSP); die Betrag-plus-`€`-Anzeigestellen
  konsumieren ihn. Der bereits bestehende ` €`-Gebrauch in
  `formatAlleAuswaehlenLabel` wird auf denselben Helper umgestellt.
- **Vertrags-Fassungsbezug.** Wo die Website die Annahme erklärt, referenziert
  sie wortgleich die geltende Fassung „14. Juli 2026" und die TERMS-URL
  (`https://github.com/nicograef/jotti/blob/main/TERMS.md`), konsistent mit der
  E-Mail-Vorlage in `TERMS.md`.

## Inventory

Bestätigte In-Range-Findings (Frontend):

- `frontend/src/service/components/ErfolgsPop.tsx — ErfolgsPop()` — Vollbild-
  Overlay: `onClick={onDismiss}` ohne `pointer-events-none` (verschluckt einen
  Tap), `bg-[color-mix(...)] backdrop-blur-[6px]` ohne Fallback/Guard, und
  `role="status"`-Live-Region wird erst beim Öffnen befüllt gemountet.
- `frontend/src/components/ui/skeleton.tsx — Skeleton()` +
  `frontend/src/index.css — @utility skeleton-shimmer` — einziger `background`
  ist ein Gradient mit `color-mix()`-Stops ohne soliden Vorab-Fallback.
- `frontend/src/components/common/AuthLayout.tsx — AuthLayout()` —
  `min-h-screen max-h-screen … overflow-hidden` am Wrapper, der auch die Karte
  hält; clippt hohe Karten (Passwort-Setzen) auf kurzen/Landscape-Viewports.

Bestätigte In-Range-Findings (Vertragswerk):

- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` — Betreff
  „Nutzungsvereinbarung anfragen", Body ist ein Feld-Abzug ohne
  Annahmeerklärung und ohne Fassungsbezug.
- `website/src/pages/fuer-vereine.astro — schritte` — beschreibt den
  abgeschafften Drei-Schritte-Prozess (Anfrage → Vereinbarung in Textform →
  Loslegen) mit Autor-Bestätigung.
- `LICENSE — Section 9 (INDEMNIFICATION)` — carve-out-freie „indemnify, defend,
  and hold harmless … from any and all claims"; `Section 12 (SEVERABILITY)` —
  geltungserhaltende Reduktion („modified to the minimum extent necessary").
- `TERMS.md — § 6(3)` + `TERMS.md — Prozess / E-Mail-Vorlage` — Referenz für die
  schlanke Freistellung (mit Vorsatz/grobe-Fahrlässigkeit-Carve-out) und den
  Ein-E-Mail-Wortlaut.

Pre-existing Layout-Defekte (per A/B gegen 68d7916^ als bestandsalt verifiziert):

- `frontend/src/admin/AdminLayout.tsx — AdminMobileHeader()` — `lg:hidden`,
  während die Sidebar (`frontend/src/components/ui/sidebar.tsx`) ab `md:block`
  erscheint → doppeltes Top-Chrome im 768–1023px-Band. Mobile-Breakpoint-Quelle:
  `frontend/src/hooks/use-mobile.ts — MOBILE_BREAKPOINT = 768`.
- `frontend/src/admin/reporting/KassenberichtePage.tsx` — DSFinV-K-Hinweiszeile:
  `flex items-center gap-4` ohne Umbruch, mit nicht-schrumpfendem Button „Archiv
  herunterladen (ZIP)" → horizontaler Überlauf auf Phone- und 768px-Breite.
- `frontend/src/admin/kasse/LaufenderBetriebSection.tsx — BewegungZeile()` —
  `truncate`-Span in `flex … justify-between` verursacht Überlauf bei 768px.
- Geldanzeige-Stellen mit `{formatCents(x)} €` (normales Leerzeichen), u. a.
  `frontend/src/admin/reporting/LiveReportingSection.tsx`,
  `ReportingResults.tsx`, `SummaryCard.tsx`, `SitzungsListe.tsx`,
  `StornoItem.tsx` — `€` kann in schmale Kacheln/Steuertabelle umbrechen.

Bestehende Muster zum Nachnutzen:

- `frontend/src/lib/utils.ts — formatCents()`, `formatAlleAuswaehlenLabel()`
  (nutzt bereits ` €`).
- Overlay-Fallback-Muster: `frontend/src/components/ui/dialog.tsx`,
  `drawer.tsx`, `alert-dialog.tsx`, `sheet.tsx`
  (`bg-black/10 … supports-[backdrop-filter]:backdrop-blur-xs`).

## Resolved decisions

- **Erfolgs-Pop:** bleibt erhalten, wird nur gehärtet (nicht durch Toast
  ersetzt).
- **Admin-Tablet-Chrome:** Desktop-Sidebar besitzt 768px+ (Hamburger-Header
  `md:hidden`); die feste Sidebar gilt bewusst ab `md`.
- Alle pre-existing Layout-Defekte werden mitbehoben (ausdrücklich gewünscht).

## Open questions / Risks

- Die Sidebar belegt bei 768px ~256px Breite; nach der Chrome-Angleichung
  müssen die Overflow-Fixes (Phase 4) für ~512px effektive Inhaltsbreite
  greifen — Verifikation bei 768px mit sichtbarer Sidebar, nicht nur am nackten
  Viewport.
- Vertragswerk-Änderungen sind eine Umsetzung der Review-Einschätzung, keine
  Rechtsberatung; der finale Wortlaut sollte vom Autor gegengelesen werden.

---

## Phase 1: Erfolgs-Pop härten

### Context

- `frontend/src/service/components/ErfolgsPop.tsx — ErfolgsPop()` — die drei
  Befunde (Tap-Verschlucken, fehlender Backdrop-Fallback, unzuverlässige
  Live-Region) sitzen alle in dieser Komponente.
- `frontend/src/components/ui/dialog.tsx` u. a. — kanonisches
  `supports-[backdrop-filter]`-Backdrop-Muster als Vorlage.
- `frontend/src/service/components/ErfolgsPop.test.tsx` (falls vorhanden) bzw.
  die Testdatei neben der Komponente — deckt Auto-Dismiss und Rendern ab.

### What to build

Den Erfolgs-Pop so anpassen, dass (1) das Overlay klick-blockierend bleibt und
ein Tap es weiterhin früher schließt (`onClick={onDismiss}`): der ursprünglich
geplante `pointer-events-none`-Durchgriff wird **nicht** umgesetzt, weil der
nachgelagerte Refetch an `onDismiss` hängt — ein durchgereichter Tap würde auf
noch nicht aktualisiertem Zustand landen (Bestellen → Kassieren, die Projektion
ist noch nicht nachgeladen); (2) die Tönung auf Browsern ohne `color-mix()`
sichtbar bleibt: solider/rgba-Vorab-Hintergrund vor der `color-mix()`-Deklaration
und Blur hinter `supports-[backdrop-filter]` gegatet, analog zu den anderen
Overlays; (3) die Erfolgsmeldung zuverlässig angesagt wird: die
`role="status"`-Live-Region wird dauerhaft gemountet und nur ihr Inhalt
getoggelt (statt die befüllte Region frisch zu mounten).

### Acceptance criteria

- [x] Das Overlay bleibt klick-blockierend, sodass der an `onDismiss` gekoppelte
      Refetch vor der nächsten Interaktion greift (verifiziert im
      Bestellen→Kassieren-Flow: der Tab-Klick wartet ~1,7 s auf Pop-Schließen +
      Nachladen, danach ist der Kassieren-Button aktiv); ein Tap schließt den Pop
      weiterhin früher (`onClick={onDismiss}`).
- [x] Auf simuliertem fehlendem `color-mix()`-Support bleibt eine sichtbare
      Abdunklung; mit `supports`-Guard bleibt der Blur nur dort aktiv, wo
      `backdrop-filter` unterstützt wird (Muster identisch zu `dialog.tsx`).
- [x] Die Live-Region ist im gemounteten Service-Screen dauerhaft vorhanden und
      wechselt nur ihren Textinhalt; der Erfolgstext wird bei Bestellen,
      Kassieren und Direktverkauf angesagt.
- [x] Auto-Dismiss nach `ANZEIGE_DAUER_MS` funktioniert unverändert; Timer wird
      bei Unmount aufgeräumt; bestehende Tests grün, inkl. Test, der belegt, dass
      ein Tap den Pop schließt.
- [x] `prefers-reduced-motion` stellt den Pop weiterhin still (globale Regel in
      `index.css` unverändert wirksam).

---

## Phase 2: Redesign-Robustheit auf Alt-Browsern und kurzen Viewports

### Context

- `frontend/src/index.css — @utility skeleton-shimmer` +
  `frontend/src/components/ui/skeleton.tsx — Skeleton()` — Gradient ohne soliden
  Fallback.
- `frontend/src/components/common/AuthLayout.tsx — AuthLayout()` —
  `overflow-hidden` am höhengedeckelten Wrapper clippt hohe Karten.
- `frontend/src/pages/PasswordPage.tsx` / das Passwort-Setzen-Formular — der
  höchste Auth-Inhalt, an dem das Clipping auftritt.

### What to build

Zwei Alt-Browser-/Kleingerät-Robustheitsfixes ohne sichtbare Änderung auf
modernen Browsern: (1) `skeleton-shimmer` bekommt eine solide
`background: var(--muted)`-Deklaration vor dem Gradient, sodass Skeletons bei
verworfener Gradient-Deklaration (fehlendes `color-mix()`) grau statt
transparent rendern; der Shimmer bleibt, wo unterstützt. (2) `AuthLayout`
verlagert das Glow-Clipping in eine eigene absolut positionierte
`inset-0 overflow-hidden -z-10 aria-hidden`-Ebene, die die drei Glow-Divs
aufnimmt; der äußere Container verliert `max-h-screen` und `overflow-hidden`
und behält `min-h-screen` mit normalem (scrollbarem) Überlauf, sodass hohe
Karten (Passwort setzen) auf kurzen und Landscape-Viewports vollständig
erreichbar bleiben.

### Acceptance criteria

- [x] Skeleton rendert bei simuliert fehlendem `color-mix()`-Support als solide
      `--muted`-Fläche (nicht transparent); auf modernen Browsern erscheint der
      Shimmer unverändert.
- [x] Das Passwort-Setzen-Formular ist auf 375×667 und in Landscape vollständig
      scrollbar erreichbar — „Passwort festlegen"-Button, „Zum Login"-Link und
      Footer sind nicht abgeschnitten (verifiziert per Screenshot bei 375×667
      und 667×375).
- [x] Die dekorativen Glows werden weiterhin am Rand geclippt (kein
      horizontaler Dokument-Überlauf durch die Glows) und bleiben `aria-hidden`
      / `print:hidden`.
- [x] Login-Screen (kurzer Inhalt) rendert visuell unverändert zu vorher.

---

## Phase 3: Vertragswerk und Website-Funnel konsistent machen

### Context

- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` — Betreff und Body der
  mailto-URL.
- `website/src/pages/fuer-vereine.astro — schritte` — die drei Prozess-Schritte.
- `website/src/components/AnfrageFormular.tsx` — Island, die den Button-Text
  rendert und `buildMailtoUrl` aufruft (Button-/CTA-Beschriftung).
- `TERMS.md — § 6(3)`, `TERMS.md — Prozess / E-Mail-Vorlage` — verbindlicher
  Wortlaut für Annahme und Fassungsbezug.
- `LICENSE — Section 9`, `LICENSE — Section 12` — anzugleichende Klauseln.

### What to build

Den öffentlichen Akquise-Pfad und die LICENSE an den in 713b247 eingeführten
Ein-E-Mail-Vertragsschluss angleichen: (1) `buildMailtoUrl` erzeugt eine
Annahme-E-Mail statt einer Anfrage — Betreff auf „Nutzungsvereinbarung jotti —
<Verein>" (Vorlagen-konform), und der Body enthält den Annahmesatz „… und
akzeptieren die Nutzungsbedingungen für jotti in der Fassung vom 14. Juli 2026
(<TERMS-URL>)" zusätzlich zu den Kontaktfeldern; die CTA-/Button-Beschriftung
wechselt von „anfragen" auf „abschließen"-Semantik. (2) Die drei Schritte in
`fuer-vereine.astro` beschreiben den Ein-E-Mail-Prozess (kein Autor-
Bestätigungsschritt mehr): eine einzige Annahme-E-Mail, danach direkt loslegen.
(3) `LICENSE` Section 9 erhält denselben Vorsatz/grobe-Fahrlässigkeit-Carve-out
wie `TERMS.md § 6(3)`; Section 12 ersetzt die geltungserhaltende Reduktion durch
eine schlichte Salvatorik (Unwirksames entfällt, der Rest bleibt wirksam).

### Acceptance criteria

- [x] Die vom Formular erzeugte mailto-URL enthält den wörtlichen Annahmesatz
      mit Fassungsdatum „14. Juli 2026" und der TERMS-URL; Betreff und
      Button-/CTA-Text spiegeln einen Vertragsabschluss, keine Anfrage.
- [x] `anfrage-mailto`-Tests decken den neuen Body ab (Annahmesatz vorhanden,
      Fassungsbezug vorhanden, Pflichtfeld-Validierung unverändert); bestehende
      Tests angepasst.
- [x] `fuer-vereine.astro` beschreibt keinen Autor-Bestätigungsschritt mehr und
      ist konsistent mit `TERMS.md`-Prozess und E-Mail-Vorlage.
- [x] `LICENSE` Section 9 enthält den Vorsatz/grobe-Fahrlässigkeit-Carve-out
      (wortgleich zur Intention von `TERMS.md § 6(3)`); Section 12 ist eine
      schlichte Salvatorik ohne geltungserhaltende Reduktion.
- [x] Kein weiterer Verweis im Repo beschreibt den alten Drei-Schritte-/
      Bestätigungs-Prozess (`grep` nach „Bestätigung", „anfragen",
      „Vereinbarung in Textform" in `website/` und Doku bereinigt).

---

## Phase 4: Admin-Layout auf Tablet und schmalen Phones (pre-existing)

### Context

- `frontend/src/admin/AdminLayout.tsx — AdminMobileHeader()` — Header-Sichtbarkeit
  (`lg:hidden` → `md:hidden`).
- `frontend/src/components/ui/sidebar.tsx`,
  `frontend/src/hooks/use-mobile.ts — MOBILE_BREAKPOINT` — Sidebar erscheint ab
  `md` (768px); Bezug für die Angleichung.
- `frontend/src/admin/reporting/KassenberichtePage.tsx` — DSFinV-K-Hinweiszeile
  mit ZIP-Button.
- `frontend/src/admin/kasse/LaufenderBetriebSection.tsx — BewegungZeile()` —
  `truncate`-Span im Flex-Row.
- `frontend/src/lib/utils.ts — formatCents()`, `formatAlleAuswaehlenLabel()` —
  Ziel für den neuen `formatEuro()`-Helper.
- Geldanzeige-Konsumenten: `LiveReportingSection.tsx`, `ReportingResults.tsx`,
  `SummaryCard.tsx`, `SitzungsListe.tsx`, `StornoItem.tsx`.

### What to build

Die vier bestandsalten Layout-Defekte im Admin-Bereich beheben, jeweils
verifiziert bei 768px mit sichtbarer Sidebar (~512px Inhaltsbreite): (1)
Doppel-Chrome auflösen — `AdminMobileHeader` ab `md` ausblenden (`md:hidden`),
sodass im 768–1023px-Band nur die Desktop-Sidebar das Chrome trägt. (2)
`KassenberichtePage`-Hinweiszeile darf umbrechen: die Zeile stapelt auf
schmalen Breiten (`flex-col`, ab `sm`/`md` wieder Row) bzw. erlaubt Umbruch,
sodass der ZIP-Button keinen horizontalen Überlauf mehr erzeugt. (3)
`BewegungZeile` so anpassen, dass der `truncate`-Span innerhalb der verfügbaren
Breite bleibt (korrekte `min-w-0`-Kette / Flex-Constraints), kein Überlauf bei
768px. (4) `formatEuro(cents)`-Helper einführen (`` `${formatCents(cents)} €` ``)
und die Betrag-plus-`€`-Anzeigestellen sowie `formatAlleAuswaehlenLabel` darauf
umstellen, damit `€` nie vom Betrag getrennt umbricht.

### Acceptance criteria

- [x] Im 768–1023px-Band zeigt der Admin-Bereich nur die Desktop-Sidebar, kein
      zusätzliches Hamburger-Header-Chrome; bei <768px bleibt der Hamburger-
      Header sichtbar und die Sidebar off-canvas (verifiziert bei 767px und
      768px).
- [x] `/admin/kassenberichte` erzeugt keinen horizontalen Überlauf mehr bei
      375×667, 360×800, 390×844, 414×896 und 768×1024 (mit sichtbarer Sidebar);
      `scrollWidth === clientWidth`.
- [x] `/admin/kasse` (Kassentag-Stepper, laufender Betrieb) erzeugt keinen
      horizontalen Überlauf mehr bei 768×1024; `BewegungZeile`-Text kürzt per
      Ellipse statt zu überlaufen.
- [x] `formatEuro()` existiert in `frontend/src/lib/utils.ts`; alle im Inventory
      genannten Anzeigestellen und `formatAlleAuswaehlenLabel` nutzen ihn;
      Betrag und `€` brechen in Kacheln (`SummaryCard`) und Steuertabelle
      (`ReportingResults`) nicht mehr getrennt um.
- [x] Frontend-Suite grün (`make check-frontend` bzw. `pnpm test --run` +
      `pnpm lint`); die vorab automatisiert gemessenen Overflow-Stellen der
      betroffenen Routen sind auf allen sieben Referenz-Viewports sauber.
