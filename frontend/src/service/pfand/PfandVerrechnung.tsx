import { formatEuro } from '@/lib/utils'

interface PfandVerrechnungProps {
  kaufsummeCents: number
  pfandCents: number
}

// Netto-Zeile eines verrechneten Abschlusses: Ab einer Differenz von null
// zahlt die Kasse den Betrag aus.
export function PfandVerrechnung(props: PfandVerrechnungProps) {
  const differenzCents = props.kaufsummeCents - props.pfandCents
  return (
    <dl
      aria-label="Verrechnung"
      className="flex flex-col gap-1 px-4 pt-3 text-sm tabular-nums"
    >
      <div className="flex items-baseline justify-between gap-3">
        <dt>Kaufsumme</dt>
        <dd>{formatEuro(props.kaufsummeCents)}</dd>
      </div>
      <div className="flex items-baseline justify-between gap-3">
        <dt>Pfand zurück</dt>
        <dd>{formatEuro(-props.pfandCents)}</dd>
      </div>
      <div className="flex items-baseline justify-between gap-3 pt-1 text-[15px] font-semibold">
        <dt>{differenzCents > 0 ? 'Zu zahlen' : 'Auszahlen'}</dt>
        <dd className="text-xl font-bold">
          {formatEuro(Math.abs(differenzCents))}
        </dd>
      </div>
    </dl>
  )
}
