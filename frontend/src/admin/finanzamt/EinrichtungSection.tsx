import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { NavLink } from 'react-router'
import { toast } from 'sonner'

import { StatusDot } from '@/admin/components/StatusDot'
import { WarnKarte } from '@/admin/components/WarnKarte'
import { formatDatum } from '@/admin/reporting/utils'
import { useTSEStatus } from '@/admin/tse/hooks'
import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { BetreiberForm } from './BetreiberForm'
import { useBetreiber, useKassenidentitaet } from './hooks'

const LEITFADEN_URL = 'https://jotti.rocks/docs/leitfaden/finanzamt-anmelden/'

// Vereinsdaten gelten als erledigt, wenn die Pflichtfelder der Betreiber-Query
// (Vereinsname, Straße, PLZ, Ort — vgl. betreiberSchema im Backend) gefüllt
// sind. Es gibt bewusst kein separates „vollständig"-Flag im Backend.
function vereinsdatenErledigt(betreiber: {
  vereinsname: string
  strasse: string
  plz: string
  ort: string
}): boolean {
  return (
    betreiber.vereinsname.trim() !== '' &&
    betreiber.strasse.trim() !== '' &&
    betreiber.plz.trim() !== '' &&
    betreiber.ort.trim() !== ''
  )
}

// Neutrale/erledigte Checklisten-Karte mit Nummer, Titel und Status-Icon. Der
// Fehlerzustand eines Schritts wird stattdessen über die WarnKarte gerendert
// (siehe Schritt 2/3), damit die Destructive-Token an einer Stelle wohnen.
function SchrittKarte({
  nummer,
  titel,
  erledigt,
  children,
}: {
  nummer: number
  titel: string
  erledigt: boolean
  children: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-1.5 rounded-lg border bg-sidebar p-4">
      <span className="flex items-center gap-2 text-sm font-semibold">
        {erledigt ? (
          <Check className="size-4 text-primary" aria-hidden />
        ) : (
          <StatusDot zustand="neutral" label="offen" />
        )}
        {nummer} · {titel}
      </span>
      {children}
    </div>
  )
}

