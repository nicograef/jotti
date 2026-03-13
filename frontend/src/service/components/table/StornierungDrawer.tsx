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

interface StornierungDrawerProps {
  backend: Pick<TischBackend, 'produkteStornieren'>
  tisch: Tisch
  unbezahltePositionen: Position[]
  mengen: Record<string, number>
  produkteStorniert: () => void
}

export function StornierungDrawer(props: StornierungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const selectedPositionen = selectPositionen(
    props.unbezahltePositionen,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.produkteStornieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(selectedPositionen),
        kommentar,
      })
      props.produkteStorniert()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Stornierung ausführen',
          error,
          byCode: {
            position_nicht_stornierbar:
              'Mindestens eine Position ist nicht mehr stornierbar. Bitte Auswahl aktualisieren.',
          },
        }),
      )
    }

    setLoading(false)
  }

  const onOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) {
      setKommentar('')
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <Button
          variant="destructive"
          disabled={noPositionenSelected}
          className="cursor-pointer hover:shadow-sm w-full"
        >
          Stornierung
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Stornierung für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Positionen prüfen und stornieren.
            </DrawerDescription>
          </DrawerHeader>
          <Receipt
            positionen={toReceiptItems(selectedPositionen)}
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
              variant="destructive"
              disabled={loading || noPositionenSelected}
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
