import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { Textarea } from '@/components/ui/textarea'
import { useActionSubmit } from '@/hooks/use-action-submit'
import {
  type TSESignaturauftrag,
  type TSESignaturQueue,
} from '@/lib/EinstellungenBackend'

import { useTSESignaturauftraege, useTSESignaturQueue } from '../settings/hooks'

const STATUS_LABEL: Record<TSESignaturauftrag['status'], string> = {
  offen: 'Offen',
  erledigt: 'Erledigt',
  fehlgeschlagen: 'Fehlgeschlagen',
  verworfen: 'Verworfen',
  tse_nicht_konfiguriert: 'TSE nicht konfiguriert',
}

function istEndgueltigMarkiert(status: TSESignaturauftrag['status']): boolean {
  return status === 'fehlgeschlagen' || status === 'tse_nicht_konfiguriert'
}

function istVerwerfbar(status: TSESignaturauftrag['status']): boolean {
  return status === 'offen' || status === 'fehlgeschlagen'
}

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

function VerwerfenDialog({
  auftrag,
  onVerwerfen,
}: {
  auftrag: TSESignaturauftrag
  onVerwerfen: (id: number, grund: string) => Promise<void>
}) {
  const [open, setOpen] = useState(false)
  const [grund, setGrund] = useState('')
  const { loading, run } = useActionSubmit({
    actionLabel: 'Signaturauftrag verwerfen',
  })

  const onOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) setGrund('')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Button
        size="sm"
        variant="outline"
        disabled={loading}
        onClick={() => {
          setOpen(true)
        }}
      >
        Verwerfen
      </Button>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Signaturauftrag verwerfen</DialogTitle>
          <DialogDescription>
            Der Auftrag wird endgültig verworfen und nicht mehr signiert. Der
            Grund wird zusammen mit Benutzer und Zeitpunkt protokolliert.
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-2">
          <Label htmlFor="verwerfen-grund">Begründung</Label>
          <Textarea
            id="verwerfen-grund"
            value={grund}
            onChange={(e) => {
              setGrund(e.target.value)
            }}
            placeholder="Warum wird dieser Auftrag verworfen?"
          />
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DialogClose>
          <Button
            variant="destructive"
            disabled={loading || grund.trim() === ''}
            onClick={() =>
              void run(async () => {
                await onVerwerfen(auftrag.id, grund.trim())
                setOpen(false)
                setGrund('')
                toast.success('Signaturauftrag verworfen.')
              })
            }
          >
            {loading ? <Spinner /> : null} Verwerfen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function SignaturauftragRow({
  auftrag,
  onZuruecksetzen,
  onVerwerfen,
}: {
  auftrag: TSESignaturauftrag
  onZuruecksetzen: (id: number) => Promise<void>
  onVerwerfen: (id: number, grund: string) => Promise<void>
}) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Signaturauftrag zurücksetzen',
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
          {STATUS_LABEL[auftrag.status]} · {auftrag.processType}
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
      {auftrag.status === 'verworfen' && auftrag.verworfenGrund !== '' && (
        <p className="text-sm text-muted-foreground">
          Verworfen von {auftrag.verworfenVon || 'unbekannt'}
          {auftrag.verworfenAm !== null &&
            ` am ${new Date(auftrag.verworfenAm).toLocaleString('de-DE')}`}
          : {auftrag.verworfenGrund}
        </p>
      )}
      {(istEndgueltigMarkiert(auftrag.status) ||
        istVerwerfbar(auftrag.status)) && (
        <div className="flex gap-2">
          {istEndgueltigMarkiert(auftrag.status) && (
            <Button
              size="sm"
              disabled={loading}
              onClick={() =>
                void run(async () => {
                  await onZuruecksetzen(auftrag.id)
                  toast.success('Signaturauftrag wieder eingereiht.')
                })
              }
            >
              Zurücksetzen
            </Button>
          )}
          {istVerwerfbar(auftrag.status) && (
            <VerwerfenDialog auftrag={auftrag} onVerwerfen={onVerwerfen} />
          )}
        </div>
      )}
    </div>
  )
}

export function SignaturauftraegeSection() {
  const {
    auftraege,
    isPending,
    error,
    zuruecksetzen,
    zuruecksetzenGesamt,
    verwerfen,
  } = useTSESignaturauftraege()
  const { queue } = useTSESignaturQueue()
  const { loading: gesamtLoading, run: runGesamt } = useActionSubmit({
    actionLabel: 'Alle Signaturaufträge zurücksetzen',
  })

  const zuruecksetzbar = auftraege.filter((a) =>
    istEndgueltigMarkiert(a.status),
  ).length

  let liste
  if (isPending) {
    liste = (
      <p className="text-muted-foreground text-sm">Lade Signaturaufträge…</p>
    )
  } else if (error) {
    liste = (
      <p className="text-destructive text-sm">
        Fehler beim Laden der Signaturaufträge.
      </p>
    )
  } else if (auftraege.length === 0) {
    liste = (
      <p className="text-muted-foreground text-sm">
        Keine Signaturaufträge — bisher wurden alle Vorgänge direkt signiert.
      </p>
    )
  } else {
    liste = (
      <div className="rounded-md border px-4 scrollbar-thin scrollbar-thumb-rounded scrollbar-thumb-muted max-h-96 overflow-auto">
        {auftraege.map((auftrag) => (
          <SignaturauftragRow
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
        <CardTitle>Signaturaufträge</CardTitle>
        <CardDescription>
          Zustand der Signatur-Queue und Verwaltung der Signaturaufträge.
          Endgültig fehlgeschlagene und ohne TSE erfasste Aufträge lassen sich
          wieder einreihen oder mit Begründung verwerfen.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {queue && <QueueZustand queue={queue} />}
        {zuruecksetzbar > 0 && (
          <div>
            <Button
              variant="outline"
              size="sm"
              disabled={gesamtLoading}
              onClick={() =>
                void runGesamt(async () => {
                  await zuruecksetzenGesamt()
                  toast.success('Alle markierten Aufträge wieder eingereiht.')
                })
              }
            >
              {gesamtLoading ? <Spinner /> : null} Alle zurücksetzen (
              {zuruecksetzbar})
            </Button>
          </div>
        )}
        {liste}
      </CardContent>
    </Card>
  )
}
