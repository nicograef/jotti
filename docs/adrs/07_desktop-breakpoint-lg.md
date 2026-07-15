# ADR 07: Einheitliche Desktop-Schwelle bei `lg` (1024px)

- **Status:** akzeptiert (2026-07-15)
- **Kontext-Dokumente:** `docs/prds/prd-service-split-screen-tablet.md`;
  Admin-UI-Breakpoint-Analyse (2026-07-15)

## Kontext

Das Frontend kannte zwei „Desktop"-Schwellen:

- **Shell/Sidebar** (Admin): Der Wechsel von mobiler Drawer-Navigation zur
  persistenten Sidebar lief bei `md` (768px), gesteuert über `useIsMobile`
  (`MOBILE_BREAKPOINT`).
- **Zweispaltige Inhalte:** Die Master-Detail-Ansichten des Admin (Kassenberichte,
  Benutzer) und der geplante Service-Split
  (`prd-service-split-screen-tablet.md`) klappen erst bei `lg` (1024px) auf.

Dazwischen (768–1023px) trug der Admin bereits die feste Sidebar (16rem),
während seine Inhalte noch einspaltig blieben — ein unnötig enger Streifen.
Zweispaltigkeit bei 768px wäre zu gedrängt; der Service-Split hat `lg` bewusst
gewählt.

## Entscheidung

Es gibt genau eine „Desktop"-Schwelle: `lg` (1024px).

- Unter 1024px gilt als mobil/Tablet: Drawer-Navigation, einspaltige Inhalte,
  Bottom-Sheets.
- Ab 1024px: persistente Sidebar und zweispaltige Inhalte.

Umgesetzt über `useIsMobile` (`MOBILE_BREAKPOINT = 1024`), das den
Drawer-/Persistent-Wechsel der Sidebar steuert, und den mobilen Admin-Kopf
(`lg:hidden`). Die zweispaltigen Inhalte liegen bereits auf `lg`.

Die vendorte shadcn-Sidebar (`src/components/ui/sidebar.tsx`, bewusst aus
Formatierung und Linting ausgenommen) bleibt unverändert: Ihre internen
`md:`-Klassen greifen nur im Nicht-mobil-Zweig, der dank `useIsMobile` erst ab
1024px gerendert wird — sie sind damit effektiv `lg`-gegatet. Die Schwelle wird
über die vorgesehene Eingabe (`useIsMobile`) gesteuert, nicht durch Forken der
Primitive.

## Konsequenzen

- Der enge 768–1023px-Streifen entfällt: Querformat-Tablets in diesem Bereich
  bekommen volle Inhaltsbreite mit Drawer-Navigation statt einer festen Sidebar.
- Eine kohärente Linie über Admin-Shell, Admin-Inhalte und Service-Split; das
  Service-Split-PRD referenziert dieselbe Schwelle.
- Das Padding-Idiom (`md:px-8`, `xl:px-12`) bleibt orthogonal und unverändert; es
  steuert Abstände, nicht den Layout-Modus.
- Auf Landscape-Tablets (768–1023px) ist die Sidebar nun ein Hamburger-Drawer
  statt einer festen Leiste. Für die selten genutzte Admin-Navigation ist die
  volle Inhaltsbreite der bessere Tausch.
