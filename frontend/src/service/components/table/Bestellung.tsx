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

type VariantQuantityMap = Record<number, number>

export function Bestellung({
  backend,
  tisch,
  onBestellungAufgegeben,
}: BestellungProps) {
  const { loading, products } = useActiveProducts()
  const [quantities, setQuantities] = useState<VariantQuantityMap>({})

  if (loading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <BestellungDrawer
        backend={backend}
        tisch={tisch}
        products={products}
        quantities={quantities}
        bestellungAufgegeben={() => {
          setQuantities({})
          toast.success(`Bestellung wurde aufgegeben.`)
          onBestellungAufgegeben()
        }}
      />
      <ProductList
        products={products}
        variantQuantities={quantities}
        onAdd={(variantId) => {
          setQuantities((prev) => ({
            ...prev,
            [variantId]: (prev[variantId] || 0) + 1,
          }))
        }}
        onRemove={(variantId) => {
          setQuantities((prev) => {
            const currentQuantity = prev[variantId] || 0
            if (currentQuantity <= 0) return prev
            return {
              ...prev,
              [variantId]: currentQuantity - 1,
            }
          })
        }}
      />
    </>
  )
}
