// Two-column Service layout from lg (docs/decisions.md D07/D08), rendered only there; useIsMobile decides in the caller.
// `h-full` takes the height from the page's height-bounded flex container, so header or tab height changes need no calc.
export function ServiceSplitLayout({
  auswahl,
  abschluss,
}: {
  auswahl: React.ReactNode
  // Ein <aside> aus der jeweiligen Abschluss-Inhaltskomponente, variant="spalte".
  abschluss: React.ReactNode
}) {
  return (
    <div className="grid h-full grid-cols-[minmax(0,1fr)_22rem] gap-6 xl:grid-cols-[minmax(0,1fr)_26rem]">
      <div className="min-h-0 overflow-y-auto">{auswahl}</div>
      {abschluss}
    </div>
  )
}
