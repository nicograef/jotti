import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import { KassierAbschluss } from '../kassieren/KassierAbschluss'
import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { ReceiptPosition } from '../Receipt'
import type { PositionInput } from '../schemas'
import { direktverkaufBackend } from './hooks'

interface DirektverkaufAbschlussProps {
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  receiptItems: ReceiptPosition[]
  positionen: PositionInput[]
  totalCents: number
  // Liefert Kommentar und verkaufId; beide überstehen das Aushängen des Abschlusses.
  vorgang: Vorgang<number>
  verkaufAbgeschlossen: () => void
  // Leert den Korb ohne Erfolgsmeldung, damit der nächste Versuch eine neue verkaufId zieht.
  verkaufVerworfen: () => void
  pfandAusgezahlt: () => void
  variant: 'sheet' | 'spalte'
}

export function DirektverkaufAbschluss(props: DirektverkaufAbschlussProps) {
  return (
    <KassierAbschluss
      variant={props.variant}
      kopf={{
        eyebrow: 'Verkauf abschließen',
        title: 'Direktverkauf',
        description: 'Verkauf abschließen',
      }}
      leerHinweis="Produkte auswählen, um zu kassieren."
      aktion={{
        button: 'Verkauf abschließen',
        fehlerLabel: 'Verkauf abschließen',
      }}
      receiptItems={props.receiptItems}
      totalCents={props.totalCents}
      kassieren={(eingabe) => {
        const inhalt = { positionen: props.positionen, ...eingabe }
        return direktverkaufBackend.direktverkaufTaetigen({
          verkaufId: props.vorgang.schluesselFuer(inhalt),
          ...inhalt,
        })
      }}
      kommentar={props.vorgang.kommentar}
      setKommentar={props.vorgang.setKommentar}
      kassiert={props.verkaufAbgeschlossen}
      kaufVerworfen={props.verkaufVerworfen}
      pfandProdukte={props.pfandProdukte}
      pfandVorgang={props.pfandVorgang}
      pfandTischId={0}
      pfandAusgezahlt={props.pfandAusgezahlt}
    />
  )
}
