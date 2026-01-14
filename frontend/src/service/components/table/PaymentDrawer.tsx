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
  DrawerTrigger,
} from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'

import type { OrderProduct } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { CommentField } from './CommentField'
import { Receipt } from './Receipt'

interface PaymentDrawerProps {
  backend: Pick<TableBackend, 'registerTablePayment'>
  table: Table
  unpaidProducts: OrderProduct[]
  quantities: Record<number, number>
  paymentRegistered: () => void
}

export function PaymentDrawer(props: PaymentDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const productsToPay = orderProducts(props.unpaidProducts, props.quantities)
  const totalPrice = calculateTotalPrice(productsToPay)
  const noProductsSelected = productsToPay.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.registerTablePayment({
        tableId: props.table.id,
        products: productsToPay,
        comment: comment,
      })
      props.paymentRegistered()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  const onOpenChange = (isOpen: boolean) => {
    if (noProductsSelected) {
      setOpen(false)
    } else {
      setOpen(isOpen)
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <Button
          disabled={noProductsSelected}
          className="cursor-pointer hover:shadow-sm w-full"
        >
          Zahlung
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Zahlung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Zahlung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <Receipt products={productsToPay} totalPrice={totalPrice} />
          <div className="px-4">
            <CommentField
              onChange={(value) => {
                setComment(value)
              }}
            />
          </div>
          <DrawerFooter>
            <Button
              disabled={loading}
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : <></>} Zahlung registrieren
            </Button>
            <DrawerClose asChild>
              <Button variant="outline" disabled={loading}>
                Abbrechen
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}

function orderProducts(
  products: OrderProduct[],
  selectedQuantity: Record<number, number>,
): OrderProduct[] {
  return products
    .map((product) => ({
      ...product,
      quantity: selectedQuantity[product.id] || 0,
    }))
    .filter((product) => product.quantity > 0)
}

function calculateTotalPrice(paymentProducts: OrderProduct[]): number {
  return paymentProducts.reduce(
    (total, product) => total + product.netPriceCents * product.quantity,
    0,
  )
}
