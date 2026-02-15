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

import type { Product, Variant } from '../../product/Product'
import type { LineItem } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { CommentField } from './CommentField'
import { Receipt } from './Receipt'

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
  const [comment, setComment] = useState('')
  const orderedVariants = lineItems(props.products, props.quantities)
  const totalPrice = calculateTotalPrice(orderedVariants)
  const noVariantsSelected = orderedVariants.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.placeTableOrder({
        tableId: props.table.id,
        variants: orderedVariants,
        comment: comment,
      })
      props.orderPlaced()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  const onOpenChange = (isOpen: boolean) => {
    if (noVariantsSelected) {
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
            disabled={noVariantsSelected}
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
          <Receipt variants={orderedVariants} totalPrice={totalPrice} />
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

function lineItems(
  products: Product[],
  selectedQuantity: Record<number, number>,
): LineItem[] {
  const allVariants: Variant[] = products.flatMap((p) => p.variants)
  return allVariants
    .map((variant) => ({
      id: variant.id,
      name: variant.name,
      priceCents: variant.priceCents,
      quantity: selectedQuantity[variant.id] || 0,
    }))
    .filter((variant) => variant.quantity > 0)
}

function calculateTotalPrice(lineItems: LineItem[]): number {
  return lineItems.reduce(
    (total, variant) => total + variant.priceCents * variant.quantity,
    0,
  )
}
