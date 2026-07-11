import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
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
import { DockActionButton } from './DockActionButton'
import { calculateTotalPrice, toBestellungData } from './drawerUtils'
import { Receipt } from './Receipt'

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
        <DockActionButton
          label="Bestellung überprüfen"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>Bestellung für {props.tisch.name}</DrawerTitle>
          <DrawerDescription>
            Überprüfe deine Bestellung vor dem Absenden.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <Receipt positionen={receiptItems} totalPrice={totalPrice} />
          <div className="px-4">
            <KommentarField
              onChange={(value) => {
                setKommentar(value)
              }}
            />
          </div>
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
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
      </DrawerContent>
    </Drawer>
  )
}
