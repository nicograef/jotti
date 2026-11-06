import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { PfandBackend } from '../pfand/PfandBackend'
import { KassierAbschluss } from '../table/KassierAbschluss'
import type { ReceiptPosition } from '../table/Receipt'
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
        return props.backend.direktverkaufTaetigen({
          verkaufId: props.vorgang.schluesselFuer(inhalt),
          ...inhalt,
        })
      }}
      kommentar={props.vorgang.kommentar}
      setKommentar={props.vorgang.setKommentar}
      kassiert={props.verkaufAbgeschlossen}
      kaufVerworfen={props.verkaufVerworfen}
      pfandBackend={props.pfandBackend}
      pfandProdukte={props.pfandProdukte}
      pfandVorgang={props.pfandVorgang}
      pfandTischId={0}
      pfandAusgezahlt={props.pfandAusgezahlt}
    />
  )
}
