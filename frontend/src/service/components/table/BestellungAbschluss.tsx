import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
} from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useVorgangId } from '@/hooks/use-vorgang-id'

import type { BestellPositionInput } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import { KommentarField } from './CommentField'
import { GesamtZeile } from './DrawerParts'
import type { ReceiptPosition } from './Receipt'
import { Receipt } from './Receipt'

interface BestellungAbschlussProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  receiptItems: ReceiptPosition[]
  positionen: BestellPositionInput[]
  totalCents: number
  bestellungAufgenommen: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Tisch-Bestellens (Beleg,
// Kommentar, Gesamt, „Bestellung aufnehmen"). Trägt den vollständigen Zustand
// samt bestellungId-Lebenszyklus und Submit-/Fehler-/Retry-Verhalten und wird
// sowohl im Handy-Drawer als auch in der festen Spalte gerendert.
export function BestellungAbschluss(props: BestellungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')

  const noPositionenSelected = props.positionen.length === 0

  // Beginnt eine Zusammenstellung aus dem Leerzustand, startet der Kommentar
  // leer: In der dauerhaften Spalte überlebt er sonst einen Auswahl-Reset und
  // würde aus einem abgebrochenen Vorgang in den nächsten wandern.
  // React-idiomatischer State-Sync im Render (wie beim Tischwechsel in
  // TablePage), nicht per Effekt.
  const [warLeer, setWarLeer] = useState(noPositionenSelected)
  if (warLeer !== noPositionenSelected) {
    setWarLeer(noPositionenSelected)
    if (warLeer) {
      setKommentar('')
    }
  }

  // bestellungId je fachlichem Vorgang, an die Nutzdaten gebunden: Ein
  // Wiederholversuch mit unveränderten Nutzdaten behält seinen Schlüssel und
  // bucht serverseitig kein zweites Mal; jede Änderung (Positionen, Kommentar)
  // beginnt einen neuen Vorgang mit neuem Schlüssel — auch nach einem
  // erfolgreichen Abschluss, der die Auswahl leert.
  const bestellungId = useVorgangId({
    tischId: props.tisch.id,
    positionen: props.positionen,
    kommentar,
  })

  const { loading, run } = useActionSubmit({
    actionLabel: 'Bestellung aufnehmen',
    byCode: {
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      setKommentar('')
      props.bestellungAufgenommen()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.bestellungAufnehmen({
        bestellungId,
        tischId: props.tisch.id,
        positionen: props.positionen,
        kommentar,
      })
    })
  }

  const inhalt = (
    <>
      <AbschlussHeader
        variant={props.variant}
        eyebrow="Bestellung für"
        title={props.tisch.name}
        description={`Bestellung für ${props.tisch.name}`}
      />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          <AbschlussLeer>Produkte auswählen, um zu bestellen.</AbschlussLeer>
        ) : (
          <>
            <Receipt positionen={props.receiptItems} />
            <div className="px-4">
              <KommentarField
                value={kommentar}
                onChange={(value) => {
                  setKommentar(value)
                }}
              />
            </div>
          </>
        )}
      </DrawerBody>
      <DrawerFooter className="mx-auto w-full max-w-sm">
        <GesamtZeile label="Gesamt" betrag={props.totalCents} />
        <Button
          disabled={loading || noPositionenSelected}
          onClick={() => {
            void onSubmit()
          }}
        >
          {loading ? <Spinner /> : null} Bestellung aufnehmen
        </Button>
        {props.variant === 'sheet' && (
          <DrawerClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DrawerClose>
        )}
      </DrawerFooter>
    </>
  )

  if (props.variant === 'sheet') {
    return <DrawerContent pending={loading}>{inhalt}</DrawerContent>
  }

  // Feste Spalte: dieselben Body/Footer-Primitive wie im Sheet, nur in einem
  // eigenen, unabhängig scrollenden Container. group/drawer-content +
  // data-pending übernehmen das Body-Dimming des Drawers während des Submits.
  return (
    <aside
      data-pending={loading || undefined}
      className="group/drawer-content flex min-h-0 flex-col overflow-hidden rounded-xl border bg-popover text-sm text-popover-foreground"
    >
      {inhalt}
    </aside>
  )
}
