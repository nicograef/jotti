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

import type { Cancelation } from '../../table/Cancelation'
import type { Delivery } from '../../table/Delivery'
import { useTableHistory } from '../../table/hooks'
import type { Order, OrderProduct } from '../../table/Order'
import type { Payment } from '../../table/Payment'
import { Comment } from './CommentField'
import { Receipt } from './Receipt'

interface TableHistoryProps {
  tableId: number
  userId: number | null
}

const initialOrderState: {
  order: Order | null
  open: boolean
} = {
  order: null,
  open: false,
}

const initialPaymentState: {
  payment: Payment | null
  open: boolean
} = {
  payment: null,
  open: false,
}

const initialCancelationState: {
  cancelation: Cancelation | null
  open: boolean
} = {
  cancelation: null,
  open: false,
}

const initialDeliveryState: {
  delivery: Delivery | null
  open: boolean
} = {
  delivery: null,
  open: false,
}

export function TableHistory({ tableId, userId }: TableHistoryProps) {
  const { loading, history } = useTableHistory(tableId)
  const [order, setOrder] = useState(initialOrderState)
  const [payment, setPayment] = useState(initialPaymentState)
  const [cancelation, setCancelation] = useState(initialCancelationState)
  const [delivery, setDelivery] = useState(initialDeliveryState)

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : history.map((item) => {
              if (Object.prototype.hasOwnProperty.call(item, 'registeredAt')) {
                const payment = item as Payment
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Zahlung -${(payment.totalPaymentCents / 100).toFixed(2)} €`}
                    date={payment.registeredAt}
                    isFromUser={userId === payment.userId}
                    comment={payment.comment}
                    onClick={() => {
                      setPayment({ payment, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'placedAt')
              ) {
                const order = item as Order
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Bestellung +${(order.totalPriceCents / 100).toFixed(2)} €`}
                    date={order.placedAt}
                    isFromUser={userId === order.userId}
                    comment={order.comment}
                    onClick={() => {
                      setOrder({ order, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'canceledAt')
              ) {
                const cancelation = item as Cancelation
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Stornierung -${(cancelation.totalCancelationCents / 100).toFixed(2)} €`}
                    date={cancelation.canceledAt}
                    isFromUser={userId === cancelation.userId}
                    comment={cancelation.comment}
                    onClick={() => {
                      setCancelation({ cancelation, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'deliveredAt')
              ) {
                const delivery = item as Delivery
                return (
                  <HistoryItem
                    key={item.id}
                    title="Auslieferung"
                    date={delivery.deliveredAt}
                    isFromUser={userId === delivery.userId}
                    comment={delivery.comment}
                    onClick={() => {
                      setDelivery({ delivery, open: true })
                    }}
                  />
                )
              } else {
                return null
              }
            })}
      </ItemGroup>
      {order.order && (
        <Details
          title="Bestellung"
          id={order.order.id}
          isFromUser={userId === order.order.userId}
          open={order.open}
          onClose={() => {
            setOrder(initialOrderState)
          }}
          date={order.order.placedAt}
          comment={order.order.comment}
          products={order.order.products}
          totalPrice={order.order.totalPriceCents}
        />
      )}
      {payment.payment && (
        <Details
          title="Zahlung"
          id={payment.payment.id}
          isFromUser={userId === payment.payment.userId}
          open={payment.open}
          onClose={() => {
            setPayment(initialPaymentState)
          }}
          date={payment.payment.registeredAt}
          comment={payment.payment.comment}
          products={payment.payment.products}
          totalPrice={payment.payment.totalPaymentCents}
        />
      )}
      {cancelation.cancelation && (
        <Details
          title="Stornierung"
          id={cancelation.cancelation.id}
          isFromUser={userId === cancelation.cancelation.userId}
          open={cancelation.open}
          onClose={() => {
            setCancelation(initialCancelationState)
          }}
          date={cancelation.cancelation.canceledAt}
          comment={cancelation.cancelation.comment}
          products={cancelation.cancelation.products}
          totalPrice={cancelation.cancelation.totalCancelationCents}
        />
      )}
      {delivery.delivery && (
        <Details
          title="Auslieferung"
          id={delivery.delivery.id}
          isFromUser={userId === delivery.delivery.userId}
          open={delivery.open}
          onClose={() => {
            setDelivery(initialDeliveryState)
          }}
          date={delivery.delivery.deliveredAt}
          comment={delivery.delivery.comment}
          products={delivery.delivery.products}
        />
      )}
    </>
  )
}

function HistoryItem({
  title,
  date,
  isFromUser,
  comment,
  onClick,
}: {
  title: string
  date: string
  isFromUser: boolean
  comment: string
  onClick: () => void
}) {
  return (
    <Item variant="outline" className={isFromUser ? 'border-primary' : ''}>
      <ItemContent>
        <ItemTitle>{title}</ItemTitle>
        <ItemDescription>
          {new Date(date).toLocaleString()}
          {comment && (
            <>
              <br />
              {comment}
            </>
          )}
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

function Details({
  open,
  onClose,
  title,
  id,
  date,
  isFromUser,
  comment,
  products,
  totalPrice,
}: {
  open: boolean
  onClose: () => void
  title: string
  id: string
  date: string
  isFromUser: boolean
  comment: string
  products: OrderProduct[]
  totalPrice?: number
}) {
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
              {title} {id.slice(0, 8)}
            </DrawerTitle>
            <DrawerDescription>
              {isFromUser ? 'Du am ' : ''}
              {new Date(date).toLocaleDateString()} um{' '}
              {new Date(date).toLocaleTimeString()} Uhr
            </DrawerDescription>
          </DrawerHeader>
          <Receipt products={products} totalPrice={totalPrice} />
          {comment && (
            <div className="px-4">
              <Comment value={comment} />
            </div>
          )}
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
