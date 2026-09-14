import { useQueryClient } from '@tanstack/react-query'
import { Check } from 'lucide-react'
import { type ReactNode, useEffect, useRef, useState } from 'react'

import { AdminPageHeader } from '@/admin/components/AdminPageHeader'
import { WarnKarte } from '@/admin/components/WarnKarte'
import { formatDatumLang } from '@/admin/reporting/utils'
import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { cn, formatEuro } from '@/lib/utils'

import { EroeffnenSection } from './EroeffnenSection'
import {
  GELDTRANSIT_LISTE_KEY,
  KASSENBESTAND_KEY,
  useAktiveKassensitzung,
  useKassenbestand,
} from './hooks'
import { KasseAbschliessenSection } from './KasseAbschliessenSection'
import type { AktiveKassensitzung } from './KasseBackend'
import { KassensitzungStatus } from './Kassensitzung'
import { LaufenderBetriebSection } from './LaufenderBetriebSection'

type StepState = 'done' | 'active' | 'inactive'

// Der Pop-Zustand wird beim Mounten erfasst, damit ein späteres Neurendern die
// Animation nicht abreißt.
function ErledigtHaekchen({ animiert }: { animiert: boolean }) {
  const [poppen] = useState(animiert)
  return <Check className={cn('size-4', poppen && 'animate-pop')} />
}

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

function EroeffnetKarte({
  kassensitzung,
  anfangsbestandCents,
  animieren,
}: {
  kassensitzung: AktiveKassensitzung
  anfangsbestandCents: number | null
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
            ? ` · Wechselgeld (Anfangsbestand): ${formatEuro(anfangsbestandCents)}`
            : ''}
        </p>
      </CardContent>
    </Card>
  )
}

export function KassensitzungPage() {
  const { kassensitzung, isPending, isError, refetch } =
    useAktiveKassensitzung()
  // Eigener Abruf für Schritt 1 (Anfangsbestand); TanStack Query dedupliziert
  // ihn mit dem Abruf in LaufenderBetriebSection.
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)
  const queryClient = useQueryClient()

  // Erkennt den Wechsel von geschlossener zu offener Kasse, um nur nach dem
  // Eröffnen zu animieren. Ref statt State (ein Rendern risse die Animation ab);
  // bleibt null, solange die erste Abfrage lädt, sonst gälte der Anfangszustand
  // als Wechsel.
  const istOffen = kassensitzung != null
  const zuletztOffenRef = useRef<boolean | null>(null)
  // eslint-disable-next-line react-hooks/refs
  const geradeEroeffnet = zuletztOffenRef.current === false && istOffen
  useEffect(() => {
    if (!isPending) {
      zuletztOffenRef.current = istOffen
    }
  }, [isPending, istOffen])

  // Hinter der Barriere lehnt das Backend jede Buchung ab (kasse_wird_abgeschlossen).
  const abschlussUnterbrochen =
    kassensitzung?.status === KassensitzungStatus.WIRD_ABGESCHLOSSEN

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
                    buchenMoeglich={!abschlussUnterbrochen}
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
                  {abschlussUnterbrochen && (
                    <WarnKarte className="mb-4">
                      Abschluss unterbrochen — erneut abschließen
                    </WarnKarte>
                  )}
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
