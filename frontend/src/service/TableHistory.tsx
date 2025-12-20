import { Eye } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'

import type { Order } from './table/Order'
import { useTableOrders } from './table/orderHooks'

interface TableHistoryProps {
  tableId: number
  userId: number | null
}

const initialOrderDetailsState: {
  order: Order | null
  open: boolean
} = {
  order: null,
  open: false,
}

export function TableHistory({ tableId, userId }: TableHistoryProps) {
  const { loading, orders } = useTableOrders(tableId)
  const [orderDetails, setOrderDetails] = useState(initialOrderDetailsState)

  const sortedOrders = orders.sort((a, b) => {
    return new Date(b.placedAt).getTime() - new Date(a.placedAt).getTime()
  })

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <HistoryItemSkeleton key={index} />
            ))
          : sortedOrders.map((order) => (
              <HistoryItem
                key={order.id}
                userId={userId}
                order={order}
                onClick={() => {
                  setOrderDetails({ order, open: true })
                }}
              />
            ))}
      </ItemGroup>
      <OrderDetails
        order={orderDetails.order}
        open={orderDetails.open}
        onClose={() => {
          setOrderDetails(initialOrderDetailsState)
        }}
      />
    </>
  )
}

function HistoryItem({
  userId,
  order,
  onClick,
}: {
  userId: number | null
  order: Order
  onClick: () => void
}) {
  return (
    <Item
      variant="outline"
      className={order.userId === userId ? 'border-primary' : ''}
    >
      <ItemContent>
        <ItemTitle>Bestellung {order.id.slice(0, 8)}</ItemTitle>
        <ItemDescription>
          <span className="font-bold">
            {(order.totalNetPriceCents / 100).toFixed(2)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;am {new Date(order.placedAt).toLocaleString()}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full cursor-pointer"
          aria-label="Details anzeigen"
          onClick={onClick}
        >
          <Eye />
        </Button>
      </ItemActions>
    </Item>
  )
}

function HistoryItemSkeleton() {
  return (
    <Item variant="outline">
      <ItemContent>
        <ItemTitle>
          <Skeleton className="h-6 w-32" />
        </ItemTitle>
        <Skeleton className="h-4 w-48" />
      </ItemContent>
      <ItemActions>
        <Skeleton className="h-8 w-8 rounded-full" />
      </ItemActions>
    </Item>
  )
}

interface OrderDetailsProps {
  order: Order | null
  open: boolean
  onClose: () => void
}

function OrderDetails({ order, open, onClose }: OrderDetailsProps) {
  if (!order) return null

  return (
    <Drawer
      open={open}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Bestellung {order.id.slice(0, 8)}</DrawerTitle>
            <DrawerDescription>
              Aufgegeben am {new Date(order.placedAt).toLocaleDateString()} um{' '}
              {new Date(order.placedAt).toLocaleTimeString()} Uhr
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 space-y-2">
            {order.products.map((product) => {
              return (
                <div
                  key={product.id}
                  className="flex justify-between border-b pb-2"
                >
                  <div>
                    {product.quantity} x {product.name}
                  </div>
                  <div>
                    €{' '}
                    {((product.netPriceCents / 100) * product.quantity).toFixed(
                      2,
                    )}
                  </div>
                </div>
              )
            })}
            <div className="flex justify-between font-bold pt-2">
              <div>Gesamt</div>
              <div>€ {(order.totalNetPriceCents / 100).toFixed(2)}</div>
            </div>
          </div>
          <DrawerFooter>
            <DrawerClose asChild>
              <Button variant="outline">Schließen</Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
