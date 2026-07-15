// Leerzustand der festen Abschluss-Spalte: ein Hinweistext, solange nichts
// ausgewählt ist. Gleichbedeutend mit dem heute deaktivierten Dock-Button —
// der Aktionsbutton der Spalte ist in diesem Zustand ebenfalls deaktiviert.
// Im Handy-Drawer nie sichtbar (der öffnet nur mit Auswahl).
export function AbschlussLeer({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-40 items-center justify-center px-6 py-10 text-center text-sm text-muted-foreground">
      {children}
    </div>
  )
}
