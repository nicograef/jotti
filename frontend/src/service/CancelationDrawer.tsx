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
import type { Table } from './table/Table'
import type { TableBackend } from './table/TableBackend'

interface CancelationDrawerProps {
  backend: Pick<TableBackend, 'cancelTableProducts'>
  table: Table
  unpaidProducts: OrderProduct[]
  quantities: Record<number, number>
  productsCanceled: () => void
}

export function CancelationDrawer(props: CancelationDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const productsToCancel = buildOrderProducts(
    props.unpaidProducts,
    props.quantities,
  )
  const totalPrice = calculateTotalPrice(productsToCancel)
  const noProductsSelected = productsToCancel.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.cancelTableProducts({
        tableId: props.table.id,
        products: productsToCancel,
      })
      props.productsCanceled()
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
          variant="destructive"
          disabled={noProductsSelected}
          className="cursor-pointer hover:shadow-sm w-full"
        >
          Stornierung
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Stornierung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Sollen diese Produkte wirklich storniert werden?
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 space-y-2">
            {productsToCancel.map((product) => {
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
              variant="destructive"
              disabled={loading}
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : <></>} Produkte stornieren
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

function buildOrderProducts(
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

function calculateTotalPrice(cancelationProducts: OrderProduct[]): number {
  return cancelationProducts.reduce(
    (total, product) => total + product.netPriceCents * product.quantity,
    0,
  )
}
