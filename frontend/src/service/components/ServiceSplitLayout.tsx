// Zweispaltiges Service-Layout ab lg (1024px, ADR 07/08): links die Auswahl,
// rechts die dauerhaft sichtbare Abschluss-Spalte. Beide Spalten scrollen
// unabhängig. Die Höhe kommt über `h-full` vom höhenbegrenzten Flex-Container
// der Seite (Viewport minus Header und Reiter-Zeile), statt aus einem fest
// verdrahteten calc — so bleibt das Layout robust, wenn sich Header oder
// Reiter-Höhe ändern. Nur ab lg gerendert: die aufrufende Fläche entscheidet
// per useIsMobile, welcher Container mountet.
export function ServiceSplitLayout({
  auswahl,
  abschluss,
}: {
  auswahl: React.ReactNode
  // Die Abschluss-Spalte selbst (ein <aside> aus der jeweiligen
  // Abschluss-Inhaltskomponente, variant="spalte").
  abschluss: React.ReactNode
}) {
  return (
    <div className="grid h-full grid-cols-[minmax(0,1fr)_22rem] gap-6 xl:grid-cols-[minmax(0,1fr)_26rem]">
      <div className="min-h-0 overflow-y-auto">{auswahl}</div>
      {abschluss}
    </div>
  )
}
