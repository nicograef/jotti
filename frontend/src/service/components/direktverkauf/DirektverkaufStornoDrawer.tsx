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
} from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useMengen } from '@/hooks/use-mengen'
import { formatCents } from '@/lib/utils'

import type { DirektverkaufHistorieEintrag } from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import { KommentarField } from '../table/CommentField'
import { calculateTotalPrice, toAuswahlPositionen } from '../table/drawerUtils'

interface DirektverkaufStornoDrawerProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufStornieren'>
  verkauf: DirektverkaufHistorieEintrag
  onClose: () => void
  onStorniert: () => void
}

export function DirektverkaufStornoDrawer({
  backend,
  verkauf,
  onClose,
  onStorniert,
}: DirektverkaufStornoDrawerProps) {
  const [kommentar, setKommentar] = useState('')
  const { mengen, add, remove } = useMengen<string>(
    (positionId) =>
      verkauf.offenePositionen.find((p) => p.positionId === positionId)
        ?.menge ?? 0,
  )

  const selectedPositionen = verkauf.offenePositionen
    .map((position) => ({
      ...position,
      menge: mengen[position.positionId] || 0,
    }))
    .filter((position) => position.menge > 0)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3

  const { loading, run } = useActionSubmit({
    actionLabel: 'Stornierung ausführen',
    byCode: {
      position_nicht_stornierbar:
        'Mindestens eine Position ist nicht mehr stornierbar. Bitte Auswahl aktualisieren.',
      kasse_nicht_geoeffnet:
        'Es ist keine Kassensitzung geöffnet. Bitte zuerst die Kasse öffnen.',
      verkauf_not_found: 'Der Verkauf wurde nicht gefunden.',
    },
    onSuccess: () => {
      onStorniert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await backend.direktverkaufStornieren({
        verkaufId: verkauf.verkaufId,
        positionen: selectedPositionen.map((position) => ({
          positionId: position.positionId,
          menge: position.menge,
        })),
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
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Verkauf stornieren</DrawerTitle>
            <DrawerDescription>
              Positionen aus diesem Verkauf zum Stornieren auswählen.
            </DrawerDescription>
          </DrawerHeader>
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(verkauf.offenePositionen)}
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
          <DrawerFooter>
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
        </div>
      </DrawerContent>
    </Drawer>
  )
}
