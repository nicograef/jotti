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

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { CommentField } from './CommentField'
import { calculateTotalPrice, selectPositionen } from './drawerUtils'
import { Receipt } from './Receipt'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungRegistrieren'>
  tisch: Tisch
  unbezahltePositionen: Position[]
  quantities: Record<number, number>
  zahlungRegistriert: () => void
}

export function ZahlungDrawer(props: ZahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [comment, setComment] = useState('')
  const positionenToPay = selectPositionen(
    props.unbezahltePositionen,
    props.quantities,
  )
  const totalPrice = calculateTotalPrice(positionenToPay)
  const noPositionenSelected = positionenToPay.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.zahlungRegistrieren({
        tischId: props.tisch.id,
        positionen: positionenToPay,
        kommentar: comment,
      })
      props.zahlungRegistriert()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error('Aktion fehlgeschlagen')
    }

    setLoading(false)
  }

  const onOpenChange = (isOpen: boolean) => {
    if (noPositionenSelected) {
      setOpen(false)
    } else {
      setOpen(isOpen)
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <Button
          disabled={noPositionenSelected}
          className="cursor-pointer hover:shadow-sm w-full"
        >
          Zahlung
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Zahlung für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Zahlung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <Receipt positionen={positionenToPay} totalPrice={totalPrice} />
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
