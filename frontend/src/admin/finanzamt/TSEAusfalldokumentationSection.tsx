import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { type TSENachsignierAuftrag } from '@/lib/EinstellungenBackend'

import { useTSENachsignierAuftraege } from '../settings/hooks'

const NACHSIGNIER_STATUS_LABEL: Record<
  TSENachsignierAuftrag['status'],
  string
> = {
  offen: 'Offen',
  erledigt: 'Erledigt',
  fehlgeschlagen: 'Fehlgeschlagen',
  verworfen: 'Verworfen',
  tse_nicht_konfiguriert: 'TSE nicht konfiguriert',
}

function NachsignierAuftragRow({
  auftrag,
  onZuruecksetzen,
  onVerwerfen,
}: {
  auftrag: TSENachsignierAuftrag
  onZuruecksetzen: (id: number) => Promise<void>
  onVerwerfen: (id: number) => Promise<void>
}) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Nachsignier-Auftrag aktualisieren',
  })

  const zeitraum = `${new Date(auftrag.erstelltAm).toLocaleString('de-DE')} – ${
    auftrag.erledigtAm !== null
      ? new Date(auftrag.erledigtAm).toLocaleString('de-DE')
      : 'offen'
  }`

  return (
    <div className="flex flex-col gap-2 py-4 border-b last:border-b-0">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div className="font-medium">
          {NACHSIGNIER_STATUS_LABEL[auftrag.status]} · {auftrag.processType}
        </div>
        <div className="text-sm text-muted-foreground">{zeitraum}</div>
      </div>
      <div className="text-sm text-muted-foreground">
        Transaktion: {auftrag.txId}
        {auftrag.versuche > 0 && <> · {auftrag.versuche} Versuche</>}
      </div>
      {auftrag.letzterFehler !== '' && (
        <p className="text-sm text-destructive">{auftrag.letzterFehler}</p>
      )}
      {auftrag.status === 'fehlgeschlagen' && (
        <div className="flex gap-2">
          <Button
            size="sm"
            disabled={loading}
            onClick={() =>
              void run(async () => {
                await onZuruecksetzen(auftrag.id)
                toast.success('Nachsignier-Auftrag wieder eingereiht.')
              })
            }
          >
            Zurücksetzen
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={loading}
            onClick={() =>
              void run(async () => {
                await onVerwerfen(auftrag.id)
                toast.success('Nachsignier-Auftrag verworfen.')
              })
            }
          >
            Verwerfen
          </Button>
        </div>
      )}
    </div>
  )
}

export function TSEAusfalldokumentationSection() {
  const { auftraege, isPending, error, zuruecksetzen, verwerfen } =
    useTSENachsignierAuftraege()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade Nachsignierungen…</p>
    )
  } else if (error) {
    inhalt = (
      <p className="text-destructive text-sm">
        Fehler beim Laden der Nachsignierungen.
      </p>
    )
  } else if (auftraege.length === 0) {
    inhalt = (
      <p className="text-muted-foreground text-sm">
        Keine Nachsignierungen — bisher wurden alle Vorgänge direkt signiert.
      </p>
    )
  } else {
    inhalt = (
      <div className="rounded-md border px-4 scrollbar-thin scrollbar-thumb-rounded scrollbar-thumb-muted max-h-96 overflow-auto">
        {auftraege.map((auftrag) => (
          <NachsignierAuftragRow
            key={auftrag.id}
            auftrag={auftrag}
            onZuruecksetzen={zuruecksetzen}
            onVerwerfen={verwerfen}
          />
        ))}
      </div>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>TSE-Ausfalldokumentation</CardTitle>
        <CardDescription>
          Vorgänge, die während eines TSE-Ausfalls erfasst wurden, werden hier
          automatisch nachsigniert. Die Liste dokumentiert zugleich die
          Ausfallzeiten (Beginn, Ende, Grund) für die gesetzliche
          Ausfalldokumentation. Fehlgeschlagene Aufträge können wieder
          eingereiht oder verworfen werden.
        </CardDescription>
      </CardHeader>
      <CardContent>{inhalt}</CardContent>
    </Card>
  )
}
