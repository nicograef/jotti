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

import type { OrderProduct } from './table/Order'
import type { PaymentProduct } from './table/Payment'
import type { Table } from './table/Table'
import type { TableBackend } from './table/TableBackend'

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
  const productsToPay = buildPaymentProducts(
    props.unpaidProducts,
    props.quantities,
  )
  const totalPrice = calculateTotalPrice(productsToPay)
  const noProductsSelected = productsToPay.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.registerTablePayment({
        tableId: props.table.id,
        products: productsToPay,
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
        <div className="text-center">
          <Button
            disabled={noProductsSelected}
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
          >
            Zahlung überprüfen
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Zahlung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Zahlung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 space-y-2">
            {productsToPay.map((product) => {
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
              <div>€ {(totalPrice / 100).toFixed(2)}</div>
            </div>
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

function buildPaymentProducts(
  products: OrderProduct[],
  selectedQuantity: Record<number, number>,
): PaymentProduct[] {
  return products
    .map((product) => ({
      ...product,
      quantity: selectedQuantity[product.id] || 0,
    }))
    .filter((product) => product.quantity > 0)
}

function calculateTotalPrice(paymentProducts: PaymentProduct[]): number {
  return paymentProducts.reduce(
    (total, product) => total + product.netPriceCents * product.quantity,
    0,
  )
}
