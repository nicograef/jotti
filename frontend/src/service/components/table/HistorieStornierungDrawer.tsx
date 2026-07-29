import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
} from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useMengen } from '@/hooks/use-mengen'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'

import type { Bestellung } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import type { Umbuchung } from '../../table/Umbuchung'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import { ActionHint } from './ActionHint'
import { KommentarField } from './CommentField'
import { GesamtZeile, QuelleDrawerHeader } from './DrawerParts'
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
  quelle: Bestellung | Umbuchung
  onClose: () => void
  onStornierungErteilt: () => void
}

export function HistorieStornierungDrawer({
  backend,
  tisch,
  quelle,
  onClose,
  onStornierungErteilt,
}: HistorieStornierungDrawerProps) {
  const positionen = quelle.stornierbarePositionen
  const [kommentar, setKommentar] = useState('')
  const { mengen, add, remove } = useMengen<string>(
    (positionId) =>
      positionen.find((p) => p.positionId === positionId)?.menge ?? 0,
  )

  // Die Positionsauswahl meldet bereits useMengen; der getippte Grund kommt
  // hinzu, weil er ein zweites Mal formuliert werden müsste.
  useOffenerVorgang(kommentar.trim() !== '')

  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3
  // Grund am Button nur für die Positionsauswahl: Die Kommentar-Pflicht nennt
  // bereits das KommentarField dauerhaft, ein zweiter Hinweis wäre redundant.
  const hinweisGrund = noPositionenSelected ? 'Positionen auswählen' : null

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
      open
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DrawerContent pending={loading}>
        <QuelleDrawerHeader quelle={quelle}>
          Positionen aus diesem Vorgang zum Stornieren auswählen.
        </QuelleDrawerHeader>
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
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
          {!noPositionenSelected && (
            <GesamtZeile label="Stornierung gesamt" betrag={totalPrice} />
          )}
          <KommentarField
            required
            invalid={kommentarInvalid}
            onChange={(value) => {
              setKommentar(value)
            }}
          />
          <ActionHint reason={hinweisGrund} />
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
