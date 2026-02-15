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
import { AuthSingleton } from '@/lib/Auth'
import { formatCents } from '@/lib/utils'

import { useTableUnpaidVariants } from '../../table/hooks'
import type { LineItem } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { CancelationDrawer } from './CancelationDrawer'
import { PaymentDrawer } from './PaymentDrawer'

interface PaymentProps {
  backend: Pick<TableBackend, 'registerTablePayment' | 'cancelTableVariants'>
  table: Table
  onPaymentRegistered: () => void
  onVariantsCanceled: () => void
}

export function Payment({
  table,
  backend,
  onPaymentRegistered,
  onVariantsCanceled,
}: PaymentProps) {
  const { variants, loading, reload } = useTableUnpaidVariants(table.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const unpaidQuantities: Record<number, number> = {}
  variants.forEach((variant) => {
    unpaidQuantities[variant.id] = variant.quantity
  })

  const onAdd = (variantId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[variantId] || 0
      if (currentQuantity >= (unpaidQuantities[variantId] || 0)) return prev
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
      <div className="flex gap-2">
        {AuthSingleton.canCancel && (
          <div className="flex-1">
            <CancelationDrawer
              backend={backend}
              table={table}
              unpaidVariants={variants}
              quantities={quantities}
              variantsCanceled={() => {
                setQuantities({})
                toast.success(`Stornierung erfolgreich.`)
                onVariantsCanceled()
                reload()
              }}
            />
          </div>
        )}
        <div className="flex-1">
          <PaymentDrawer
            backend={backend}
            table={table}
            unpaidVariants={variants}
            quantities={quantities}
            paymentRegistered={() => {
              setQuantities({})
              toast.success(`Zahlung erfolgreich.`)
              onPaymentRegistered()
              reload()
            }}
          />
        </div>
      </div>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
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
                unpaidQuantity={unpaidQuantities[variant.id] || 0}
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
  unpaidQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function VariantItem({
  variant,
  quantity,
  unpaidQuantity,
  onAdd,
  onRemove,
}: VariantItemProps) {
  return (
    <Item key={variant.id} variant="outline">
      <ItemContent>
        <ItemTitle>{variant.name}</ItemTitle>
        <ItemDescription>
          <span className="font-bold">
            {formatCents(variant.priceCents)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;noch {unpaidQuantity - quantity} unbezahlt
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
