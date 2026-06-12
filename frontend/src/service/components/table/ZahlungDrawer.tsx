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
import { Input } from '@/components/ui/input'
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
          <div className="px-4 pt-3 flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="zielbetrag">inklusive Trinkgeld</Label>
              <div className="flex items-center gap-1.5">
                <Input
                  id="zielbetrag"
                  inputMode="decimal"
                  placeholder="0,00"
                  value={zielbetragEuro}
                  onChange={(e) => {
                    setZielbetragEuro(e.target.value)
                  }}
                  className="w-24 text-right"
                  spellCheck={false}
                />
                <span>€</span>
              </div>
            </div>
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="erhalten">Erhalten</Label>
              <div className="flex items-center gap-1.5">
                <Input
                  id="erhalten"
                  inputMode="decimal"
                  placeholder="0,00"
                  value={erhaltenEuro}
                  onChange={(e) => {
                    setErhaltenEuro(e.target.value)
                  }}
                  className="w-24 text-right"
                  spellCheck={false}
                />
                <span>€</span>
              </div>
            </div>
            {rueckgeldCents !== null && (
              <div className="flex justify-between font-medium">
                <div>Rückgeld</div>
                <div>{formatCents(rueckgeldCents)}&nbsp;€</div>
              </div>
            )}
            {trinkgeldCents !== null && (
              <div className="flex justify-between font-medium">
                <div>Trinkgeld</div>
                <div>{formatCents(trinkgeldCents)}&nbsp;€</div>
              </div>
            )}
          </div>
          <div className="px-4 pt-3">
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
              {loading ? <Spinner /> : null} Kassieren
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
