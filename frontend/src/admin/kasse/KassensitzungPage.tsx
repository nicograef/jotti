import { useQueryClient } from '@tanstack/react-query'
import { Check } from 'lucide-react'
import { type ReactNode, useEffect, useRef, useState } from 'react'

import { AdminPageHeader } from '@/admin/components/AdminPageHeader'
import { formatDatumLang } from '@/admin/reporting/utils'
import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { cn, formatCents } from '@/lib/utils'

import { EroeffnenSection } from './EroeffnenSection'
import {
  GELDTRANSIT_LISTE_KEY,
  KASSENBESTAND_KEY,
  useKassenbestand,
  useOffeneKassensitzung,
} from './hooks'
import { KasseAbschliessenSection } from './KasseAbschliessenSection'
import type { OffeneKassensitzung } from './KasseBackend'
import { LaufenderBetriebSection } from './LaufenderBetriebSection'

export { EroeffnenSection } from './EroeffnenSection'
export { KasseAbschliessenSection } from './KasseAbschliessenSection'

type StepState = 'done' | 'active' | 'inactive'

// ErledigtHaekchen rendert das Häkchen des erledigten Schritts. Es poppt
// (Motion-Inventar „Statuswechsel", 350 ms), wenn der Schritt gerade auf
// „erledigt" wechselt — nicht, wenn die Seite bereits erledigt lädt. Der
// Pop-Zustand wird beim Mounten erfasst, damit ein späteres Neurendern (z. B.
// eintreffender Kassenbestand) die Animation nicht abreißt.
function ErledigtHaekchen({ animiert }: { animiert: boolean }) {
  const [poppen] = useState(animiert)
  return <Check className={cn('size-4', poppen && 'animate-pop')} />
}

// StepperRow rendert die Nummern-Schiene (Kreis + Verbindungslinie) links und den
// Schritt-Inhalt rechts. Der erledigte Schritt (done) bekommt ein Häkchen, der
// aktive Schritt einen umrandeten Kreis, inaktive Schritte sind ausgegraut.
function StepperRow({
  nummer,
  state,
  istLetzter,
  markerAnimiert,
  children,
}: {
  nummer: number
  state: StepState
  istLetzter?: boolean
  // Lässt das erledigt-Häkchen einmalig poppen, wenn der Schritt gerade auf
  // „erledigt" wechselt (nur Schritt 1 nach dem Eröffnen).
  markerAnimiert?: boolean
  children: ReactNode
}) {
  return (
    <div className="flex items-stretch gap-4">
      <div className="flex w-7 shrink-0 flex-col items-center">
        <span
          className={cn(
            'flex size-7 items-center justify-center rounded-full text-sm font-bold',
            state === 'done' && 'bg-primary text-primary-foreground',
            state === 'active' &&
              'border-2 border-primary bg-background text-primary',
            state === 'inactive' &&
              'border-2 border-border bg-background text-muted-foreground',
          )}
        >
          {state === 'done' ? (
            <ErledigtHaekchen animiert={markerAnimiert ?? false} />
          ) : (
            nummer
          )}
        </span>
        {!istLetzter && <span className="mt-1 w-0.5 flex-1 bg-border" />}
      </div>
      <div
        className={cn('min-w-0 flex-1', state === 'inactive' && 'opacity-50')}
      >
        {children}
      </div>
    </div>
  )
}

// EroeffnetKarte ist die flache „Kasse eröffnet"-Karte (Schritt 1) bei laufender
// Sitzung: Eröffnungszeitpunkt und Anfangsbestand als Einzeiler. Der
// Anfangsbestand stammt aus der Kassenbestand-Aufschlüsselung.
function EroeffnetKarte({
  kassensitzung,
  anfangsbestandCents,
  animieren,
}: {
  kassensitzung: OffeneKassensitzung
  anfangsbestandCents: number | null
  // Lässt die Karte einmalig mit fadeUp eintreten, wenn sie gerade durch das
  // Eröffnen erscheint (nicht beim Laden einer bereits offenen Kasse).
  animieren: boolean
}) {
  const eroeffnetAm = new Date(kassensitzung.eroeffnetAm).toLocaleString(
    'de-DE',
    { dateStyle: 'medium', timeStyle: 'short' },
  )
  // Beim Mount erfasst, damit ein späteres Neurendern die Animation nicht abreißt.
  const [initialAnimieren] = useState(animieren)
  return (
    <Card
      className={cn(
        'bg-muted/30',
        initialAnimieren &&
          'animate-fade-up [animation-duration:450ms] [animation-timing-function:cubic-bezier(0.2,0.7,0.3,1)]',
      )}
    >
      <CardContent className="py-4">
        <div className="text-sm font-semibold">1 · Kasse eröffnet</div>
        <p className="mt-0.5 text-sm text-muted-foreground">
          {eroeffnetAm}
          {anfangsbestandCents !== null
            ? ` · Wechselgeld (Anfangsbestand): ${formatCents(anfangsbestandCents)} €`
            : ''}
        </p>
      </CardContent>
    </Card>
  )
}

