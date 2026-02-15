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

import type { OrderVariant } from '../../table/Order'
import type { Table } from '../../table/Table'
import type { TableBackend } from '../../table/TableBackend'
import { CommentField } from './CommentField'
import { Receipt } from './Receipt'

interface DeliveryDrawerProps {
  backend: Pick<TableBackend, 'deliverTableVariants'>
  table: Table
  undeliveredVariants: OrderVariant[]
  quantities: Record<number, number>
  variantsDelivered: () => void
}

export function DeliveryDrawer(props: DeliveryDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const variantsToDeliver = selectVariants(
    props.undeliveredVariants,
    props.quantities,
  )
  const noVariantsSelected = variantsToDeliver.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.deliverTableVariants({
        tableId: props.table.id,
        variants: variantsToDeliver,
        comment: comment,
      })
      props.variantsDelivered()
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
            disabled={noVariantsSelected}
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
          >
            Varianten liefern
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Lieferung für {props.table.name}</DrawerTitle>
            <DrawerDescription>
              Wurden diese Varianten an den Tisch ausgeliefert?
            </DrawerDescription>
          </DrawerHeader>
          <Receipt variants={variantsToDeliver} />
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
              {loading ? <Spinner /> : <></>} Varianten liefern
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
  variants: OrderVariant[],
  selectedQuantity: Record<number, number>,
): OrderVariant[] {
  return variants
    .map((variant) => ({
      ...variant,
      quantity: selectedQuantity[variant.id] || 0,
    }))
    .filter((variant) => variant.quantity > 0)
}
