import { TriangleAlert } from 'lucide-react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import { useTSESignaturQueue } from '../tse/hooks'
import { type TSESignaturQueue } from '../tse/TSEBackend'

function formatDauer(sekunden: number): string {
  if (sekunden < 60) return `${String(sekunden)} s`
  const minuten = Math.floor(sekunden / 60)
  const rest = sekunden % 60
  return rest === 0
    ? `${String(minuten)} min`
    : `${String(minuten)} min ${String(rest)} s`
}

function Kennzahl({ label, wert }: { label: string; wert: string }) {
  return (
    <div className="flex flex-col rounded-md border p-3">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-lg font-semibold tabular-nums">{wert}</span>
    </div>
  )
}

function QueueZustand({ queue }: { queue: TSESignaturQueue }) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <Kennzahl label="Offene Aufträge" wert={String(queue.offeneAuftraege)} />
      <Kennzahl
        label="Ältester offen"
        wert={
          queue.offeneAuftraege === 0
            ? '—'
            : formatDauer(queue.rueckstandSekunden)
        }
      />
      <Kennzahl
        label="Signaturen/Minute"
        wert={queue.signaturenProMinute.toFixed(1)}
      />
      <Kennzahl
        label="Signierdauer p95"
        wert={`${queue.signierdauerP95Sekunden.toFixed(1)} s`}
      />
    </div>
  )
}

export function SignaturauftraegeSection() {
  const { queue } = useTSESignaturQueue()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Signaturaufträge</CardTitle>
        <CardDescription>
          Zustand der Signatur-Queue: Rückstand und Durchsatz der
          TSE-Signierung.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {queue && (
          <>
            <QueueZustand queue={queue} />
            {queue.fehlgeschlageneAuftraege > 0 && (
              <Alert variant="destructive">
                <TriangleAlert className="size-4" />
                <AlertTitle>
                  {queue.fehlgeschlageneAuftraege === 1
                    ? '1 Vorgang konnte nicht signiert werden'
                    : `${String(queue.fehlgeschlageneAuftraege)} Vorgänge konnten nicht signiert werden`}
                </AlertTitle>
                {queue.letzterFehler && (
                  <AlertDescription>{queue.letzterFehler}</AlertDescription>
                )}
              </Alert>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
