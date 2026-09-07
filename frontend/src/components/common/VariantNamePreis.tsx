import { formatEuro } from '@/lib/utils'

// VariantNamePreis ist das Name/Preis-Paar der Admin-Produkt-Chips; VariantChip
// ist sein einziger Consumer. Es MUSS in einem Flex-Container stehen: der Name
// wächst und kürzt sich bei Überlänge (min-w-0 flex-1 truncate) und schiebt den
// Preis an die feste Spaltenposition am rechten Rand; der Preis bleibt
// inhaltsbreit und dadurch unverdrängbar (shrink-0, tabular-nums für gleich
// breite Ziffern). Die Basis-Schriftgröße liefert der aufrufende Container; die
// Aktion des Chips (Switch) steht neben diesem Container, nicht im Paar.
// Die Bestellliste des Service nutzt das Paar bewusst nicht: dort bricht der
// Variantenname um, statt zu kürzen, weil zwei gekürzte Namen desselben
// Produkts gleich aussehen können (ProductList).
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
