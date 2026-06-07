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
import { useActionSubmit } from '@/hooks/use-action-submit'

import type { Produkt } from '../../product/Produkt'
import type { BestellPositionInput } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import { calculateTotalPrice } from './drawerUtils'
import { Receipt, type ReceiptPosition } from './Receipt'

interface BestellungDrawerProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  products: Produkt[]
  mengen: Record<number, number>
  bestellungAufgenommen: () => void
}

export function BestellungDrawer(props: BestellungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const { receiptItems, inputItems } = toBestellungData(
    props.products,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(receiptItems)
  const noPositionenSelected = inputItems.length === 0

  const { loading, run } = useActionSubmit({
    actionLabel: 'Bestellung aufnehmen',
    byCode: {
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      props.bestellungAufgenommen()
      setOpen(false)
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.bestellungAufnehmen({
        tischId: props.tisch.id,
        positionen: inputItems,
        kommentar,
      })
    })
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
          <Receipt positionen={receiptItems} totalPrice={totalPrice} />
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
              {loading ? <Spinner /> : null} Bestellung aufnehmen
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

function toBestellungData(
  products: Produkt[],
  ausgewaehlteMengen: Record<number, number>,
): { receiptItems: ReceiptPosition[]; inputItems: BestellPositionInput[] } {
  const items = products.flatMap((p) =>
    p.varianten
      .filter((v) => (ausgewaehlteMengen[v.id] || 0) > 0)
      .map((v) => ({
        produktId: p.id,
        varianteId: v.id,
        name: v.name,
        einzelpreis: v.preisCents,
        menge: ausgewaehlteMengen[v.id],
      })),
  )

  return {
    receiptItems: items.map((i) => ({
      name: i.name,
      einzelpreis: i.einzelpreis,
      menge: i.menge,
    })),
    inputItems: items.map((i) => ({
      produktId: i.produktId,
      varianteId: i.varianteId,
      menge: i.menge,
    })),
  }
}
