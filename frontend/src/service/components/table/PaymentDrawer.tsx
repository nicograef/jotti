import { useState } from 'react'
import { toast } from 'sonner'

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
import { calculateTotalPrice, selectVariants } from './drawerUtils'
import { Receipt } from './Receipt'

interface PaymentDrawerProps {
  backend: Pick<TableBackend, 'registerTablePayment'>
  table: Table
  unpaidVariants: LineItem[]
  quantities: Record<number, number>
  paymentRegistered: () => void
}

export function PaymentDrawer(props: PaymentDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const variantsToPay = selectVariants(props.unpaidVariants, props.quantities)
  const totalPrice = calculateTotalPrice(variantsToPay)
  const noVariantsSelected = variantsToPay.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.registerTablePayment({
        tableId: props.table.id,
        variants: variantsToPay,
        comment: comment,
      })
      props.paymentRegistered()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error('Aktion fehlgeschlagen')
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
          disabled={noVariantsSelected}
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
          <Receipt variants={variantsToPay} totalPrice={totalPrice} />
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
