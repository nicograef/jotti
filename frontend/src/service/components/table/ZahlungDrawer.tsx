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
import { DockActionSlot } from '../ServiceDock'
import { KommentarField } from './CommentField'
import { DockActionButton } from './DockActionButton'
import {
  calculateTotalPrice,
  calculateZahlungsbetraege,
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
  restNachZahlungCents: number
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
      <DockActionSlot>
        <div className="flex items-center justify-between gap-3 text-[13px] text-muted-foreground">
          <span>Nach dieser Zahlung noch offen</span>
          <span className="font-semibold tabular-nums text-foreground">
            {formatCents(props.restNachZahlungCents)}&nbsp;€
          </span>
        </div>
      </DockActionSlot>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Kassieren"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Zahlung für
          </p>
          <DrawerTitle className="text-[22px] font-semibold">
            {props.tisch.name}
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            Zahlung für {props.tisch.name}
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <Receipt
            positionen={toReceiptItems(positionenToPay)}
            totalPrice={totalPrice}
          />
          <div className="px-4 pt-3 flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="erhalten">Erhalten</Label>
              <EuroInput
                id="erhalten"
                value={erhaltenEuro}
                onValueChange={setErhaltenEuro}
                className="w-28"
              />
            </div>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="zielbetrag">Zahlbetrag inkl. Trinkgeld</Label>
              <EuroInput
                id="zielbetrag"
                value={zielbetragEuro}
                onValueChange={setZielbetragEuro}
                aria-describedby="zielbetrag-hinweis"
                className="w-28"
              />
            </div>
            <p
              id="zielbetrag-hinweis"
              className="text-xs text-muted-foreground"
            >
              Nur ausfüllen, wenn der Gast aufrundet: den vollen Betrag
              inklusive Trinkgeld eintragen, dann rechnet die Kasse das Rückgeld
              passend.
            </p>
            {rueckgeldCents !== null && (
              <div className="flex items-baseline justify-between pt-1">
                <div className="text-[15px] font-semibold">Rückgeld</div>
                <div className="text-xl font-bold tabular-nums">
                  {formatCents(rueckgeldCents)}&nbsp;€
                </div>
              </div>
            )}
            {trinkgeldCents !== null && (
              <>
                <div className="flex justify-between font-medium">
                  <div>Trinkgeld</div>
                  <div className="tabular-nums">
                    {formatCents(trinkgeldCents)}&nbsp;€
                  </div>
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
