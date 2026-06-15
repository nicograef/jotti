import { toast } from 'sonner'

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
  onBestellungAufgenommen: () => void
}

export function Bestellung({
  backend,
  tisch,
  products,
  productsLoading,
  onBestellungAufgenommen,
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
          toast.success(`Bestellung wurde aufgenommen.`)
          onBestellungAufgenommen()
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
