import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { AbschlussContainer } from './AbschlussContainer'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import type { BestellPositionInput } from './Bestellung'
import { KommentarField } from './CommentField'
import { GesamtZeile } from './DrawerParts'
import type { ReceiptPosition } from './Receipt'
import { Receipt } from './Receipt'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'

interface BestellungAbschlussProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  receiptItems: ReceiptPosition[]
  positionen: BestellPositionInput[]
  totalCents: number
  // Liefert die bestellungId für diesen Versuch (siehe Vorgang.schluesselFuer).
  schluesselFuer: (inhalt: unknown) => string
  bestellungAufgenommen: () => void
  variant: 'sheet' | 'spalte'
}

export function BestellungAbschluss(props: BestellungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')

  const noPositionenSelected = props.positionen.length === 0

  // A selection starting from empty is a new Vorgang; the persistent column carries no Kommentar over from an abandoned one.
  const warLeerRef = useRef(noPositionenSelected)
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      setKommentar('')
    }
    warLeerRef.current = noPositionenSelected
  }, [noPositionenSelected])

  const { loading, run } = useActionSubmit({
    actionLabel: 'Bestellung aufnehmen',
    onSuccess: () => {
      setKommentar('')
      props.bestellungAufgenommen()
    },
  })

  const onSubmit = async () => {
    const bestellungId = props.schluesselFuer({
      positionen: props.positionen,
      kommentar,
    })
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

  return (
    <AbschlussContainer variant={props.variant} pending={loading}>
      {inhalt}
    </AbschlussContainer>
  )
}
