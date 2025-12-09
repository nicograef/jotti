import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { useTableUnpaidProducts } from '@/lib/order/hooks'
import type { OrderProduct } from '@/lib/order/Order'
import type { OrderBackend } from '@/lib/order/OrderBackend'
import type { TablePublic } from '@/lib/table/Table'

import { PaymentDrawer } from './PaymentDrawer'

interface PaymentProps {
  backend: Pick<OrderBackend, 'registerPayment'>
  table: TablePublic
}

export function Payment({ table, backend }: PaymentProps) {
  const { products, loading } = useTableUnpaidProducts(table.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  if (loading) {
    return <ProductListSkeleton />
  }

  const unpaidQuantities: Record<number, number> = {}
  products.forEach((product) => {
    unpaidQuantities[product.id] = product.quantity
  })

  return (
    <>
      <PaymentDrawer
        backend={backend}
        table={table}
        unpaidProducts={products}
        quantities={quantities}
        paymentRegistered={() => {
          setQuantities({})
          toast.success(`Zahlung wurde registriert.`)
        }}
      />
      <ProductList
        products={products}
        quantities={quantities}
        unpaidQuantities={unpaidQuantities}
        onAdd={(productId) => {
          setQuantities((prev) => {
            const currentQuantity = prev[productId] || 0
            if (currentQuantity >= (unpaidQuantities[productId] || 0))
              return prev
            return {
              ...prev,
              [productId]: currentQuantity + 1,
            }
          })
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

interface ProductListProps {
  products: OrderProduct[]
  quantities: Record<number, number>
  unpaidQuantities: Record<number, number>
  onAdd: (productId: number) => void
  onRemove: (productId: number) => void
}

function ProductList(props: ProductListProps) {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.products.map((product) => (
        <Item key={product.id} variant="outline">
          <ItemContent>
            <ItemTitle>{product.name}</ItemTitle>
            <ItemDescription>
              <span className="font-bold">
                {(product.netPriceCents / 100).toFixed(2)}&nbsp;€
              </span>
              &nbsp; &ndash; &nbsp;
              {props.unpaidQuantities[product.id] || 0} offen
            </ItemDescription>
          </ItemContent>
          <ItemActions>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full"
              aria-label="Produkt entfernen"
              onClick={() => {
                props.onRemove(product.id)
              }}
            >
              <Minus />
            </Button>
            <span className="text-lg mx-1">
              {props.quantities[product.id] || 0}
            </span>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full"
              aria-label="Produkt hinzufügen"
              onClick={() => {
                props.onAdd(product.id)
              }}
            >
              <Plus />
            </Button>
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  )
}

export function ProductListSkeleton() {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {Array.from({ length: 6 }).map((_, index) => (
        <Item key={`skeleton-${index.toString()}`} variant="outline">
          <ItemContent>
            <Skeleton className="h-4 w-24" />
          </ItemContent>
          <ItemActions>
            <Plus />
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  )
}
