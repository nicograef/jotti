import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import type { Produkt } from '@/lib/produktSchemas'
import { parseCents } from '@/lib/utils'

import { PfandAbschnitt } from '../pfand/PfandAbschnitt'
import type { PfandBackend } from '../pfand/PfandBackend'
import { PfandVerrechnung } from '../pfand/PfandVerrechnung'
import {
  istPfandrueckgabeAbweichend,
  pfandAuswahl,
  type PfandVorgang,
} from '../pfand/usePfandVorgang'
import { AbschlussContainer } from './AbschlussContainer'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import { BarzahlungFelder } from './BarzahlungFelder'
import type { Position } from './Bestellung'
import { KommentarField } from './CommentField'
import {
  calculateZahlungsbetraege,
  toPositionRefs,
  toReceiptItems,
} from './drawerUtils'
import { Receipt } from './Receipt'
import { RestbetragZeile } from './RestbetragZeile'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'

interface ZahlungAbschlussProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  tisch: Tisch
  // Zu kassierende Positionen; `menge` ist die Auswahl-Menge, nicht die volle Position.
  positionenToPay: Position[]
  totalCents: number
  restNachZahlungCents: number
  zahlungKassiert: () => void
  pfandAusgezahlt: () => void
  variant: 'sheet' | 'spalte'
}

// Die Zahlung trägt keinen Client-Idempotenz-Schlüssel: die Idempotenz ist
// zustandsbasiert (bereits bezahlte Positionen → position_nicht_bezahlbar), und
// der Loading-Guard verhindert den Doppel-Submit. Eine Pfandrückgabe, allein
// oder verrechnet, trägt ihre rueckgabeId.
export function ZahlungAbschluss(props: ZahlungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)

  const noPositionenSelected = props.positionenToPay.length === 0
  const pfandAngebot = props.pfandProdukte.length > 0
  const pfand = pfandAuswahl(props.pfandProdukte, props.pfandVorgang)
  const mitPfand = pfand.summeCents > 0
  const nurPfand = noPositionenSelected && mitPfand
  // Mit Pfand ist nur die positive Differenz bar zu kassieren.
  const zuZahlenCents = props.totalCents - pfand.summeCents
  const barzahlung = !mitPfand || zuZahlenCents > 0
  const pfandVerrechnung = mitPfand
    ? {
        pfand: {
          rueckgabeId: pfand.rueckgabeId,
          positionen: pfand.positionen,
        },
      }
    : {}

  // In der dauerhaften Spalte überlebt der Eingabe-State sonst einen
  // Auswahl-Reset; eine neue Zusammenstellung startet deshalb mit leeren
  // Eingaben.
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
    zuZahlenCents,
    parseCents(erhaltenEuro),
    parseCents(zielbetragEuro),
  )

  const pfandVerwerfen = (error: unknown) => {
    if (istPfandrueckgabeAbweichend(error)) {
      pfand.steuerung.reset()
    }
  }

  const zahlung = useActionSubmit({
    actionLabel: 'Zahlung kassieren',
    onError: pfandVerwerfen,
    onSuccess: () => {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
      pfand.steuerung.reset()
      props.zahlungKassiert()
    },
  })
  const pfandAuszahlung = useActionSubmit({
    actionLabel: 'Pfand auszahlen',
    onError: pfandVerwerfen,
    onSuccess: () => {
      pfand.steuerung.reset()
      setKommentar('')
      props.pfandAusgezahlt()
    },
  })
  const loading = zahlung.loading || pfandAuszahlung.loading

  const onSubmit = async () => {
    await zahlung.run(async () => {
      await props.backend.zahlungKassieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(props.positionenToPay),
        kommentar,
        ...pfandVerrechnung,
      })
    })
  }

  const onPfandSubmit = async () => {
    await pfandAuszahlung.run(async () => {
      await props.pfandBackend.pfandZurueckgeben({
        rueckgabeId: pfand.rueckgabeId,
        tischId: props.tisch.id,
        positionen: pfand.positionen,
        kommentar,
      })
    })
  }

  const kommentarFeld = (
    <div className="px-4 pt-3">
      <KommentarField
        value={kommentar}
        onChange={(value) => {
          setKommentar(value)
        }}
      />
    </div>
  )

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
          !pfandAngebot && (
            <AbschlussLeer>
              Positionen auswählen, um zu kassieren.
            </AbschlussLeer>
          )
        ) : (
          <Receipt
            positionen={toReceiptItems(props.positionenToPay)}
            totalPrice={props.totalCents}
          />
        )}
        {pfandAngebot && (
          <PfandAbschnitt
            pfandProdukte={props.pfandProdukte}
            steuerung={pfand.steuerung}
            summeCents={pfand.summeCents}
            disabled={loading}
          />
        )}
        {!noPositionenSelected && (
          <>
            {mitPfand && (
              <PfandVerrechnung
                kaufsummeCents={props.totalCents}
                pfandCents={pfand.summeCents}
              />
            )}
            {barzahlung && (
              <BarzahlungFelder
                gesamtCents={zuZahlenCents}
                erhaltenEuro={erhaltenEuro}
                onErhaltenEuroChange={setErhaltenEuro}
                zielbetragEuro={zielbetragEuro}
                onZielbetragEuroChange={setZielbetragEuro}
                andererAktiv={andererAktiv}
                onAndererAktivChange={setAndererAktiv}
                rueckgeldCents={rueckgeldCents}
                trinkgeldCents={trinkgeldCents}
              />
            )}
            {kommentarFeld}
          </>
        )}
        {nurPfand && kommentarFeld}
      </DrawerBody>
      <DrawerFooter className="mx-auto w-full max-w-sm">
        {props.variant === 'spalte' && (
          <RestbetragZeile cents={props.restNachZahlungCents} />
        )}
        <Button
          disabled={loading || (noPositionenSelected && !nurPfand)}
          onClick={() => {
            void (nurPfand ? onPfandSubmit() : onSubmit())
          }}
        >
          {loading ? <Spinner /> : null}{' '}
          {nurPfand ? 'Pfand auszahlen' : 'Kassieren'}
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
