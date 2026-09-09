import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { parseCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AbschlussContainer } from './AbschlussContainer'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import { BarzahlungFelder } from './BarzahlungFelder'
import { KommentarField } from './CommentField'
import {
  calculateZahlungsbetraege,
  toPositionRefs,
  toReceiptItems,
} from './drawerUtils'
import { Receipt } from './Receipt'
import { RestbetragZeile } from './RestbetragZeile'

interface ZahlungAbschlussProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  // Die ausgewählten, zu kassierenden Positionen (mit Auswahl-Menge).
  positionenToPay: Position[]
  totalCents: number
  restNachZahlungCents: number
  zahlungKassiert: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Tisch-Kassierens (Beleg, Erhalten,
// Zielbetrag inkl. Trinkgeld, Rückgeld, Trinkgeld-Hinweis, Kommentar,
// „Kassieren"). Trägt den vollständigen Eingabe-State und das Submit-/Fehler-/
// Retry-Verhalten und wird sowohl im Handy-Drawer als auch in der festen Spalte
// gerendert. Kein Client-Idempotenz-Schlüssel: die Idempotenz ist zustandsbasiert
// (bereits bezahlte Positionen → position_nicht_bezahlbar), und der
// Loading-Guard verhindert den Doppel-Submit.
export function ZahlungAbschluss(props: ZahlungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)

  const noPositionenSelected = props.positionenToPay.length === 0

  // In der dauerhaften Spalte überlebt der Eingabe-State sonst über einen
  // Auswahl-Reset hinweg. Beim Beginn einer neuen Zusammenstellung (aus dem
  // Leerzustand) starten die Eingaben deshalb leer, damit nichts aus einer
  // abgebrochenen Zahlung übertragen wird.
  const warLeerRef = useRef(noPositionenSelected)
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
    }
    warLeerRef.current = noPositionenSelected
  }, [noPositionenSelected])

  const { rueckgeldCents, trinkgeldCents } = calculateZahlungsbetraege(
    props.totalCents,
    parseCents(erhaltenEuro),
    parseCents(zielbetragEuro),
  )

  const { loading, run } = useActionSubmit({
    actionLabel: 'Zahlung kassieren',
    onSuccess: () => {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
      props.zahlungKassiert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.zahlungKassieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(props.positionenToPay),
        kommentar,
      })
    })
  }

  const inhalt = (
    <>
      <AbschlussHeader
        variant={props.variant}
        eyebrow="Zahlung für"
        title={props.tisch.name}
        description={`Zahlung für ${props.tisch.name}`}
      />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          <AbschlussLeer>Positionen auswählen, um zu kassieren.</AbschlussLeer>
        ) : (
          <>
            <Receipt
              positionen={toReceiptItems(props.positionenToPay)}
              totalPrice={props.totalCents}
            />
            <BarzahlungFelder
              gesamtCents={props.totalCents}
              erhaltenEuro={erhaltenEuro}
              onErhaltenEuroChange={setErhaltenEuro}
              zielbetragEuro={zielbetragEuro}
              onZielbetragEuroChange={setZielbetragEuro}
              andererAktiv={andererAktiv}
              onAndererAktivChange={setAndererAktiv}
              rueckgeldCents={rueckgeldCents}
              trinkgeldCents={trinkgeldCents}
            />
            <div className="px-4 pt-3">
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
        {props.variant === 'spalte' && (
          <RestbetragZeile cents={props.restNachZahlungCents} />
        )}
        <Button
          disabled={loading || noPositionenSelected}
          onClick={() => {
            void onSubmit()
          }}
        >
          {loading ? <Spinner /> : null} Kassieren
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
