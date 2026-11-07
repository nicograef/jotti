// Leerzustand der festen Abschluss-Spalte; im Handy-Drawer nie sichtbar (der
// öffnet ohne Auswahl nur für die Pfandrückgabe, die ihn ersetzt).
export function AbschlussLeer({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-40 items-center justify-center px-6 py-10 text-center text-sm text-muted-foreground">
      {children}
    </div>
  )
}
