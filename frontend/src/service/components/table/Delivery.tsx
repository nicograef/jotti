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

import { useTableUndeliveredProducts } from '../../table/hooks'
import type { OrderProduct } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { DeliveryDrawer } from './DeliveryDrawer'

interface DeliveryProps {
  backend: Pick<TableBackend, 'deliverTableProducts'>
  table: Table
  onProductsDelivered: () => void
}

export function Delivery({
  table,
  backend,
  onProductsDelivered,
}: DeliveryProps) {
  const { products, loading, reload } = useTableUndeliveredProducts(table.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const undeliveredQuantities: Record<number, number> = {}
  products.forEach((product) => {
    undeliveredQuantities[product.id] = product.quantity
  })

  const onAdd = (productId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[productId] || 0
      if (currentQuantity >= (undeliveredQuantities[productId] || 0))
        return prev
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
      {' '}
      <DeliveryDrawer
        backend={backend}
        table={table}
        undeliveredProducts={products}
        quantities={quantities}
        productsDelivered={() => {
          setQuantities({})
          toast.success(`Lieferung wurde registriert.`)
          onProductsDelivered()
          void reload()
        }}
      />
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 mt-4">
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
                undeliveredQuantity={undeliveredQuantities[product.id] || 0}
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
  undeliveredQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function ProductItem({
  product,
  quantity,
  undeliveredQuantity,
  onAdd,
  onRemove,
}: ProductItemProps) {
  return (
    <Item key={product.id} variant="outline">
      <ItemContent>
        <ItemTitle>{product.name}</ItemTitle>
        <ItemDescription>
          noch {undeliveredQuantity - quantity} zu liefern
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
