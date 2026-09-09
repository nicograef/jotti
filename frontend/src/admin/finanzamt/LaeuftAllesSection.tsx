import { ChevronDown } from 'lucide-react'

import { StatusDot } from '@/admin/components/StatusDot'
import {
  RUECKSTAND_WARN_SEKUNDEN,
  useTSESignaturQueue,
  useTSEStatus,
  useTSEStoerungen,
} from '@/admin/tse/hooks'
import { tseAmpel } from '@/admin/tse/tseAmpel'
import type { TSESignaturQueue, TSEStoerung } from '@/admin/tse/TSEBackend'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'

function formatDauer(sekunden: number): string {
  if (sekunden < 60) return `${String(sekunden)} s`
  const minuten = Math.floor(sekunden / 60)
  const rest = sekunden % 60
  return rest === 0
    ? `${String(minuten)} min`
    : `${String(minuten)} min ${String(rest)} s`
}

const STOERUNG_GRUND_LABEL: Record<TSEStoerung['grundArt'], string> = {
  tse_fehler: 'TSE-Fehler',
  rueckstand: 'Signatur-Rückstand',
  keine_konfiguration: 'TSE nicht konfiguriert',
}

// Kachel einer Roh-Metrik. `breit` ist für Fließtext gedacht (Fehlertext): über
// beide Spalten, kleiner gesetzt und umbrechend statt einstellig-groß.
function Kennzahl({
  label,
  wert,
  breit,
}: {
  label: string
  wert: string
  breit?: boolean
}) {
  return (
    <div
      className={cn(
        'flex flex-col rounded-md border p-3',
        breit && 'col-span-2',
      )}
    >
      <span className="text-sm text-muted-foreground">{label}</span>
      <span
        className={cn(
          'font-semibold',
          breit ? 'text-sm break-words' : 'text-lg tabular-nums',
        )}
      >
        {wert}
      </span>
    </div>
  )
}

// Aufklappbarer Detailblock: der Trigger zeigt das Label mit Pfeil, der Inhalt
// erscheint darunter.
function DetailCollapsible({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <Collapsible className="mt-1">
      <CollapsibleTrigger className="group flex items-center gap-1 text-xs text-muted-foreground hover:underline">
        {label}
        <ChevronDown className="size-3 transition-transform group-data-[state=open]:rotate-180" />
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2">{children}</CollapsibleContent>
    </Collapsible>
  )
}

// Klartext-Zusammenfassung der Signatur-Warteschlange plus die Roh-Metriken als
// aufklappbare technische Details. Fehlgeschlagene Signaturen stehen vorn: sie
// bleiben unabhängig von der Warteschlange liegen, bis der Kassenabschluss sie
// als Ausfall ausweist.
function SignaturPanel({ queue }: { queue: TSESignaturQueue | undefined }) {
  const offene = queue?.offeneAuftraege ?? 0
  const fehlgeschlagen = queue?.fehlgeschlageneAuftraege ?? 0
  const rueckstandSekunden = queue?.rueckstandSekunden ?? 0

  const saetze: string[] = []
  if (fehlgeschlagen > 0) {
    saetze.push(
      `${String(fehlgeschlagen)} ${fehlgeschlagen === 1 ? 'Vorgang ist' : 'Vorgänge sind'} fehlgeschlagen.`,
    )
  }
  if (offene === 0) {
    saetze.push('Keine Vorgänge in der Warteschlange.')
  } else {
    const warten = `${String(offene)} ${offene === 1 ? 'Vorgang wartet' : 'Vorgänge warten'} (ältester ${formatDauer(rueckstandSekunden)})`
    // Beruhigt wird nur unterhalb der Warnschwelle; darüber ist der Rückstand
    // derselbe Fehlerzustand, den die Ampel oben rot meldet.
    saetze.push(
      rueckstandSekunden < RUECKSTAND_WARN_SEKUNDEN
        ? `${warten} — normal bei vollem Betrieb.`
        : `${warten} — der Rückstand ist zu groß.`,
    )
  }
  if (fehlgeschlagen === 0) {
    saetze.push('Kein Vorgang fehlgeschlagen.')
  }
  const klartext = saetze.join(' ')

  return (
    <div className="flex flex-col gap-1 rounded-lg border p-4">
      <span className="text-sm font-semibold">Signatur-Warteschlange</span>
      <span className="text-sm leading-relaxed text-muted-foreground">
        {klartext}
      </span>
      {queue && (
        <DetailCollapsible label="Technische Details">
          <div className="grid grid-cols-2 gap-3">
            <Kennzahl
              label="Offene Aufträge"
              wert={String(queue.offeneAuftraege)}
            />
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
            <Kennzahl
              label="Fehlgeschlagen"
              wert={String(queue.fehlgeschlageneAuftraege)}
            />
            <Kennzahl
              label="Letzter Fehler"
              wert={queue.letzterFehler === '' ? '—' : queue.letzterFehler}
              breit
            />
          </div>
        </DetailCollapsible>
      )}
    </div>
  )
}

function StoerungRow({ stoerung }: { stoerung: TSEStoerung }) {
  const zeitraum = `${new Date(stoerung.beginn).toLocaleString('de-DE')} – ${
    stoerung.ende !== null
      ? new Date(stoerung.ende).toLocaleString('de-DE')
      : 'andauernd'
  }`

  return (
    <div className="flex flex-col gap-1 border-b py-3 last:border-b-0">
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

// Klartext-Zusammenfassung des Störungsprotokolls plus die vollständige
// Störungsliste als aufklappbares Protokoll.
function StoerungPanel({ stoerungen }: { stoerungen: TSEStoerung[] }) {
  const anzahl = stoerungen.length
  const klartext =
    anzahl === 0
      ? 'Keine dokumentierte Störung — die TSE-Signierung läuft ohne Ausfall.'
      : `${String(anzahl)} dokumentierte ${anzahl === 1 ? 'Störung' : 'Störungen'}. Wird automatisch für die gesetzliche Ausfalldokumentation geführt.`

  return (
    <div className="flex flex-col gap-1 rounded-lg border p-4">
      <span className="text-sm font-semibold">Störungsprotokoll</span>
      <span className="text-sm leading-relaxed text-muted-foreground">
        {klartext}
      </span>
      {anzahl > 0 && (
        <DetailCollapsible label="Protokoll ansehen">
          <div className="max-h-96 overflow-auto rounded-md border px-4">
            {stoerungen.map((stoerung) => (
              <StoerungRow key={stoerung.id} stoerung={stoerung} />
            ))}
          </div>
        </DetailCollapsible>
      )}
    </div>
  )
}

export function LaeuftAllesSection() {
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { queue } = useTSESignaturQueue()
  const { stoerungen } = useTSEStoerungen()

  const ampel = tseAmpel(tseStatus, tseLoading, queue)

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between gap-4 space-y-0">
        <CardTitle>Läuft alles?</CardTitle>
        <span
          className={cn(
            'inline-flex items-center gap-2 text-sm font-medium',
            ampel.fehler ? 'text-destructive' : 'text-primary',
          )}
        >
          <StatusDot
            zustand={ampel.fehler ? 'fehler' : 'ok'}
            label={ampel.ueberschrift}
          />
          {ampel.ueberschrift}
        </span>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <SignaturPanel queue={queue} />
        <StoerungPanel stoerungen={stoerungen} />
      </CardContent>
    </Card>
  )
}
