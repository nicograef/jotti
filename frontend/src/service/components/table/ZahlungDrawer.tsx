import { useState } from 'react'

import { EuroInput } from '@/components/common/EuroInput'
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
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, parseCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import {
  calculateTotalPrice,
  calculateZahlungsbetraege,
  selectPositionen,
  toPositionRefs,
  toReceiptItems,
} from './drawerUtils'
import { Receipt } from './Receipt'
import { StickyActionBar } from './StickyActionBar'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  unbezahltePositionen: Position[]
  mengen: Record<string, number>
  zahlungKassiert: () => void
}

export function ZahlungDrawer(props: ZahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const [kommentar, setKommentar] = useState('')
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const positionenToPay = selectPositionen(
    props.unbezahltePositionen,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(positionenToPay)
  const anzahl = positionenToPay.reduce((sum, p) => sum + p.menge, 0)
  const { rueckgeldCents, trinkgeldCents } = calculateZahlungsbetraege(
    totalPrice,
    parseCents(erhaltenEuro),
    parseCents(zielbetragEuro),
  )
  const noPositionenSelected = positionenToPay.length === 0

  const { loading, run } = useActionSubmit({
    actionLabel: 'Zahlung kassieren',
    byCode: {
      position_nicht_bezahlbar:
        'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      props.zahlungKassiert()
      setOpen(false)
      setErhaltenEuro('')
      setZielbetragEuro('')
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.zahlungKassieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(positionenToPay),
        kommentar,
      })
    })
  }

  const onOpenChange = (isOpen: boolean) => {
    const nextOpen = noPositionenSelected ? false : isOpen
    setOpen(nextOpen)
    if (!nextOpen) {
      setErhaltenEuro('')
      setZielbetragEuro('')
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <StickyActionBar
          label="Kassieren"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>Zahlung für {props.tisch.name}</DrawerTitle>
          <DrawerDescription>
            Überprüfe deine Zahlung vor dem Absenden.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <Receipt
            positionen={toReceiptItems(positionenToPay)}
            totalPrice={totalPrice}
          />
          <div className="px-4 pt-3 flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="zielbetrag">inklusive Trinkgeld</Label>
              <EuroInput
                id="zielbetrag"
                value={zielbetragEuro}
                onValueChange={setZielbetragEuro}
                className="w-28"
              />
            </div>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="erhalten">Erhalten</Label>
              <EuroInput
                id="erhalten"
                value={erhaltenEuro}
                onValueChange={setErhaltenEuro}
                className="w-28"
              />
            </div>
            {rueckgeldCents !== null && (
              <div className="flex justify-between font-medium">
                <div>Rückgeld</div>
                <div>{formatCents(rueckgeldCents)}&nbsp;€</div>
              </div>
            )}
            {trinkgeldCents !== null && (
              <>
                <div className="flex justify-between font-medium">
                  <div>Trinkgeld</div>
                  <div>{formatCents(trinkgeldCents)}&nbsp;€</div>
                </div>
                <p className="text-xs text-muted-foreground">
                  Trinkgeld wird nicht als Kasseneinnahme gebucht und gehört
                  nicht in die Kassenlade.
                </p>
              </>
            )}
          </div>
          <div className="px-4 pt-3">
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
            {loading ? <Spinner /> : null} Kassieren
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
