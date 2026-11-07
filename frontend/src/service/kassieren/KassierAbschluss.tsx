import { type ReactNode, useCallback, useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import type { Produkt } from '@/lib/produktSchemas'
import { parseCents } from '@/lib/utils'

import { AbschlussContainer } from '../AbschlussContainer'
import { AbschlussHeader } from '../AbschlussHeader'
import { AbschlussLeer } from '../AbschlussLeer'
import { KommentarField } from '../CommentField'
import { calculateZahlungsbetraege } from '../drawerUtils'
import { pfandBackend } from '../pfand/hooks'
import { PfandAbschnitt } from '../pfand/PfandAbschnitt'
import {
  istPfandrueckgabeAbweichend,
  pfandAuswahl,
  type PfandVorgang,
} from '../pfand/pfandAuswahl'
import { PfandVerrechnung } from '../pfand/PfandVerrechnung'
import { Receipt, type ReceiptPosition } from '../Receipt'
import type { PfandVerrechnung as PfandVerrechnungEingabe } from '../schemas'
import { BarzahlungFelder } from './BarzahlungFelder'

// Was der Kauf zusätzlich zu seinen Positionen bucht; `pfand` fehlt ohne Pfand.
export interface KassierEingabe {
  kommentar: string
  pfand?: PfandVerrechnungEingabe
}

interface KassierAbschlussProps {
  kopf: { eyebrow: string; title: string; description: string }
  leerHinweis: string
  // Beschriftung des Kauf-Buttons und Name der Aktion in der Fehlermeldung.
  aktion: { button: string; fehlerLabel: string }
  receiptItems: ReceiptPosition[]
  totalCents: number
  // Nur Direktverkauf: die Tisch-Zahlung kommentiert keine Positionen.
  kommentarBearbeiten?: (id: string, text: string) => void
  // Liegt im Vorgang der Seite, damit ein Retry nach dem Aushängen denselben Inhalt trägt.
  kommentar: string
  setKommentar: (kommentar: string) => void
  kassieren: (eingabe: KassierEingabe) => Promise<void>
  kassiert: () => void
  // Läuft, wenn der Server den Kauf mit `pfandrueckgabe_abweichend` abweist.
  kaufVerworfen?: () => void
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  // 0 für den Direktverkauf.
  pfandTischId: number
  pfandAusgezahlt: () => void
  // Nur in der Spalte, über dem Button.
  spaltenFuss?: ReactNode
  variant: 'sheet' | 'spalte'
}

function useBarzahlungEingaben() {
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)
  const zuruecksetzen = useCallback(() => {
    setErhaltenEuro('')
    setZielbetragEuro('')
    setAndererAktiv(false)
  }, [])
  return {
    erhaltenEuro,
    setErhaltenEuro,
    zielbetragEuro,
    setZielbetragEuro,
    andererAktiv,
    setAndererAktiv,
    zuruecksetzen,
  }
}

// Gemeinsamer Abschluss von Zahlung und Direktverkauf: Barzahlung mit Rückgeld
// und Trinkgeld, Pfand verrechnet oder allein ausgezahlt.
export function KassierAbschluss(props: KassierAbschlussProps) {
  const eingaben = useBarzahlungEingaben()

  const noPositionenSelected = props.receiptItems.length === 0
  const pfandAngebot = props.pfandProdukte.length > 0
  const pfand = pfandAuswahl(props.pfandProdukte, props.pfandVorgang)
  const mitPfand = pfand.summeCents > 0
  const nurPfand = noPositionenSelected && mitPfand
  // Mit Pfand ist nur die positive Differenz bar zu kassieren.
  const zuZahlenCents = props.totalCents - pfand.summeCents
  const barzahlung = !mitPfand || zuZahlenCents > 0

  // In der dauerhaften Spalte überlebt der Eingabe-State sonst einen
  // Auswahl-Reset; eine neue Zusammenstellung startet mit leeren Eingaben.
  const warLeerRef = useRef(noPositionenSelected)
  const { zuruecksetzen } = eingaben
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      zuruecksetzen()
    }
    warLeerRef.current = noPositionenSelected
  }, [noPositionenSelected, zuruecksetzen])

  const { rueckgeldCents, trinkgeldCents } = calculateZahlungsbetraege(
    zuZahlenCents,
    parseCents(eingaben.erhaltenEuro),
    parseCents(eingaben.zielbetragEuro),
  )

  // Der Korb bleibt sonst stehen: mit demselben Schlüssel wiese der Server jeden Versuch erneut ab.
  // Schon gebuchte Ware steht in der Historie, ihr erneutes Erfassen ist eine bewusste Handlung.
  const kauf = useActionSubmit({
    actionLabel: props.aktion.fehlerLabel,
    onError: (error) => {
      if (istPfandrueckgabeAbweichend(error)) {
        pfand.steuerung.reset()
        props.kaufVerworfen?.()
      }
    },
    onSuccess: () => {
      eingaben.zuruecksetzen()
      pfand.steuerung.reset()
      props.kassiert()
    },
  })
  const pfandAuszahlung = useActionSubmit({
    actionLabel: 'Pfand auszahlen',
    onError: (error) => {
      if (istPfandrueckgabeAbweichend(error)) {
        pfand.steuerung.reset()
      }
    },
    onSuccess: () => {
      pfand.steuerung.reset()
      props.setKommentar('')
      props.pfandAusgezahlt()
    },
  })
  const loading = kauf.loading || pfandAuszahlung.loading

  const onSubmit = async () => {
    await kauf.run(async () => {
      await props.kassieren({
        kommentar: props.kommentar,
        ...(mitPfand
          ? {
              pfand: {
                rueckgabeId: pfand.rueckgabeId,
                positionen: pfand.positionen,
              },
            }
          : {}),
      })
    })
  }

  const onPfandSubmit = async () => {
    await pfandAuszahlung.run(async () => {
      await pfandBackend.pfandZurueckgeben({
        rueckgabeId: pfand.rueckgabeId,
        tischId: props.pfandTischId,
        positionen: pfand.positionen,
        kommentar: props.kommentar,
      })
    })
  }

  const kommentarFeld = (
    <div className="px-4 pt-3">
      <KommentarField value={props.kommentar} onChange={props.setKommentar} />
    </div>
  )

  return (
    <AbschlussContainer variant={props.variant} pending={loading}>
      <AbschlussHeader variant={props.variant} {...props.kopf} />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          !pfandAngebot && <AbschlussLeer>{props.leerHinweis}</AbschlussLeer>
        ) : (
          <Receipt
            positionen={props.receiptItems}
            totalPrice={props.totalCents}
            kommentarBearbeiten={props.kommentarBearbeiten}
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
                erhaltenEuro={eingaben.erhaltenEuro}
                onErhaltenEuroChange={eingaben.setErhaltenEuro}
                zielbetragEuro={eingaben.zielbetragEuro}
                onZielbetragEuroChange={eingaben.setZielbetragEuro}
                andererAktiv={eingaben.andererAktiv}
                onAndererAktivChange={eingaben.setAndererAktiv}
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
        {props.variant === 'spalte' && props.spaltenFuss}
        <Button
          disabled={loading || (noPositionenSelected && !nurPfand)}
          onClick={() => {
            void (nurPfand ? onPfandSubmit() : onSubmit())
          }}
        >
          {loading ? <Spinner /> : null}{' '}
          {nurPfand ? 'Pfand auszahlen' : props.aktion.button}
        </Button>
        {props.variant === 'sheet' && (
          <DrawerClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DrawerClose>
        )}
      </DrawerFooter>
    </AbschlussContainer>
  )
}
