// Zweispaltiges Service-Layout ab lg (1024px, ADR 07/08): links die Auswahl,
// rechts die dauerhaft sichtbare Abschluss-Spalte. Beide Spalten scrollen
// unabhängig; die Höhe ist an den Viewport gebunden (abzüglich Header und der
// Reiter-Zeile darüber), damit unter jeder Spalte ein eigener Scrollbereich
// entsteht statt eines gemeinsamen Seitenscrolls. Nur ab lg gerendert — die
// aufrufende Fläche entscheidet per useIsMobile, welcher Container mountet.
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
    <div className="grid h-[calc(100dvh-9rem)] grid-cols-[minmax(0,1fr)_22rem] gap-6 xl:h-[calc(100dvh-10rem)] xl:grid-cols-[minmax(0,1fr)_26rem]">
      <div className="min-h-0 overflow-y-auto">{auswahl}</div>
      {abschluss}
    </div>
  )
}
