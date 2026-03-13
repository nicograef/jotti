import { useState } from 'react'
import { toast } from 'sonner'

import type { Produkt } from '../../product/Product'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { BestellungDrawer } from './BestellungDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface BestellungProps {
  backend: Pick<TischBackend, 'bestellungAufgeben'>
  tisch: Tisch
  products: Produkt[]
  productsLoading: boolean
  onBestellungAufgegeben: () => void
}

type VariantMengenMap = Record<number, number>

export function Bestellung({
  backend,
  tisch,
  products,
  productsLoading,
  onBestellungAufgegeben,
}: BestellungProps) {
  const [mengen, setMengen] = useState<VariantMengenMap>({})

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
        bestellungAufgegeben={() => {
          setMengen({})
          toast.success(`Bestellung wurde aufgegeben.`)
          onBestellungAufgegeben()
        }}
      />
      <ProductList
        products={products}
        variantMengen={mengen}
        onAdd={(variantId) => {
          setMengen((prev) => ({
            ...prev,
            [variantId]: (prev[variantId] || 0) + 1,
          }))
        }}
        onRemove={(variantId) => {
          setMengen((prev) => {
            const aktuelleMenge = prev[variantId] || 0
            if (aktuelleMenge <= 0) return prev
            return {
              ...prev,
              [variantId]: aktuelleMenge - 1,
            }
          })
        }}
      />
    </>
  )
}
