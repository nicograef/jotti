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

import { useTableUndeliveredVariants } from '../../table/hooks'
import type { LineItem } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { DeliveryDrawer } from './DeliveryDrawer'

interface DeliveryProps {
  backend: Pick<TableBackend, 'deliverTableVariants'>
  table: Table
  onVariantsDelivered: () => void
}

export function Delivery({
  table,
  backend,
  onVariantsDelivered,
}: DeliveryProps) {
  const { variants, loading, reload } = useTableUndeliveredVariants(table.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const undeliveredQuantities: Record<number, number> = {}
  variants.forEach((variant) => {
    undeliveredQuantities[variant.id] = variant.quantity
  })

  const onAdd = (variantId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[variantId] || 0
      if (currentQuantity >= (undeliveredQuantities[variantId] || 0))
        return prev
      return {
        ...prev,
        [variantId]: currentQuantity + 1,
      }
    })
  }

  const onRemove = (variantId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[variantId] || 0
      if (currentQuantity <= 0) return prev
      return {
        ...prev,
        [variantId]: currentQuantity - 1,
      }
    })
  }

  return (
    <>
      {' '}
      <DeliveryDrawer
        backend={backend}
        table={table}
        undeliveredVariants={variants}
        quantities={quantities}
        variantsDelivered={() => {
          setQuantities({})
          toast.success(`Lieferung wurde registriert.`)
          onVariantsDelivered()
          reload()
        }}
      />
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 mt-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <VariantItemSkeleton key={index} />
            ))
          : variants.map((variant) => (
              <VariantItem
                key={variant.id}
                variant={variant}
                quantity={quantities[variant.id] || 0}
                undeliveredQuantity={undeliveredQuantities[variant.id] || 0}
                onAdd={() => {
                  onAdd(variant.id)
                }}
                onRemove={() => {
                  onRemove(variant.id)
                }}
              />
            ))}
      </ItemGroup>
    </>
  )
}

interface VariantItemProps {
  variant: LineItem
  quantity: number
  undeliveredQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function VariantItem({
  variant,
  quantity,
  undeliveredQuantity,
  onAdd,
  onRemove,
}: VariantItemProps) {
  return (
    <Item key={variant.id} variant="outline">
      <ItemContent>
        <ItemTitle>{variant.name}</ItemTitle>
        <ItemDescription>
          noch {undeliveredQuantity - quantity} zu liefern
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Variante entfernen"
          onClick={onRemove}
        >
          <Minus />
        </Button>
        <span className="text-lg mx-1">{quantity}</span>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Variante hinzufügen"
          onClick={onAdd}
        >
          <Plus />
        </Button>
      </ItemActions>
    </Item>
  )
}

function VariantItemSkeleton() {
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