export function KassensitzungPage() {
  const { kassensitzung, isPending, isError, refetch } =
    useOffeneKassensitzung()
  // Kassenbestand-Aufschlüsselung für Schritt 1 (Anfangsbestand); TanStack Query
  // dedupliziert mit dem Abruf innerhalb von LaufenderBetriebSection.
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)
  const queryClient = useQueryClient()

  // „Gerade eröffnet" erkennt den Wechsel von geschlossener zu offener Kasse,
  // um die erledigt-Karte nur nach dem Eröffnen zu animieren — nicht beim Laden
  // einer bereits offenen Kasse. Der Ref bleibt `null`, solange die erste
  // Abfrage lädt (`isPending`), damit der Anfangszustand nicht als Wechsel gilt.
  // Er liegt bewusst in einem Ref (kein zusätzliches Rendern, das die Animation
  // abreißen würde); der Schreibzugriff erfolgt nur im Effekt.
  const istOffen = kassensitzung != null
  const zuletztOffenRef = useRef<boolean | null>(null)
  // eslint-disable-next-line react-hooks/refs
  const geradeEroeffnet = zuletztOffenRef.current === false && istOffen
  useEffect(() => {
    if (!isPending) {
      zuletztOffenRef.current = istOffen
    }
  }, [isPending, istOffen])

  const titel = kassensitzung
    ? `Kassentag Nr. ${String(kassensitzung.zNr)} — ${kassensitzung.bezeichnung}`
    : 'Kassentag'
  const unterzeile = kassensitzung
    ? `${formatDatumLang(kassensitzung.datum)} · Ein Kassentag läuft von der Eröffnung bis zum Tagesabschluss (Z-Bon).`
    : 'Ein Kassentag läuft von der Eröffnung bis zum Tagesabschluss (Z-Bon).'

  const header = (
    <AdminPageHeader
      titel={titel}
      unterzeile={unterzeile}
      glowFarben={['orange', 'teal']}
    />
  )

  if (isPending) {
    return (
      <>
        {header}
        <p className="mt-4 text-muted-foreground">Laden…</p>
      </>
    )
  }

  // Expliziter Fehlerzustand statt des Leer-Defaults — sonst wirkt die Kasse bei
  // Netzabbruch fälschlich geschlossen.
  if (isError) {
    return (
      <>
        {header}
        <LadefehlerAlert
          titel="Kassendaten konnten nicht geladen werden"
          onErneutVersuchen={() => void refetch()}
          className="mt-4"
        />
      </>
    )
  }

  // Nach einer Geldtransit-Buchung Kassenbestand und Bewegungsliste neu laden.
  const invalidateKasse = () => {
    void queryClient.invalidateQueries({ queryKey: [KASSENBESTAND_KEY] })
    void queryClient.invalidateQueries({ queryKey: [GELDTRANSIT_LISTE_KEY] })
  }

  return (
    <>
      {header}

      <div className="mt-6 flex max-w-4xl flex-col gap-4">
        {kassensitzung ? (
          <>
            <StepperRow
              nummer={1}
              state="done"
              markerAnimiert={geradeEroeffnet}
            >
              <EroeffnetKarte
                kassensitzung={kassensitzung}
                anfangsbestandCents={kassenbestand?.anfangsbestandCents ?? null}
                animieren={geradeEroeffnet}
              />
            </StepperRow>

            <StepperRow nummer={2} state="active">
              <Card>
                <CardHeader>
                  <CardTitle>2 · Laufender Betrieb</CardTitle>
                </CardHeader>
                <CardContent>
                  <LaufenderBetriebSection
                    kassensitzungNr={kassensitzung.zNr}
                    onBuchung={invalidateKasse}
                  />
                </CardContent>
              </Card>
            </StepperRow>

            <StepperRow nummer={3} state="active" istLetzter>
              <Card>
                <CardHeader>
                  <CardTitle>
                    3 · Am Ende des Tages: Kasse abschließen
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <KasseAbschliessenSection
                    kassensitzungNr={kassensitzung.zNr}
                    onSuccess={() => void refetch()}
                  />
                </CardContent>
              </Card>
            </StepperRow>
          </>
        ) : (
          <>
            <StepperRow nummer={1} state="active">
              <EroeffnenSection onSuccess={() => void refetch()} />
            </StepperRow>

            <StepperRow nummer={2} state="inactive">
              <Card>
                <CardHeader>
                  <CardTitle>2 · Laufender Betrieb</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">
                    Sobald die Kasse eröffnet ist, erscheinen hier der
                    Soll-Bestand und die heutigen Kassenbewegungen.
                  </p>
                </CardContent>
              </Card>
            </StepperRow>

            <StepperRow nummer={3} state="inactive" istLetzter>
              <Card>
                <CardHeader>
                  <CardTitle>
                    3 · Am Ende des Tages: Kasse abschließen
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">
                    Am Ende des Tages werden hier Kassensturz und Tagesabschluss
                    (Z-Bon) in einem Schritt gebucht.
                  </p>
                </CardContent>
              </Card>
            </StepperRow>
          </>
        )}
      </div>
    </>
  )
}
