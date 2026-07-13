import { formatCents } from '@/lib/utils'

// VariantNamePreis ist das geteilte Name/Preis-Paar der Varianten-Listen —
// Admin-Produkt-Chips (VariantChip) und Service-Bestellen (VariantRow). Es MUSS
// in einem Flex-Container stehen: der Name wächst und kürzt sich bei Überlänge
// (min-w-0 flex-1 truncate) und schiebt den Preis an die feste Spaltenposition
// am rechten Rand; der Preis bleibt inhaltsbreit und dadurch unverdrängbar
// (shrink-0, tabular-nums für gleich breite Ziffern). Die Basis-Schriftgröße
// und die nachfolgende Aktion (Stepper bzw. Switch) liefert der aufrufende
// Container als weiteres Flex-Geschwister.
export function VariantNamePreis({
  name,
  preisCents,
}: {
  name: string
  preisCents: number
}) {
  return (
    <>
      <span className="min-w-0 flex-1 truncate font-medium">{name}</span>
      <span className="shrink-0 text-sm font-semibold tabular-nums">
        {formatCents(preisCents)}&nbsp;€
      </span>
    </>
  )
}
