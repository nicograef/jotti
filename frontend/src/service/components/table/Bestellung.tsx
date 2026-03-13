import { useState } from 'react'
import { toast } from 'sonner'

import { useActiveProducts } from '../../product/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { BestellungDrawer } from './BestellungDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface BestellungProps {
  backend: Pick<TischBackend, 'bestellungAufgeben'>
  tisch: Tisch
  onBestellungAufgegeben: () => void
}

type VariantMengenMap = Record<number, number>

export function Bestellung({
  backend,
  tisch,
  onBestellungAufgegeben,
}: BestellungProps) {
  const { loading, products } = useActiveProducts()
  const [mengen, setMengen] = useState<VariantMengenMap>({})

  if (loading) {
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
