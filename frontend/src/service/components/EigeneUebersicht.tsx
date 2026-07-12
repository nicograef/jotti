import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import type { EigeneUebersicht } from '../table/Tisch'

interface EigeneUebersichtProps {
  uebersicht: EigeneUebersicht
  loading: boolean
}

export function EigeneUebersichtKarten({
  uebersicht,
  loading,
}: EigeneUebersichtProps) {
  if (loading) {
    return (
      <div className="my-4 grid grid-cols-2 divide-x rounded-xl bg-muted/60 px-4 py-3">
        {[0, 1].map((spalte) => (
          <div key={spalte} className="px-4 first:pl-0 last:pr-0">
            <Skeleton className="mb-1 h-3 w-20" />
            <Skeleton className="h-5 w-28" />
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="my-4 grid grid-cols-2 divide-x rounded-xl bg-muted/60 px-4 py-3">
      <StatSpalte
        label="Bestellungen"
        anzahl={uebersicht.anzahlBestellungen}
        cents={uebersicht.bestellungenCents}
      />
      <StatSpalte
        label="Kassiert"
        anzahl={uebersicht.anzahlZahlungen}
        cents={uebersicht.zahlungenCents}
      />
    </div>
  )
}

function StatSpalte({
  label,
  anzahl,
  cents,
}: {
  label: string
  anzahl: number
  cents: number
}) {
  return (
    <div className="px-4 first:pl-0 last:pr-0">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="flex items-baseline gap-1.5">
        <span className="text-base font-bold tabular-nums">{anzahl}</span>
        <span className="text-[13px] text-muted-foreground tabular-nums">
          · {formatCents(cents)}&nbsp;€
        </span>
      </div>
    </div>
  )
}
