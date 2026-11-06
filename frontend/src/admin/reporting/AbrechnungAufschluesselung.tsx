import { formatEuro } from '@/lib/utils'

// Woraus sich der abzugebende Betrag einer Servicekraft zusammensetzt.
export function AbrechnungAufschluesselung({
  kassiertCents,
  ruecknahmenCents,
  pfandCents,
}: {
  kassiertCents: number
  ruecknahmenCents: number
  pfandCents: number
}) {
  return (
    <span className="text-xs text-muted-foreground">
      Kassiert {formatEuro(kassiertCents)} · Rücknahmen{' '}
      {formatEuro(ruecknahmenCents)}
      {pfandCents !== 0 && ` · Pfand ${formatEuro(pfandCents)}`}
    </span>
  )
}
