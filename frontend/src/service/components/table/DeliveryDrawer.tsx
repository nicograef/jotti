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

interface DeliveryDrawerProps {
  backend: Pick<TableBackend, 'deliverTableProducts'>
  table: Table
  undeliveredProducts: OrderProduct[]
  quantities: Record<number, number>
  productsDelivered: () => void
}

export function DeliveryDrawer(props: DeliveryDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const productsToDeliver = buildOrderProducts(
    props.undeliveredProducts,
    props.quantities,
  )
  const noProductsSelected = productsToDeliver.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.deliverTableProducts({
        tableId: props.table.id,
        products: productsToDeliver,
        comment: comment,
      })
      props.productsDelivered()
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
            Produkte liefern
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Lieferung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Wurden diese Produkte an den Tisch ausgeliefert?
            </DrawerDescription>
          </DrawerHeader>
          <Receipt products={productsToDeliver} />
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
              {loading ? <Spinner /> : <></>} Produkte liefern
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
