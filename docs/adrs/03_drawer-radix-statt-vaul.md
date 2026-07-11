# ADR 03: Drawer auf Radix Dialog statt vaul

- **Status:** akzeptiert (2026-07-11)
- **Kontext-Dokumente:** PR #81 (Commit `bb938fb`); Umsetzungsplan
  `plan-pwa-ios-drawer-fixes.md` und Praxistest-Bugreport vom 09.07.2026
  (nach Merge gelöscht, siehe Git-Historie)

## Kontext

Alle Kern-Workflows (Bestellen, Kassieren, Stornieren, Umbuchen,
Direktverkauf) laufen über Bottom-Sheet-Drawer. Bis v0.15.0 basierten sie
auf der Bibliothek vaul. Der Praxistest bei einem Verein am 09.07.2026
zeigte in der installierten iOS-PWA (`display-mode: standalone`) massive
Fehler: Flackern beim Öffnen, versetzte `fixed`-Elemente, tote
Touch-Flächen. Im gewöhnlichen Browser-Tab trat nichts davon auf — der
Fehler ist nur auf echten Geräten mit installierter PWA sichtbar.

Ursache ist vauls Scroll-Lock: Beim Öffnen führt vaul ein unkompensiertes
`scrollTo(0, 0)` aus, das im Standalone-Modus von iOS Safari das Layout
verschiebt. Der Fehler ist upstream bekannt (vaul#505), unbeantwortet, und
vaul 1.1.2 ist das letzte Release. Zusätzlich fehlte den Drawern ein
definierter Scrollbereich: Bei vielen Positionen wanderte der
Primär-Button aus dem Viewport.

Erwogene Alternativen:

1. **vaul patchen oder workarounden** — Fork bzw. Monkey-Patch einer
   faktisch unmaintainten Bibliothek, mitten in den Kern-Workflows.
2. **Zweites Drawer-System nur für die iOS-PWA** — zwei Implementierungen
   desselben UI-Patterns, doppelte Pflege und Tests.
3. **Drawer als eigenes Bottom-Sheet auf Radix Dialog** — dessen
   Scroll-Lock (react-remove-scroll) funktioniert im iOS-Standalone-Modus
   korrekt (shadcn-ui#8507); vaul entfällt vollständig.

## Entscheidung

Der Drawer (`frontend/src/components/ui/drawer.tsx`) wird als eigenes
Bottom-Sheet auf Radix Dialog neu aufgebaut (Alternative 3); die
Dependency vaul wird vollständig entfernt. Es gibt genau ein
Drawer-System.

Der Layout-Vertrag der Komponente ist bindend, weil jede Abweichung die
Praxistest-Bugs zurückbringt:

- `max-h-[85dvh]` — bewusst `dvh` statt `vh`, wegen der dynamischen
  URL-Leiste von iOS Safari.
- `pb-[env(safe-area-inset-bottom,0px)]` — Home-Indicator-Bereich.
- `DrawerBody` ist der einzige Scrollbereich; Header und Footer sind
  direkte Flex-Kinder von `DrawerContent` und bleiben immer sichtbar.
- Kein Swipe-to-Dismiss und kein Drag-Handle: Radix Dialog unterstützt
  kein Ziehen; ein funktionsloser Handle würde Bedienbarkeit vortäuschen.

## Konsequenzen

- vaul wird für Drawer nicht wieder eingeführt; eine Rückkehr bräuchte ein
  neues ADR und einen Nachweis auf echtem iOS-Gerät als installierte PWA.
- Änderungen am Drawer-Layout sind auf einem echten iOS-Gerät als
  installierte PWA zu prüfen — Desktop-Browser und iOS-Browser-Tab zeigen
  den Fehler nicht. Den Scroll-Vertrag sichert zusätzlich eine
  Mobile-E2E-Regression ab (lange Positionsliste hält die Buttons im
  Viewport).
- Während eines laufenden Submits blockiert `pending` Escape- und
  Backdrop-Dismiss und dimmt den Body (Mechanik-Details als Kommentar in
  `drawer.tsx`).
