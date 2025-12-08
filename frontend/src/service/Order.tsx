import { useState } from 'react'
import { toast } from 'sonner'

import type { OrderBackend } from '@/lib/order/OrderBackend'
import type { ProductPublic } from '@/lib/product/Product'
import type { TablePublic } from '@/lib/table/Table'

import { OrderDrawer } from './OrderDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface OrderProps {
  backend: Pick<OrderBackend, 'placeOrder'>
  loading: boolean
  table: TablePublic
  products: ProductPublic[]
}

type ProductAmountMap = Record<number, number>

export function Order({ backend, loading, products, table }: OrderProps) {
  const [quantities, setQuantities] = useState<ProductAmountMap>({})
  const [drawerOpen, setDrawerOpen] = useState(false)

  if (loading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <OrderDrawer
        open={drawerOpen}
        backend={backend}
        table={table}
        products={products}
        quantities={quantities}
        cancel={() => {
          setDrawerOpen(false)
        }}
        orderPlaced={() => {
          setDrawerOpen(false)
          setQuantities({})
          toast.success(`Bestellung für ${table.name} wurde aufgegeben.`)
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
