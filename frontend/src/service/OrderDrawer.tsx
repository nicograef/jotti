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

import type { Product } from './product/Product'
import type { OrderProduct } from './table/Order'
import type { Table } from './table/Table'
import type { TableBackend } from './table/TableBackend'

interface OrderDrawerProps {
  backend: Pick<TableBackend, 'placeTableOrder'>
  table: Table
  products: Product[]
  quantities: Record<number, number>
  orderPlaced: () => void
}

export function OrderDrawer(props: OrderDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const orderedProducts = orderProducts(props.products, props.quantities)
  const totalPrice = calculateTotalPrice(orderedProducts)
  const noProductsSelected = orderedProducts.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.placeTableOrder({
        tableId: props.table.id,
        products: orderedProducts,
      })
      props.orderPlaced()
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
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
            disabled={noProductsSelected}
          >
            Bestellung überprüfen
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Bestellung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Bestellung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <div className="p-4 space-y-2">
            {orderedProducts.map((product) => {
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
              {loading ? <Spinner /> : <></>} Bestellung aufgeben
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
  products: Product[],
  selectedQuantity: Record<number, number>,
): OrderProduct[] {
  return products
    .map((product) => ({
      ...product,
      quantity: selectedQuantity[product.id] || 0,
    }))
    .filter((product) => product.quantity > 0)
}

function calculateTotalPrice(orderProducts: OrderProduct[]): number {
  return orderProducts.reduce(
    (total, product) => total + product.netPriceCents * product.quantity,
    0,
  )
}
