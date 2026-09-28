// Zweispaltiges Service-Layout ab lg (1024 px, siehe docs/decisions.md D07/D08):
// links die Auswahl, rechts die dauerhaft sichtbare Abschluss-Spalte. Die Höhe
// kommt per `h-full` vom höhenbegrenzten Flex-Container der Seite statt aus einem
// eigenen calc — so stimmt sie weiter, wenn Header- oder Reiter-Höhe sich ändern.
// Nur ab lg gerendert; useIsMobile entscheidet im Aufrufer.
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
