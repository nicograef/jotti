import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { type TSEStoerung } from '@/lib/EinstellungenBackend'

import { useTSEStoerungen } from '../settings/hooks'

const STOERUNG_GRUND_LABEL: Record<TSEStoerung['grundArt'], string> = {
  tse_fehler: 'TSE-Fehler',
  rueckstand: 'Signatur-Rückstand',
  keine_konfiguration: 'TSE nicht konfiguriert',
}

function StoerungRow({ stoerung }: { stoerung: TSEStoerung }) {
  const zeitraum = `${new Date(stoerung.beginn).toLocaleString('de-DE')} – ${
    stoerung.ende !== null
      ? new Date(stoerung.ende).toLocaleString('de-DE')
      : 'andauernd'
  }`

  return (
    <div className="flex flex-col gap-1 py-4 border-b last:border-b-0">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div className="font-medium">
          {STOERUNG_GRUND_LABEL[stoerung.grundArt]}
        </div>
        <div className="text-sm text-muted-foreground">{zeitraum}</div>
      </div>
      {stoerung.fehlertext !== '' && (
        <p className="text-sm text-muted-foreground">{stoerung.fehlertext}</p>
      )}
    </div>
  )
}

export function TSEAusfalldokumentationSection() {
  const { stoerungen, isPending, error } = useTSEStoerungen()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade Störungsprotokoll…</p>
    )
  } else if (error) {
    inhalt = (
      <p className="text-destructive text-sm">
        Fehler beim Laden des Störungsprotokolls.
      </p>
    )
  } else if (stoerungen.length === 0) {
    inhalt = (
      <p className="text-muted-foreground text-sm">
        Keine Störungen — die TSE-Signierung lief bisher ohne dokumentierten
        Ausfall.
      </p>
    )
  } else {
    inhalt = (
      <div className="rounded-md border px-4 scrollbar-thin scrollbar-thumb-rounded scrollbar-thumb-muted max-h-96 overflow-auto">
        {stoerungen.map((stoerung) => (
          <StoerungRow key={stoerung.id} stoerung={stoerung} />
        ))}
      </div>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>TSE-Ausfalldokumentation</CardTitle>
        <CardDescription>
          Das Störungsprotokoll dokumentiert die Ausfallzeiten der
          TSE-Signierung (Beginn, Ende, Grund) für die gesetzliche
          Ausfalldokumentation. Ein andauernder Zeitraum ohne Ende weist auf
          eine aktive Störung hin.
        </CardDescription>
      </CardHeader>
      <CardContent>{inhalt}</CardContent>
    </Card>
  )
}
