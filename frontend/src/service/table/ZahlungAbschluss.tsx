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
        tischBackend.zahlungKassieren({
          tischId: props.tisch.id,
          positionen: toPositionRefs(props.positionenToPay),
          ...eingabe,
        })
      }
      kassiert={props.zahlungKassiert}
      pfandProdukte={props.pfandProdukte}
      pfandVorgang={props.pfandVorgang}
      pfandTischId={props.tisch.id}
      pfandAusgezahlt={props.pfandAusgezahlt}
      spaltenFuss={<RestbetragZeile cents={props.restNachZahlungCents} />}
    />
  )
}
