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
import type { OrderBackend } from '@/lib/order/OrderBackend'
import type { ProductPublic } from '@/lib/product/Product'
import type { TablePublic } from '@/lib/table/Table'

type OrderProduct = ProductPublic & { quantity: number }

interface OrderDrawerProps {
  backend: Pick<OrderBackend, 'placeOrder'>
  open: boolean
  table: TablePublic
  products: ProductPublic[]
  quantities: Record<number, number>
  cancel: () => void
  orderPlaced: () => void
}

export function OrderDrawer(props: OrderDrawerProps) {
  const [loading, setLoading] = useState(false)
  const orderedProducts = orderProducts(props.products, props.quantities)
  const totalPrice = calculateTotalPrice(orderedProducts)

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      props.cancel()
    }
  }

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.placeOrder({
        tableId: props.table.id,
        products: orderedProducts,
      })
      props.orderPlaced()
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Drawer open={props.open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <div className="text-center">
          <Button className="cursor-pointer hover:shadow-sm w-full lg:w-1/2">
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
          <div className="my-4 space-y-2">
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
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : <></>} Jetzt Bestellen
            </Button>
            <DrawerClose asChild>
              <Button variant="outline">Abbrechen</Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}

function orderProducts(
  products: ProductPublic[],
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
