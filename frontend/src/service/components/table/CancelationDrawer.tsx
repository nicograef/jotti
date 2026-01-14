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
  const [comment, setComment] = useState('')
  const productsToCancel = orderProducts(props.unpaidProducts, props.quantities)
  const totalPrice = calculateTotalPrice(productsToCancel)
  const noProductsSelected = productsToCancel.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.cancelTableProducts({
        tableId: props.table.id,
        products: productsToCancel,
        comment: comment,
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
          <Receipt products={productsToCancel} totalPrice={totalPrice} />
          <div className="px-4">
            <CommentField
              onChange={(value) => {
                setComment(value)
              }}
            />
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

function calculateTotalPrice(cancelationProducts: OrderProduct[]): number {
  return cancelationProducts.reduce(
    (total, product) => total + product.netPriceCents * product.quantity,
    0,
  )
}
