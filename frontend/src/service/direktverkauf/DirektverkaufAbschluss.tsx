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
import { AbschlussContainer } from '../table/AbschlussContainer'
import { AbschlussHeader } from '../table/AbschlussHeader'
import { AbschlussLeer } from '../table/AbschlussLeer'
import { BarzahlungFelder } from '../table/BarzahlungFelder'
import { KommentarField } from '../table/CommentField'
import { calculateZahlungsbetraege } from '../table/drawerUtils'
import type { ReceiptPosition } from '../table/Receipt'
import { Receipt } from '../table/Receipt'
import type { VerkaufPositionInput } from './Direktverkauf'
import type { DirektverkaufBackend } from './DirektverkaufBackend'

interface DirektverkaufAbschlussProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  receiptItems: ReceiptPosition[]
  positionen: VerkaufPositionInput[]
  totalCents: number
  verkaufAbgeschlossen: () => void
  // Leert den Korb ohne Erfolgsmeldung, damit der nächste Versuch eine neue verkaufId zieht.
  verkaufVerworfen: () => void
  pfandAusgezahlt: () => void
  variant: 'sheet' | 'spalte'
}

export function DirektverkaufAbschluss(props: DirektverkaufAbschlussProps) {
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)
  const [kommentar, setKommentar] = useState('')

  const noPositionenSelected = props.positionen.length === 0
  const pfandAngebot = props.pfandProdukte.length > 0
  // One verkaufId per logical Vorgang, renewed when a selection starts from empty; a retry keeps its key.
  // The new key also resets the inputs, so the persistent column carries nothing over from an abandoned Vorgang.
  const [verkaufId, setVerkaufId] = useState(() => crypto.randomUUID())
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

  const warLeerRef = useRef(noPositionenSelected)
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      setVerkaufId(crypto.randomUUID())
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

  // Der Korb bleibt sonst stehen: mit derselben verkaufId wiese der Server jeden Versuch erneut ab.
  // Schon gebuchte Ware steht in der Historie, ihr erneutes Erfassen ist eine bewusste Handlung.
  const verkaufVerwerfen = (error: unknown) => {
    if (istPfandrueckgabeAbweichend(error)) {
      pfand.steuerung.reset()
      props.verkaufVerworfen()
    }
  }

  const verkauf = useActionSubmit({
    actionLabel: 'Verkauf abschließen',
    onError: verkaufVerwerfen,
    onSuccess: () => {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
      pfand.steuerung.reset()
      props.verkaufAbgeschlossen()
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
  const loading = verkauf.loading || pfandAuszahlung.loading

  const onSubmit = async () => {
    await verkauf.run(async () => {
      await props.backend.direktverkaufTaetigen({
        verkaufId,
        positionen: props.positionen,
        kommentar,
        ...pfandVerrechnung,
      })
    })
  }

  const onPfandSubmit = async () => {
    await pfandAuszahlung.run(async () => {
      await props.pfandBackend.pfandZurueckgeben({
        rueckgabeId: pfand.rueckgabeId,
        tischId: 0,
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
        eyebrow="Verkauf abschließen"
        title="Direktverkauf"
        description="Verkauf abschließen"
      />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          !pfandAngebot && (
            <AbschlussLeer>Produkte auswählen, um zu kassieren.</AbschlussLeer>
          )
        ) : (
          <Receipt
            positionen={props.receiptItems}
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
        <Button
          disabled={loading || (noPositionenSelected && !nurPfand)}
          onClick={() => {
            void (nurPfand ? onPfandSubmit() : onSubmit())
          }}
        >
          {loading ? <Spinner /> : null}{' '}
          {nurPfand ? 'Pfand auszahlen' : 'Verkauf abschließen'}
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
