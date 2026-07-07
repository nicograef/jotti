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
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import { calculateTotalPrice, toBestellungData } from './drawerUtils'
import { Receipt } from './Receipt'
import { StickyActionBar } from './StickyActionBar'

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
  // bestellungId pro logischem Vorgang (nicht pro Retry). Neue ID wenn Drawer öffnet.
  const [bestellungId, setBestellungId] = useState(() => crypto.randomUUID())
  const { receiptItems, inputItems } = toBestellungData(
    props.products,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(receiptItems)
  const anzahl = inputItems.reduce((sum, item) => sum + item.menge, 0)
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
        bestellungId,
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
      if (isOpen) {
        setBestellungId(crypto.randomUUID())
      }
      setOpen(isOpen)
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <StickyActionBar
          label="Bestellung überprüfen"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
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
