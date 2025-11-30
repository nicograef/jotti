import { useState } from 'react'

import type { OrderProduct } from '@/lib/order/Order'
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
  const [productAmounts, setProductAmounts] = useState<ProductAmountMap>({})

  const placeOrder = async () => {
    try {
      const orderProducts: OrderProduct[] = Object.entries(productAmounts)
        .filter(([, amount]) => amount > 0)
        .map(([productId, amount]) => ({
          id: Number(productId),
          name:
            products.find((p) => p.id === Number(productId))?.name ?? 'Unknown',
          netPriceCents:
            products.find((p) => p.id === Number(productId))?.netPriceCents ??
            0,
          quantity: amount,
        }))

      const order = await backend.placeOrder({
        tableId: table.id,
        products: orderProducts,
      })

      console.log('Order placed successfully:', order)
    } catch (error: unknown) {
      console.error(error)
    }
  }

  if (loading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <OrderDrawer
        table={table}
        products={products}
        productsAmounts={productAmounts}
        onSubmit={() => {
          void placeOrder()
        }}
      />
      <ProductList
        products={products}
        productAmounts={productAmounts}
        onAdd={(productId) => {
          setProductAmounts((prev) => ({
            ...prev,
            [productId]: (prev[productId] || 0) + 1,
          }))
        }}
        onRemove={(productId) => {
          setProductAmounts((prev) => {
            const currentAmount = prev[productId] || 0
            if (currentAmount <= 0) return prev
            return {
              ...prev,
              [productId]: currentAmount - 1,
            }
          })
        }}
      />
    </>
  )
}
