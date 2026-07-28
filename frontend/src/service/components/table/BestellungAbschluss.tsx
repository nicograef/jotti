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
  // Idempotenz-Schlüssel je Zusammenstellung: Jeder Wiederholversuch behält den
  // Schlüssel — auch mit inzwischen geänderter Auswahl, denn genau diese
  // Abweichung erkennt und meldet der Server. Er wird von TablePage vergeben,
  // wo auch der Korb liegt: Der ausgehängte Tab-Inhalt darf ihn nicht mit sich
  // nehmen, sonst würde ein Wiederholversuch zur zweiten Buchung.
  bestellungId: string
  bestellungAufgenommen: () => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`), nur mit der zuerst gesendeten Auswahl. Räumt
  // Korb und Tischzustand ab — siehe TablePage, wo beides liegt.
  vorgangBereitsGebucht: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Tisch-Bestellens (Beleg,
// Kommentar, Gesamt, „Bestellung aufnehmen"). Trägt den Kommentar und das
// Submit-/Fehler-/Retry-Verhalten und wird sowohl im Handy-Drawer als auch in
// der festen Spalte gerendert; Korb und bestellungId kommen von TablePage.
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

  const { loading, run } = useActionSubmit({
    actionLabel: 'Bestellung aufnehmen',
    byCode: {
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onCode: {
      // Der Vorgang ist gebucht, nur seine Antwort ging verloren — clientseitig
      // ist er damit abgeschlossen, auch wenn die zuletzt gesendete Auswahl eine
      // andere war. Ohne das Abräumen liefe jeder weitere Versuch wieder in
      // denselben 409, und die nachträglich ergänzte Position bliebe ungebucht.
      vorgang_daten_abweichend: () => {
        setKommentar('')
        props.vorgangBereitsGebucht()
      },
    },
    onSuccess: () => {
      setKommentar('')
      props.bestellungAufgenommen()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.bestellungAufnehmen({
        bestellungId: props.bestellungId,
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
