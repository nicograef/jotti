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
import { getActionErrorMessage } from '@/lib/errorMessages'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import {
  calculateTotalPrice,
  selectPositionen,
  toPositionRefs,
  toReceiptItems,
} from './drawerUtils'
import { Receipt } from './Receipt'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  unbezahltePositionen: Position[]
  mengen: Record<string, number>
  zahlungKassiert: () => void
}

export function ZahlungDrawer(props: ZahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const positionenToPay = selectPositionen(
    props.unbezahltePositionen,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(positionenToPay)
  const noPositionenSelected = positionenToPay.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.zahlungKassieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(positionenToPay),
        kommentar,
      })
      props.zahlungKassiert()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Zahlung kassieren',
          error,
          byCode: {
            position_nicht_bezahlbar:
              'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Auswahl aktualisieren.',
          },
        }),
      )
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
          <Receipt
            positionen={toReceiptItems(positionenToPay)}
            totalPrice={totalPrice}
          />
          <div className="px-4">
            <KommentarField
              onChange={(value) => {
                setKommentar(value)
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
              {loading ? <Spinner /> : <></>} Kassieren
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
