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

import { PaymentDrawer } from './PaymentDrawer'
import type { OrderProduct } from './table/Order'
import { useTableUnpaidProducts } from './table/orderHooks'
import type { Table } from './table/Table'
import type { TableBackend } from './table/TableBackend'

interface PaymentProps {
  backend: Pick<TableBackend, 'registerTablePayment'>
  table: Table
}

export function Payment({ table, backend }: PaymentProps) {
  const { products, loading } = useTableUnpaidProducts(table.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const unpaidQuantities: Record<number, number> = {}
  products.forEach((product) => {
    unpaidQuantities[product.id] = product.quantity
  })

  const onAdd = (productId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[productId] || 0
      if (currentQuantity >= (unpaidQuantities[productId] || 0)) return prev
      return {
        ...prev,
        [productId]: currentQuantity + 1,
      }
    })
  }

  const onRemove = (productId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[productId] || 0
      if (currentQuantity <= 0) return prev
      return {
        ...prev,
        [productId]: currentQuantity - 1,
      }
    })
  }

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
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ProductItemSkeleton key={index} />
            ))
          : products.map((product) => (
              <ProductItem
                key={product.id}
                product={product}
                quantity={quantities[product.id] || 0}
                unpaidQuantity={unpaidQuantities[product.id] || 0}
                onAdd={() => {
                  onAdd(product.id)
                }}
                onRemove={() => {
                  onRemove(product.id)
                }}
              />
            ))}
      </ItemGroup>
    </>
  )
}

interface ProductItemProps {
  product: OrderProduct
  quantity: number
  unpaidQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function ProductItem({
  product,
  quantity,
  unpaidQuantity,
  onAdd,
  onRemove,
}: ProductItemProps) {
  return (
    <Item key={product.id} variant="outline">
      <ItemContent>
        <ItemTitle>{product.name}</ItemTitle>
        <ItemDescription>
          <span className="font-bold">
            {(product.netPriceCents / 100).toFixed(2)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;noch {unpaidQuantity - quantity} offen
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Produkt entfernen"
          onClick={onRemove}
        >
          <Minus />
        </Button>
        <span className="text-lg mx-1">{quantity}</span>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Produkt hinzufügen"
          onClick={onAdd}
        >
          <Plus />
        </Button>
      </ItemActions>
    </Item>
  )
}

function ProductItemSkeleton() {
  return (
    <Item variant="outline">
      <ItemContent>
        <Skeleton className="h-4 w-24" />
      </ItemContent>
      <ItemActions>
        <Minus />
        <span className="text-lg mx-1">&nbsp;</span>
        <Plus />
      </ItemActions>
    </Item>
  )
}
