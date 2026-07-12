import type { ReactNode } from 'react'

// Einheitlicher Seitenkopf aller acht Admin-Seiten (Design-Handoff, Abschnitte
// 1a–1h): H1 24 px/700, darunter eine erklärende Unterzeile 14 px in
// muted-foreground und rechts ein Aktions-Slot (etwa der Anlegen-Button).
// Ersetzt die früheren losen H1 der einzelnen Seiten und den fixierten FAB.
export function AdminPageHeader({
  titel,
  unterzeile,
  aktionen,
}: {
  titel: string
  unterzeile?: ReactNode
  aktionen?: ReactNode
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div>
        <h1 className="text-2xl font-bold leading-8">{titel}</h1>
        {unterzeile !== undefined && (
          <p className="mt-1 text-sm text-muted-foreground">{unterzeile}</p>
        )}
      </div>
      {aktionen !== undefined && (
        <div className="flex shrink-0 items-center gap-2">{aktionen}</div>
      )}
    </div>
  )
}
