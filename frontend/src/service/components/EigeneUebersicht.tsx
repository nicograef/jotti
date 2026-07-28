import { Undo2 } from 'lucide-react'

import { Skeleton } from '@/components/ui/skeleton'
import { formatEuro } from '@/lib/utils'

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
    <div className="my-4">
      <div className="grid grid-cols-2 divide-x rounded-xl bg-muted/60 px-4 py-3">
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
      {uebersicht.anzahlRuecknahmen > 0 && (
        <RuecknahmeHinweis uebersicht={uebersicht} />
      )}
    </div>
  )
}

// RuecknahmeHinweis erklärt der Servicekraft eine Rücknahme, die auf ihre Kasse geht —
// auch wenn Admin oder Serviceleitung sie stellvertretend gebucht haben. Erscheint nur
// bei mindestens einer zugeordneten Rücknahme; sonst bleibt die Übersicht unverändert.
function RuecknahmeHinweis({ uebersicht }: { uebersicht: EigeneUebersicht }) {
  const anzahl = uebersicht.anzahlRuecknahmen
  return (
    <p className="mt-2 rounded-xl bg-muted/60 px-4 py-2 text-sm text-muted-foreground">
      <Undo2 className="mr-1.5 inline size-4 align-[-3px] text-destructive" />
      {anzahl === 1
        ? 'Eine Rücknahme'
        : `${String(anzahl)} Rücknahmen`} über{' '}
      <span className="font-medium text-foreground tabular-nums">
        {formatEuro(uebersicht.ruecknahmenCents)}
      </span>{' '}
      wurde{anzahl === 1 ? '' : 'n'} von deinen Zahlungen zurückgegeben. Du
      gibst damit{' '}
      <span className="font-semibold text-foreground tabular-nums">
        {formatEuro(uebersicht.abzugebenCents)}
      </span>{' '}
      ab.
    </p>
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
          · {formatEuro(cents)}
        </span>
      </div>
    </div>
  )
}