export function EinrichtungSection() {
  const {
    betreiber,
    isPending: betreiberLoading,
    isLoadingError: betreiberError,
    refetchBetreiber,
    saveBetreiber,
    setElsterMeldung,
    nimmElsterMeldungZurueck,
  } = useBetreiber()
  const { tseStatus, isPending: tseLoading } = useTSEStatus()
  const { kassenidentitaet } = useKassenidentitaet()

  const [formOffen, setFormOffen] = useState(false)
  const [copied, setCopied] = useState(false)

  const { loading: meldungLaeuft, run: runMeldung } = useActionSubmit({
    actionLabel: 'ELSTER-Meldung aktualisieren',
  })

  if (betreiberLoading) {
    return (
      <Card>
        <CardContent className="py-6">
          <p className="text-muted-foreground text-sm">Lade Einrichtung…</p>
        </CardContent>
      </Card>
    )
  }

  // Expliziter Ladefehler statt der leeren Checkliste — sonst wirkt der Verein
  // fälschlich unkonfiguriert und der Admin gibt seine Daten erneut ein.
  if (betreiberError) {
    return (
      <Card>
        <CardContent className="py-6">
          <LadefehlerAlert
            titel="Vereinsdaten konnten nicht geladen werden"
            onErneutVersuchen={() => void refetchBetreiber()}
          />
        </CardContent>
      </Card>
    )
  }

  const daten = betreiber ?? {
    vereinsname: '',
    strasse: '',
    plz: '',
    ort: '',
    steuernummer: null,
    ustId: null,
    elsterGemeldetAm: null,
  }

  const vereinsdatenOk = vereinsdatenErledigt(daten)
  const tseOk = !tseLoading && (tseStatus?.istKonfiguriert ?? false)
  const gemeldetAm = daten.elsterGemeldetAm
  const meldungOk = gemeldetAm !== null

  const erledigteSchritte = [vereinsdatenOk, tseOk, meldungOk].filter(
    Boolean,
  ).length

  const handleCopy = async () => {
    if (!kassenidentitaet) return
    await navigator.clipboard.writeText(kassenidentitaet.seriennummer)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  const handleSetMeldung = () =>
    void runMeldung(async () => {
      await setElsterMeldung()
      toast.success('Kassenmeldung als erledigt markiert.')
    })

  const handleResetMeldung = () =>
    void runMeldung(async () => {
      await nimmElsterMeldungZurueck()
      toast.success('Kassenmeldung zurückgesetzt.')
    })

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          Einrichtung — {erledigteSchritte} von 3 Schritten erledigt
        </CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">
        {/* 3 Spalten erst ab xl (~1280px), damit die Schritte im max-w-4xl-
            Container nicht zu schmal werden; darunter stapeln sie vertikal. */}
        <div className="grid grid-cols-1 gap-3 xl:grid-cols-3">
          <SchrittKarte
            nummer={1}
            titel="Vereinsdaten"
            erledigt={vereinsdatenOk}
          >
            <span className="text-sm leading-relaxed text-muted-foreground">
              {vereinsdatenOk
                ? `${daten.vereinsname}, ${daten.strasse}, ${daten.plz} ${daten.ort} — steht auf jedem Beleg.`
                : 'Name und Adresse des Vereins fehlen noch. Sie erscheinen auf jedem Kassenbeleg (§ 6 KassenSichV).'}
            </span>
            <button
              type="button"
              onClick={() => {
                setFormOffen((v) => !v)
              }}
              className="w-fit text-sm font-medium text-primary hover:underline"
            >
              {formOffen ? 'Schließen' : 'Bearbeiten'}
            </button>
          </SchrittKarte>

          {tseOk ? (
            <SchrittKarte nummer={2} titel="TSE aktiv" erledigt>
              <span className="text-sm leading-relaxed text-muted-foreground">
                {`Cloud-TSE verbunden${tseStatus?.umgebung ? ` (Umgebung ${tseStatus.umgebung})` : ''}. Signiert jeden Vorgang automatisch.`}
              </span>
            </SchrittKarte>
          ) : (
            <WarnKarte title="2 · TSE aktiv">
              <div className="flex flex-col items-start gap-1.5">
                <span className="text-muted-foreground">
                  Ohne aktive TSE darf nicht kassiert werden.
                </span>
                <Button asChild variant="outline" size="sm" className="w-fit">
                  <NavLink to="/admin/tse-einrichtung">TSE einrichten</NavLink>
                </Button>
              </div>
            </WarnKarte>
          )}

          {meldungOk ? (
            <SchrittKarte
              nummer={3}
              titel="Kasse beim Finanzamt melden"
              erledigt
            >
              <span className="text-sm leading-relaxed text-primary">
                Gemeldet am {formatDatum(gemeldetAm)}.
              </span>
              <button
                type="button"
                onClick={handleResetMeldung}
                disabled={meldungLaeuft}
                className="w-fit text-sm font-medium text-muted-foreground hover:underline disabled:opacity-50"
              >
                Zurücknehmen
              </button>
            </SchrittKarte>
          ) : (
            <WarnKarte title="3 · Kasse beim Finanzamt melden">
              <div className="flex flex-col gap-1.5">
                <span className="text-muted-foreground">
                  Noch offen — Frist: innerhalb 1 Monat nach Inbetriebnahme,
                  über ELSTER.{' '}
                  <span className="text-xs">(§ 146a Abs. 4 AO)</span>
                </span>
                {kassenidentitaet && (
                  // Eigenes, vollbreites Feld statt truncated <code>: die
                  // Seriennummer muss zum Abtippen in ELSTER vollständig lesbar
                  // sein (break-all bricht die UUID um, statt sie zu kürzen).
                  <div className="flex flex-col gap-1">
                    <span className="text-xs font-medium text-muted-foreground">
                      Seriennummer des elektronischen Aufzeichnungssystems
                    </span>
                    <div className="flex items-center gap-1.5">
                      <code className="min-w-0 flex-1 rounded-md border bg-background px-2 py-1 font-mono text-xs break-all">
                        {kassenidentitaet.seriennummer}
                      </code>
                      <Button
                        variant="outline"
                        size="icon"
                        className="size-7 shrink-0"
                        onClick={() => void handleCopy()}
                        aria-label="Seriennummer kopieren"
                      >
                        {copied ? (
                          <Check className="size-3.5" />
                        ) : (
                          <Copy className="size-3.5" />
                        )}
                      </Button>
                    </div>
                  </div>
                )}
                <div className="flex flex-wrap items-center gap-3">
                  <a
                    href={LEITFADEN_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm font-medium text-primary hover:underline"
                  >
                    Anleitung öffnen
                  </a>
                  <button
                    type="button"
                    onClick={handleSetMeldung}
                    disabled={meldungLaeuft}
                    className="text-sm font-medium text-muted-foreground hover:underline disabled:opacity-50"
                  >
                    Als erledigt markieren
                  </button>
                </div>
              </div>
            </WarnKarte>
          )}
        </div>

        {formOffen && (
          <div className="rounded-lg border p-4">
            <BetreiberForm
              initial={{
                vereinsname: daten.vereinsname,
                strasse: daten.strasse,
                plz: daten.plz,
                ort: daten.ort,
                steuernummer: daten.steuernummer,
                ustId: daten.ustId,
              }}
              onSave={async (b) => {
                await saveBetreiber(b)
                setFormOffen(false)
              }}
            />
          </div>
        )}
      </CardContent>
    </Card>
  )
}
