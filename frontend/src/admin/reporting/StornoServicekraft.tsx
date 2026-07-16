import { Ban } from 'lucide-react'

import { cn } from '@/lib/utils'

import type { StornierungServicekraft } from './types'
import { formatServicekraft } from './utils'

// StornoMarker markiert eine Servicekraft-Zeile mit Stornos als rotes
// Kontroll-Signal ("N Storno"). Wird nur bei ≥ 1 Storno gerendert.
export function StornoMarker({ anzahl }: { anzahl: number }) {
  return (
    <span className="mt-1.5 inline-flex w-fit items-center gap-1 text-xs font-medium text-destructive">
      <Ban className="size-3.5" />
      {anzahl} Storno
    </span>
  )
}

// StornoAggregat fasst die Stornierungen pro Servicekraft in einer Zeile
// zusammen ("felix 1 · sophie 1"). Auf der Kassenberichte-Seite steht sie über
// der Storno-Detail-Liste (Default-Abstand mb-3); in der Übersicht bildet sie
// die eingeklappte Zusammenfassung (className überschreibt den Abstand).
export function StornoAggregat({
  eintraege,
  className,
}: {
  eintraege: StornierungServicekraft[]
  className?: string
}) {
  return (
    <p className={cn('mb-3 text-sm text-muted-foreground', className)}>
      {eintraege
        .map(
          (e) =>
            `${formatServicekraft(e.userName, e.name)} ${String(e.anzahlStornierungen)}`,
        )
        .join(' · ')}
    </p>
  )
}
