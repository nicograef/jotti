import { useState } from 'react'

import type { ProductPublic } from '@/lib/product/Product'
import type { TablePublic } from '@/lib/table/Table'

import { OrderDrawer } from './OrderDrawer'
import { ProductList, ProductListSkeleton } from './ProductList'

interface OrderProps {
  loading: boolean
  table: TablePublic
  products: ProductPublic[]
}

type ProductAmountMap = Record<number, number>

export function Order({ loading, products, table }: OrderProps) {
  const [productAmounts, setProductAmounts] = useState<ProductAmountMap>({})

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
          console.log('Order submitted')
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
