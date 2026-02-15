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

import type { LineItem } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { CommentField } from './CommentField'
import { Receipt } from './Receipt'

interface CancelationDrawerProps {
  backend: Pick<TableBackend, 'cancelTableVariants'>
  table: Table
  unpaidVariants: LineItem[]
  quantities: Record<number, number>
  variantsCanceled: () => void
}

export function CancelationDrawer(props: CancelationDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const variantsToCancel = selectVariants(
    props.unpaidVariants,
    props.quantities,
  )
  const totalPrice = calculateTotalPrice(variantsToCancel)
  const noVariantsSelected = variantsToCancel.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.cancelTableVariants({
        tableId: props.table.id,
        variants: variantsToCancel,
        comment: comment,
      })
      props.variantsCanceled()
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
        <Button
          variant="destructive"
          disabled={noVariantsSelected}
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
              Sollen diese Varianten wirklich storniert werden?
            </DrawerDescription>
          </DrawerHeader>
          <Receipt variants={variantsToCancel} totalPrice={totalPrice} />
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
              {loading ? <Spinner /> : <></>} Varianten stornieren
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

function selectVariants(
  variants: LineItem[],
  selectedQuantity: Record<number, number>,
): LineItem[] {
  return variants
    .map((variant) => ({
      ...variant,
      quantity: selectedQuantity[variant.id] || 0,
    }))
    .filter((variant) => variant.quantity > 0)
}

function calculateTotalPrice(variants: LineItem[]): number {
  return variants.reduce(
    (total, variant) => total + variant.priceCents * variant.quantity,
    0,
  )
}
