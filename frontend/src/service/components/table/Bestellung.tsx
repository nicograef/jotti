import { useMengen } from '@/hooks/use-mengen'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { BestellungDrawer } from './BestellungDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface BestellungProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  products: Produkt[]
  productsLoading: boolean
  // Meldet die erfolgreiche Buchung samt Bestätigungstext an die Seite, die den
  // Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg: (nachricht: string) => void
}

export function Bestellung({
  backend,
  tisch,
  products,
  productsLoading,
  onErfolg,
}: BestellungProps) {
  const { mengen, add, remove, reset } = useMengen<number>()

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <BestellungDrawer
        backend={backend}
        tisch={tisch}
        products={products}
        mengen={mengen}
        bestellungAufgenommen={() => {
          reset()
          onErfolg('Bestellung wurde aufgenommen.')
        }}
      />
      <ProductList
        products={products}
        variantMengen={mengen}
        onAdd={add}
        onRemove={remove}
      />
    </>
  )
}
