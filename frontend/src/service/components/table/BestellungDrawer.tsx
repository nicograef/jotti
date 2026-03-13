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

import type { Produkt, Variante } from '../../product/Product'
import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import { calculateTotalPrice } from './drawerUtils'
import { Receipt } from './Receipt'

interface BestellungDrawerProps {
  backend: Pick<TischBackend, 'bestellungAufgeben'>
  tisch: Tisch
  products: Produkt[]
  mengen: Record<number, number>
  bestellungAufgegeben: () => void
}

export function BestellungDrawer(props: BestellungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const orderedPositionen = toPositionen(props.products, props.mengen)
  const totalPrice = calculateTotalPrice(orderedPositionen)
  const noPositionenSelected = orderedPositionen.length === 0

  const onSubmit = async () => {
    setLoading(true)

    try {
      await props.backend.bestellungAufgeben({
        tischId: props.tisch.id,
        positionen: orderedPositionen,
        kommentar,
      })
      props.bestellungAufgegeben()
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
            className="cursor-pointer hover:shadow-sm w-full lg:w-1/2"
            disabled={noPositionenSelected}
          >
            Bestellung überprüfen
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Bestellung für {props.tisch.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Bestellung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <Receipt positionen={orderedPositionen} totalPrice={totalPrice} />
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

function toPositionen(
  products: Produkt[],
  ausgewaehlteMengen: Record<number, number>,
): Position[] {
  const alleVarianten: Variante[] = products.flatMap((p) => p.varianten)
  return alleVarianten
    .map((variante) => ({
      id: variante.id,
      name: variante.name,
      preisCents: variante.preisCents,
      menge: ausgewaehlteMengen[variante.id] || 0,
    }))
    .filter((variante) => variante.menge > 0)
}
