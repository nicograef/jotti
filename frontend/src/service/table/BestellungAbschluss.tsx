import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import type { Vorgang } from '@/hooks/use-vorgang'

import { AbschlussContainer } from './AbschlussContainer'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import type { BestellPositionInput } from './Bestellung'
import { KommentarField } from './CommentField'
import { GesamtZeile } from './DrawerParts'
import { tischBackend } from './hooks'
import type { ReceiptPosition } from './Receipt'
import { Receipt } from './Receipt'
import type { Tisch } from './Tisch'

interface BestellungAbschlussProps {
  tisch: Tisch
  receiptItems: ReceiptPosition[]
  positionen: BestellPositionInput[]
  totalCents: number
  // Liefert Kommentar und bestellungId; beide überstehen das Aushängen des Abschlusses.
  vorgang: Vorgang<number>
  bestellungAufgenommen: () => void
  variant: 'sheet' | 'spalte'
}

export function BestellungAbschluss(props: BestellungAbschlussProps) {
  const { kommentar, setKommentar } = props.vorgang
  const noPositionenSelected = props.positionen.length === 0

  const { loading, run } = useActionSubmit({
    actionLabel: 'Bestellung aufnehmen',
    onSuccess: props.bestellungAufgenommen,
  })

  const onSubmit = async () => {
    const bestellungId = props.vorgang.schluesselFuer({
      positionen: props.positionen,
      kommentar,
    })
    await run(async () => {
      await tischBackend.bestellungAufnehmen({
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
              <KommentarField value={kommentar} onChange={setKommentar} />
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

  return (
    <AbschlussContainer variant={props.variant} pending={loading}>
      {inhalt}
    </AbschlussContainer>
  )
}
