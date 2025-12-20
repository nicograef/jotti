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

import { useTableOrders, useTablePayments } from './table/hooks'
import type { Order } from './table/Order'
import type { Payment } from './table/Payment'

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

const initialPaymentDetailsState: {
  payment: Payment | null
  open: boolean
} = {
  payment: null,
  open: false,
}

export function TableHistory({ tableId, userId }: TableHistoryProps) {
  const { loading: ordersLoading, orders } = useTableOrders(tableId)
  const { loading: paymentsLoading, payments } = useTablePayments(tableId)
  const [orderDetails, setOrderDetails] = useState(initialOrderDetailsState)
  const [paymentDetails, setPaymentDetails] = useState(
    initialPaymentDetailsState,
  )

  const sortedItems: (Order | Payment)[] = [
    ...orders.map((o) => ({ ...o, time: o.placedAt })),
    ...payments.map((p) => ({ ...p, time: p.registeredAt })),
  ].sort((a, b) => {
    return new Date(b.time).getTime() - new Date(a.time).getTime()
  })

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {ordersLoading || paymentsLoading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : sortedItems.map((item) => {
              if (Object.prototype.hasOwnProperty.call(item, 'registeredAt')) {
                return (
                  <PaymentItem
                    key={item.id}
                    payment={item as Payment}
                    userId={userId}
                    onClick={() => {
                      setPaymentDetails({
                        payment: item as Payment,
                        open: true,
                      })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'placedAt')
              ) {
                return (
                  <OrderItem
                    key={item.id}
                    order={item as Order}
                    userId={userId}
                    onClick={() => {
                      setOrderDetails({ order: item as Order, open: true })
                    }}
                  />
                )
              } else {
                return null
              }
            })}
      </ItemGroup>
      <OrderDetails
        order={orderDetails.order}
        userId={userId}
        open={orderDetails.open}
        onClose={() => {
          setOrderDetails(initialOrderDetailsState)
        }}
      />
      <PaymentDetails
        payment={paymentDetails.payment}
        userId={userId}
        open={paymentDetails.open}
        onClose={() => {
          setPaymentDetails(initialPaymentDetailsState)
        }}
      />
    </>
  )
}

function OrderItem({
  order,
  userId,
  onClick,
}: {
  order: Order
  userId: number | null
  onClick: () => void
}) {
  return (
    <Item variant="outline" className="border-amber-500">
      <ItemContent>
        <ItemTitle>
          Bestellung +{(order.totalNetPriceCents / 100).toFixed(2)}&nbsp;€
        </ItemTitle>
        <ItemDescription>
          {new Date(order.placedAt).toLocaleString()}
          {userId === order.userId ? <>&nbsp; &ndash; &nbsp;von Dir</> : ''}
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

function PaymentItem({
  payment,
  userId,
  onClick,
}: {
  payment: Payment
  userId: number | null
  onClick: () => void
}) {
  return (
    <Item variant="outline" className="border-green-500">
      <ItemContent>
        <ItemTitle>
          Zahlung -{(payment.totalPaymentCents / 100).toFixed(2)}
          &nbsp;€
        </ItemTitle>
        <ItemDescription>
          {new Date(payment.registeredAt).toLocaleString()}
          {userId === payment.userId ? <>&nbsp; &ndash; &nbsp;von Dir</> : ''}
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

function ItemSkeleton() {
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
  userId: number | null
  open: boolean
  onClose: () => void
}

function OrderDetails({ order, userId, open, onClose }: OrderDetailsProps) {
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
            <DrawerTitle>
              Bestellung {order.id.slice(0, 8)}
              {userId === order.userId ? ' von Dir' : ''}
            </DrawerTitle>
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

interface PaymentDetailsProps {
  payment: Payment | null
  userId: number | null
  open: boolean
  onClose: () => void
}

function PaymentDetails({
  payment,
  userId,
  open,
  onClose,
}: PaymentDetailsProps) {
  if (!payment) return null

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
            <DrawerTitle>
              Zahlung {payment.id.slice(0, 8)}{' '}
              {userId === payment.userId ? ' von Dir' : ''}
            </DrawerTitle>
            <DrawerDescription>
              Registriert am{' '}
              {new Date(payment.registeredAt).toLocaleDateString()} um{' '}
              {new Date(payment.registeredAt).toLocaleTimeString()} Uhr
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 space-y-2">
            {payment.products.map((product) => {
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
              <div>€ {(payment.totalPaymentCents / 100).toFixed(2)}</div>
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
