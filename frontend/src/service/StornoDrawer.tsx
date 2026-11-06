import { type ReactNode, useState } from 'react'

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

import { PositionAuswahlListe } from './PositionAuswahlListe'
import type { PositionRef } from './schemas'
import { ActionHint } from './table/ActionHint'
import { KommentarField } from './table/CommentField'
import { GesamtZeile } from './table/DrawerParts'
import {
  type AuswaehlbarePosition,
  calculateTotalPrice,
  selectPositionen,
  toAuswahlPositionen,
  toPositionRefs,
} from './table/drawerUtils'

interface StornoDrawerProps {
  // DrawerHeader mit Titel und Beschreibung des Vorgangs.
  kopf: ReactNode
  positionen: AuswaehlbarePosition[]
  stornieren: (positionen: PositionRef[], kommentar: string) => Promise<void>
  onClose: () => void
  onStorniert: () => void
}

// Teil-Storno eines Vorgangs: Positionen wählen, Grund (mind. 3 Zeichen) nennen.
export function StornoDrawer({
  kopf,
  positionen,
  stornieren,
  onClose,
  onStorniert,
}: StornoDrawerProps) {
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
    onSuccess: onStorniert,
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
        {kopf}
        <DrawerBody className="mx-auto w-full max-w-sm">
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(positionen)}
            mengen={mengen}
            onAdd={add}
            onRemove={remove}
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
