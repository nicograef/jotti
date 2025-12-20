import { useState } from 'react'
import { toast } from 'sonner'

import type { OrderBackend } from './order/OrderBackend'
import { OrderDrawer } from './OrderDrawer'
import { useActiveProducts } from './product/hooks'
import { ProductList, ProductListSkeleton } from './ProductList'
import type { Table } from './table/Table'

interface OrderProps {
  backend: Pick<OrderBackend, 'placeOrder'>
  table: Table
}

type ProductAmountMap = Record<number, number>

export function Order({ backend, table }: OrderProps) {
  const { loading, products } = useActiveProducts()
  const [quantities, setQuantities] = useState<ProductAmountMap>({})

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
        }}
      />
      <ProductList
        products={products}
        productAmounts={quantities}
        onAdd={(productId) => {
          setQuantities((prev) => ({
            ...prev,
            [productId]: (prev[productId] || 0) + 1,
          }))
        }}
        onRemove={(productId) => {
          setQuantities((prev) => {
            const currentQuantity = prev[productId] || 0
            if (currentQuantity <= 0) return prev
            return {
              ...prev,
              [productId]: currentQuantity - 1,
            }
          })
        }}
      />
    </>
  )
}
