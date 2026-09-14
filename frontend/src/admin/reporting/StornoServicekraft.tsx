import { Ban } from 'lucide-react'

import { formatServicekraft } from './utils'

// Steht bei der betroffenen Servicekraft — nicht bei der, die storniert hat.
export function StornoMarker({ anzahl }: { anzahl: number }) {
  return (
    <span className="inline-flex w-fit items-center gap-1 text-xs font-medium text-destructive">
      <Ban className="size-3.5" />
      {anzahl} Storno
    </span>
  )
}

interface StornoAggregatEintrag {
  userId: number
  userName: string
  name: string
  anzahlStornierungen: number
}

// Das führende "Betroffen:" ist nötig, damit die Zeile nicht als Aufteilung der
// darüberstehenden Kopfkennzahl gelesen wird: Eine Korrektur zählt bei jedem
// Betroffenen, Direktverkauf-Stornos fehlen ganz.
export function StornoAggregat({
  eintraege,
}: {
  eintraege: StornoAggregatEintrag[]
}) {
  return (
    <p className="mb-0 mt-0.5 text-sm text-muted-foreground">
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
