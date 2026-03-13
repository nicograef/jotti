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
import { KommentarField } from './CommentField'
import { selectPositionen } from './drawerUtils'
import { Receipt } from './Receipt'

interface LieferungDrawerProps {
  backend: Pick<TischBackend, 'produkteLiefern'>
  tisch: Tisch
  ungeliefertePositionen: Position[]
  mengen: Record<number, number>
  produkteGeliefert: () => void
}

export function LieferungDrawer(props: LieferungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const positionenToDeliver = selectPositionen(
    props.ungeliefertePositionen,
    props.mengen,
  )
  const noPositionenSelected = positionenToDeliver.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.produkteLiefern({
        tischId: props.tisch.id,
        positionen: positionenToDeliver,
        kommentar,
      })
      props.produkteGeliefert()
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
        <div className="text-center">
          <Button
            disabled={noPositionenSelected}
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
          >
            Produkte liefern
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Lieferung für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Wurden diese Produkte an den Tisch ausgeliefert?
            </DrawerDescription>
          </DrawerHeader>
          <Receipt positionen={positionenToDeliver} />
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
