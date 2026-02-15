import { useState } from 'react'
import { toast } from 'sonner'

import { useActiveProducts } from '../../product/hooks'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { OrderDrawer } from './OrderDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface OrderProps {
  backend: Pick<TableBackend, 'placeTableOrder'>
  table: Table
  onOrderPlaced: () => void
}

type VariantQuantityMap = Record<number, number>

export function Order({ backend, table, onOrderPlaced }: OrderProps) {
  const { loading, products } = useActiveProducts()
  const [quantities, setQuantities] = useState<VariantQuantityMap>({})

  if (loading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <OrderDrawer
        backend={backend}
        table={table}
        products={products}
        quantities={quantities}
        orderPlaced={() => {
          setQuantities({})
          toast.success(`Bestellung wurde aufgegeben.`)
          onOrderPlaced()
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
