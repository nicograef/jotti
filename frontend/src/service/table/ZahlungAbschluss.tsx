import type { Produkt } from '@/lib/produktSchemas'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { PfandBackend } from '../pfand/PfandBackend'
import type { Position } from './Bestellung'
import { toPositionRefs, toReceiptItems } from './drawerUtils'
import { KassierAbschluss } from './KassierAbschluss'
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
  // Liegt im Kassieren-Vorgang von TablePage, wie die Auswahl.
  kommentar: string
  setKommentar: (kommentar: string) => void
  zahlungKassiert: () => void
  pfandAusgezahlt: () => void
  variant: 'sheet' | 'spalte'
}

// Die Zahlung trägt keinen Client-Idempotenz-Schlüssel: die Idempotenz ist
// zustandsbasiert (bereits bezahlte Positionen → position_nicht_bezahlbar), und
// der Loading-Guard verhindert den Doppel-Submit. Eine Pfandrückgabe, allein
// oder verrechnet, trägt ihre rueckgabeId.
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
      kommentar={props.kommentar}
      setKommentar={props.setKommentar}
      kassieren={(eingabe) =>
        props.backend.zahlungKassieren({
          tischId: props.tisch.id,
          positionen: toPositionRefs(props.positionenToPay),
          ...eingabe,
        })
      }
      kassiert={props.zahlungKassiert}
      pfandBackend={props.pfandBackend}
      pfandProdukte={props.pfandProdukte}
      pfandVorgang={props.pfandVorgang}
      pfandTischId={props.tisch.id}
      pfandAusgezahlt={props.pfandAusgezahlt}
      spaltenFuss={<RestbetragZeile cents={props.restNachZahlungCents} />}
    />
  )
}
