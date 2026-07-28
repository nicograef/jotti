import { Ban } from 'lucide-react'

import { cn } from '@/lib/utils'

import { formatServicekraft } from './utils'

// StornoMarker markiert eine Servicekraft-Zeile mit Stornos als rotes
// Kontroll-Signal ("N Storno"). Wird nur bei ≥ 1 Storno gerendert und steht bei
// der betroffenen Servicekraft — nicht bei der, die storniert hat.
export function StornoMarker({ anzahl }: { anzahl: number }) {
  return (
    <span className="inline-flex w-fit items-center gap-1 text-xs font-medium text-destructive">
      <Ban className="size-3.5" />
      {anzahl} Storno
    </span>
  )
}

// StornoAggregatEintrag ist der Ausschnitt einer Abrechnungszeile, den das
// Aggregat braucht: die Servicekraft und ihr Storno-Zähler.
interface StornoAggregatEintrag {
  userId: number
  userName: string
  name: string
  anzahlStornierungen: number
}

// StornoAggregat fasst die Stornierungen pro Servicekraft in einer Zeile
// zusammen ("Betroffen: felix 1 · sophie 1"). Gespeist aus der Abrechnung pro
// Servicekraft (Einträge mit mindestens einem zugeordneten Storno); in der
// Übersicht bildet sie die eingeklappte Zusammenfassung der Storno-Detail-Liste.
// Das führende "Betroffen:" ist nötig, damit die Zeile nicht als Aufteilung der
// darüberstehenden Kopfkennzahl gelesen wird: Eine Korrektur zählt bei jedem
// Betroffenen, Direktverkauf-Stornos fehlen ganz.
export function StornoAggregat({
  eintraege,
  className,
}: {
  eintraege: StornoAggregatEintrag[]
  className?: string
}) {
  return (
    <p className={cn('mb-3 text-sm text-muted-foreground', className)}>
      Betroffen:{' '}
      {eintraege
        .map(
          (e) =>
            `${formatServicekraft(e.userName, e.name)} ${String(e.anzahlStornierungen)}`,
        )
        .join(' · ')}
    </p>
  )
}
