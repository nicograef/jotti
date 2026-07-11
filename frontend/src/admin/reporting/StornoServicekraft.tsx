import { Ban } from 'lucide-react'

import type { StornierungServicekraft } from './types'
import { formatBediener } from './utils'

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
// zusammen ("felix 1 · sophie 1") und steht über der Storno-Detail-Liste.
export function StornoAggregat({
  eintraege,
}: {
  eintraege: StornierungServicekraft[]
}) {
  return (
    <p className="mb-3 text-sm text-muted-foreground">
      {eintraege
        .map(
          (e) =>
            `${formatBediener(e.userName, e.name)} ${String(e.anzahlStornierungen)}`,
        )
        .join(' · ')}
    </p>
  )
}
