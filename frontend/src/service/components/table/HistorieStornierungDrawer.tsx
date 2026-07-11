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
import { formatCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import { KommentarField } from './CommentField'
import {
  calculateTotalPrice,
  selectPositionen,
  toAuswahlPositionen,
  toPositionRefs,
} from './drawerUtils'

interface HistorieStornierungDrawerProps {
  backend: Pick<TischBackend, 'stornierungErteilen'>
  tisch: Tisch
  // Ursprungsvorgang (Bestellung oder Umbuchungs-Zugang), dessen Positionen storniert
  // werden; beschriftet den Drawer.
  vorgangId: string
  positionen: Position[]
  onClose: () => void
  onStornierungErteilt: () => void
}

export function HistorieStornierungDrawer({
  backend,
  tisch,
  vorgangId,
  positionen,
  onClose,
  onStornierungErteilt,
}: HistorieStornierungDrawerProps) {
  const [kommentar, setKommentar] = useState('')
  const { mengen, add, remove } = useMengen<string>(
    (positionId) =>
      positionen.find((p) => p.positionId === positionId)?.menge ?? 0,
  )

  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3

  const { loading, run } = useActionSubmit({
    actionLabel: 'Stornierung ausführen',
    byCode: {
      position_nicht_stornierbar:
        'Mindestens eine Position ist nicht mehr stornierbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      onStornierungErteilt()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await backend.stornierungErteilen({
        tischId: tisch.id,
        positionen: toPositionRefs(selectedPositionen),
        kommentar,
      })
    })
  }

  return (
    <Drawer
      open={true}
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>
            Stornierung aus Vorgang {vorgangId.slice(0, 8)}
          </DrawerTitle>
          <DrawerDescription>
            Positionen aus diesem Vorgang zum Stornieren auswählen.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(positionen)}
            mengen={mengen}
            onAdd={(id) => {
              add(id)
            }}
            onRemove={(id) => {
              remove(id)
            }}
          />
          {!noPositionenSelected && (
            <div className="flex justify-between font-bold px-4 pt-2 pb-2 border-t-2">
              <div>Stornierung gesamt</div>
              <div>{formatCents(totalPrice)}&nbsp;€</div>
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
