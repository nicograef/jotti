import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import { toPositionRefs, toReceiptItems } from '../drawerUtils'
import { KassierAbschluss } from '../kassieren/KassierAbschluss'
import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { Position } from './Bestellung'
import { tischBackend } from './hooks'
import { RestbetragZeile } from './RestbetragZeile'
import type { Tisch } from './Tisch'

interface ZahlungAbschlussProps {
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  tisch: Tisch
  // Zu kassierende Positionen; `menge` ist die Auswahl-Menge, nicht die volle Position.
  positionenToPay: Position[]
  totalCents: number
  restNachZahlungCents: number
  // Kassieren-Vorgang von TablePage: Kommentar und zahlungId überstehen wie die
  // Auswahl das Aushängen des Abschlusses.
  vorgang: Vorgang<string>
  zahlungKassiert: () => void
  pfandAusgezahlt: () => void
  variant: 'sheet' | 'spalte'
}

// Die zahlungId hängt an Positionen und Pfand, nicht am Kommentar: wer ihn nach
// einer verlorenen Antwort ändert, wiederholt dieselbe Zahlung, statt eine
// Teilzahlung doppelt zu buchen.
export function ZahlungAbschluss(props: ZahlungAbschlussProps) {
  return (
    <KassierAbschluss
      variant={props.variant}
      kopf={{
        eyebrow: 'Zahlung für',
        title: props.tisch.name,
        description: `Zahlung für ${props.tisch.name}`,
      }}
      leerHinweis="Positionen auswählen, um zu kassieren."
      aktion={{ button: 'Kassieren', fehlerLabel: 'Zahlung kassieren' }}
      receiptItems={toReceiptItems(props.positionenToPay)}
      totalCents={props.totalCents}
      kommentar={props.vorgang.kommentar}
      setKommentar={props.vorgang.setKommentar}
      kassieren={(eingabe) => {
        const positionen = toPositionRefs(props.positionenToPay)
        return tischBackend.zahlungKassieren({
          zahlungId: props.vorgang.schluesselFuer({
            positionen,
            pfand: eingabe.pfand,
          }),
          tischId: props.tisch.id,
          positionen,
          ...eingabe,
        })
      }}
      kassiert={props.zahlungKassiert}
      pfandProdukte={props.pfandProdukte}
      pfandVorgang={props.pfandVorgang}
      pfandTischId={props.tisch.id}
      pfandAusgezahlt={props.pfandAusgezahlt}
      spaltenFuss={<RestbetragZeile cents={props.restNachZahlungCents} />}
    />
  )
}
