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
import { selectPositionen, toPositionRefs, toReceiptItems } from './drawerUtils'
import { Receipt } from './Receipt'

interface AusgabeDrawerProps {
  backend: Pick<TischBackend, 'ausgabeBestaetigen'>
  tisch: Tisch
  ausstehendePositionen: Position[]
  mengen: Record<string, number>
  ausgabeBestaetigt: () => void
}

export function AusgabeDrawer(props: AusgabeDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const positionenToDeliver = selectPositionen(
    props.ausstehendePositionen,
    props.mengen,
  )
  const noPositionenSelected = positionenToDeliver.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.ausgabeBestaetigen({
        tischId: props.tisch.id,
        positionen: toPositionRefs(positionenToDeliver),
        kommentar,
      })
      props.ausgabeBestaetigt()
      setOpen(false)
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Ausgabe bestätigen',
          error,
          byCode: {
            position_nicht_ausgebbar:
              'Mindestens eine Position ist nicht mehr ausgebbar. Bitte Auswahl aktualisieren.',
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
        <div className="text-center">
          <Button
            disabled={noPositionenSelected}
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
          >
            Ausgabe bestätigen
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Ausgabe für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Wurden diese Produkte an den Tisch übergeben?
            </DrawerDescription>
          </DrawerHeader>
          <Receipt positionen={toReceiptItems(positionenToDeliver)} />
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
              {loading ? <Spinner /> : <></>} Ausgabe bestätigen
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
