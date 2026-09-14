import { formatEuro } from '@/lib/utils'

// MUSS in einem Flex-Container stehen: Der Name wächst und kürzt sich
// (min-w-0 flex-1 truncate), der Preis bleibt inhaltsbreit (shrink-0).
// Die Bestellliste des Service nutzt das Paar nicht: Dort bricht der
// Variantenname um, weil zwei gekürzte Namen gleich aussehen können.
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
        {formatEuro(preisCents)}
      </span>
    </>
  )
}
