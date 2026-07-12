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

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { KommentarField } from '../table/CommentField'
import { DockActionButton } from '../table/DockActionButton'
import { calculateZahlungsbetraege } from '../table/drawerUtils'
import type { ReceiptPosition } from '../table/Receipt'
import { Receipt } from '../table/Receipt'

interface VerkaufPositionInput {
  produktId: number
  varianteId: number
  menge: number
}

interface DirektverkaufDrawerProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  receiptItems: ReceiptPosition[]
  positionen: VerkaufPositionInput[]
  anzahl: number
  totalCents: number
  verkaufAbgeschlossen: () => void
}

export function DirektverkaufDrawer(props: DirektverkaufDrawerProps) {
  const [open, setOpen] = useState(false)
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [kommentar, setKommentar] = useState('')
  // verkaufId pro logischem Vorgang, nicht pro Retry.
  const [verkaufId, setVerkaufId] = useState(() => crypto.randomUUID())

  const { rueckgeldCents } = calculateZahlungsbetraege(
    props.totalCents,
    parseCents(erhaltenEuro),
    0,
  )
  const noPositionenSelected = props.positionen.length === 0

  const { loading, run } = useActionSubmit({
    actionLabel: 'Verkauf abschließen',
    byCode: {
      kasse_nicht_geoeffnet:
        'Es ist keine Kassensitzung geöffnet. Bitte zuerst die Kasse öffnen.',
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      setOpen(false)
      setErhaltenEuro('')
      setKommentar('')
      props.verkaufAbgeschlossen()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.direktverkaufTaetigen({
        verkaufId,
        positionen: props.positionen,
        kommentar,
      })
    })
  }

  const onOpenChange = (isOpen: boolean) => {
    const nextOpen = noPositionenSelected ? false : isOpen
    if (nextOpen) {
      setVerkaufId(crypto.randomUUID())
    }
    setOpen(nextOpen)
    if (!nextOpen) {
      setErhaltenEuro('')
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Kassieren"
          anzahl={props.anzahl}
          summeCents={props.totalCents}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>Verkauf abschließen</DrawerTitle>
          <DrawerDescription>
            Überprüfe den Verkauf vor dem Absenden.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <Receipt
            positionen={props.receiptItems}
            totalPrice={props.totalCents}
          />
          <div className="flex flex-col gap-2 px-4 pt-3">
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
            {loading ? <Spinner /> : null} Verkauf abschließen
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
