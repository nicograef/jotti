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
} from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useMengen } from '@/hooks/use-mengen'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'
import { formatEuro } from '@/lib/utils'

import type { VerkaufPosition } from './direktverkauf/Direktverkauf'
import { PositionAuswahlListe } from './PositionAuswahlListe'
import type { PositionRef } from './schemas'
import { KommentarField } from './table/CommentField'
import {
  calculateTotalPrice,
  selectPositionen,
  toAuswahlPositionen,
  toPositionRefs,
} from './table/drawerUtils'

interface VorgangStornoDrawerProps {
  // Der Titel lautet „<bezeichnung> stornieren“.
  bezeichnung: string
  offenePositionen: VerkaufPosition[]
  stornieren: (positionen: PositionRef[], kommentar: string) => Promise<void>
  onClose: () => void
  onStorniert: () => void
}

export function VorgangStornoDrawer({
  bezeichnung,
  offenePositionen,
  stornieren,
  onClose,
  onStorniert,
}: VorgangStornoDrawerProps) {
  const [kommentar, setKommentar] = useState('')
  const { mengen, add, remove } = useMengen<string>(
    (positionId) =>
      offenePositionen.find((p) => p.positionId === positionId)?.menge ?? 0,
  )

  // Die Positionsauswahl meldet bereits useMengen; der getippte Grund kommt
  // hinzu, weil er ein zweites Mal formuliert werden müsste.
  useOffenerVorgang(kommentar.trim() !== '')

  const selectedPositionen = selectPositionen(offenePositionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3

  const { loading, run } = useActionSubmit({
    actionLabel: 'Stornierung ausführen',
    onSuccess: () => {
      onStorniert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await stornieren(toPositionRefs(selectedPositionen), kommentar)
    })
  }

  return (
    <Drawer
      open
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>{bezeichnung} stornieren</DrawerTitle>
          <DrawerDescription>
            Die zu stornierenden Positionen auswählen.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(offenePositionen)}
            mengen={mengen}
            onAdd={add}
            onRemove={remove}
          />
          {!noPositionenSelected && (
            <div className="flex justify-between font-bold px-4 pt-2 pb-2 border-t-2">
              <div>Stornierung gesamt</div>
              <div>{formatEuro(totalPrice)}</div>
            </div>
          )}
          <div className="px-4">
            <KommentarField
              required
              invalid={kommentarInvalid}
              onChange={(value) => {
                setKommentar(value)
              }}
            />
          </div>
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
          <Button
            variant="destructive"
            disabled={loading || noPositionenSelected || kommentarInvalid}
            onClick={() => {
              void onSubmit()
            }}
          >
            {loading ? <Spinner /> : null} Stornierung erteilen
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
