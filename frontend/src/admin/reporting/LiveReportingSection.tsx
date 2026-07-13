import {
  Ban,
  ChevronDown,
  LayoutDashboard,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { type ReactNode, useState } from 'react'
import { NavLink } from 'react-router'

import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { formatCents } from '@/lib/utils'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { StatusDot } from '../components/StatusDot'
import { StornoItem } from './StornoItem'
import { StornoAggregat } from './StornoServicekraft'
import { SummaryCard } from './SummaryCard'
import type { LiveReportingData } from './types'
import { formatBediener, formatDatum, formatStand } from './utils'

// Nach fünf Einträgen wird die Liste offener Tische gekürzt; „Alle n anzeigen"
// blendet den Rest ein (Design-Handoff 1a).
const OFFENE_TISCHE_VORSCHAU = 5

export function LiveReportingSection({
  liveData,
  loading,
  dataUpdatedAt,
  onRefresh,
  statusZeile,
}: {
  liveData: LiveReportingData | null
  loading: boolean
  dataUpdatedAt: number
  onRefresh: () => void
  statusZeile?: ReactNode
}) {
  const [tischeAusgeklappt, setTischeAusgeklappt] = useState(false)

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (liveData === null) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <LayoutDashboard />
          </EmptyMedia>
          <EmptyTitle>Keine Kassensitzung geöffnet</EmptyTitle>
          <EmptyDescription>
            <NavLink to="/admin/kasse" className="underline underline-offset-4">
              Zur Kassensitzungs-Seite
            </NavLink>
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    )
  }

  const summary = liveData.summary
  const servicekraefte = liveData.breakdowns.servicekraefte
  const stornierungenProServicekraft =
    liveData.breakdowns.stornierungenProServicekraft
  const stornoAnzahlByUserId = new Map(
    stornierungenProServicekraft.map((s) => [s.userId, s.anzahlStornierungen]),
  )

  const offeneTische = liveData.offeneTische
  const sichtbareTische = tischeAusgeklappt
    ? offeneTische
    : offeneTische.slice(0, OFFENE_TISCHE_VORSCHAU)

  return (
    <div data-testid="live-reporting-section" className="space-y-6">
      <AdminPageHeader
        titel="Übersicht"
        unterzeile={
          <>
            <strong>{formatDatum(liveData.datum)}</strong>{' '}
            {liveData.bezeichnung}
          </>
        }
        aktionen={
          <>
            <span className="inline-flex items-center gap-1.5 whitespace-nowrap text-xs text-muted-foreground">
              <StatusDot zustand="ok" label="Live" />
              Live · aktualisiert {formatStand(dataUpdatedAt)}
            </span>
            <Button variant="outline" size="sm" onClick={onRefresh}>
              <RefreshCw className="size-4" />
              Jetzt
            </Button>
          </>
        }
      />

      {statusZeile}

      {/* Kennzahlen: Hero-Karte „Kassierter Umsatz" plus vier Nebenkarten */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        <div className="col-span-2 flex flex-col gap-1.5 rounded-xl bg-muted/60 p-5 lg:col-span-1">
          <span className="text-sm font-medium text-muted-foreground">
            Kassierter Umsatz
          </span>
          <span className="text-3xl font-extrabold tracking-tight">
            {formatCents(summary.gesamtUmsatzCents)} €
          </span>
          <span className="text-xs text-muted-foreground">
            bereits bezahlt, Stornos abgezogen
          </span>
        </div>
        <SummaryCard
          title="Noch offen"
          value={`${formatCents(liveData.offeneSaldiCents)} €`}
          sub={`auf ${String(offeneTische.length)} Tischen bestellt, noch nicht bezahlt`}
        />
        <SummaryCard
          title="Bestellt gesamt"
          value={`${formatCents(summary.gesamtBestellungenCents)} €`}
          sub="bezahlt + offen zusammen"
        />
        <SummaryCard
          title="Direktverkauf"
          value={`${formatCents(summary.direktverkaufUmsatzCents)} €`}
          sub={`${String(summary.anzahlDirektverkaeufe)} Verkäufe ohne Tisch`}
        />
        <SummaryCard
          title="Storniert"
          value={`${formatCents(summary.gesamtStornierungenCents)} €`}
          sub={`${String(summary.anzahlStornierungen)} Stornierungen`}
        />
      </div>

      {/* Offene Tische und Team nebeneinander */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-xl border p-5">
          <div className="mb-3 flex items-baseline justify-between">
            <h2 className="text-base font-semibold">Offene Tische</h2>
            <span className="text-sm text-muted-foreground">
              {offeneTische.length} Tische ·{' '}
              {formatCents(liveData.offeneSaldiCents)} €
            </span>
          </div>
          {offeneTische.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Alle Tische sind abgerechnet.
            </p>
          ) : (
            <div className="flex flex-col">
              {sichtbareTische.map((t) => (
                <div
                  key={t.tischId}
                  className="flex items-center justify-between border-b py-2 text-sm last:border-b-0"
                >
                  <span>{t.tischName}</span>
                  <span className="font-semibold">
                    {formatCents(t.saldoCents)} €
                  </span>
                </div>
              ))}
              {!tischeAusgeklappt &&
                offeneTische.length > OFFENE_TISCHE_VORSCHAU && (
                  <button
                    type="button"
                    onClick={() => {
                      setTischeAusgeklappt(true)
                    }}
                    className="pt-2 text-left text-sm font-medium text-primary hover:underline"
                  >
                    Alle {offeneTische.length} anzeigen
                  </button>
                )}
            </div>
          )}
        </div>

        <div className="rounded-xl border p-5">
          <div className="mb-3 flex items-baseline justify-between">
            <h2 className="text-base font-semibold">Team</h2>
            <span className="text-sm text-muted-foreground">
              {servicekraefte.length} Servicekräfte aktiv
            </span>
          </div>
          {servicekraefte.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Noch keine Zahlungen und keine offenen Tische.
            </p>
          ) : (
            <div className="flex flex-col">
              {servicekraefte.map((sk) => {
                const tischNamen = sk.offeneTische
                  .map((t) => t.tischName)
                  .join(', ')
                const stornoAnzahl = stornoAnzahlByUserId.get(sk.userId) ?? 0
                return (
                  <div
                    key={sk.userId}
                    className="flex items-center gap-3 border-b py-2.5 last:border-b-0"
                  >
                    <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <span className="text-sm font-medium">
                        {formatBediener(sk.userName, sk.name)}
                      </span>
                      {sk.erledigt ? (
                        <span className="text-xs font-medium text-primary">
                          Alles abgerechnet
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground">
                          Offen:{' '}
                          <span className="whitespace-nowrap font-medium">
                            {formatCents(sk.offenCents)} €
                          </span>{' '}
                          auf {sk.offeneTische.length}{' '}
                          {sk.offeneTische.length === 1 ? 'Tisch' : 'Tischen'}
                          {tischNamen && ` (${tischNamen})`}
                          {stornoAnzahl > 0 && (
                            <>
                              {' · '}
                              <span className="font-medium text-destructive">
                                {stornoAnzahl} Storno
                                {stornoAnzahl === 1 ? '' : 's'}
                              </span>
                            </>
                          )}
                        </span>
                      )}
                    </div>
                    <span className="whitespace-nowrap text-sm font-semibold">
                      {formatCents(sk.zahlungenCents)} €
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      {/* Stornierungen: eingeklappte Zeile, Aufklappen zeigt die Detail-Liste */}
      {liveData.stornierungen.length > 0 && (
        <Collapsible>
          <div className="rounded-lg border">
            <div className="flex items-center gap-3 px-4 py-3">
              <Ban className="size-4 shrink-0 text-destructive" aria-hidden />
              <div className="min-w-0 flex-1">
                <p className="text-sm">
                  <strong>
                    {summary.anzahlStornierungen} Stornierung
                    {summary.anzahlStornierungen === 1 ? '' : 'en'}
                  </strong>{' '}
                  · {formatCents(summary.gesamtStornierungenCents)} €
                </p>
                {stornierungenProServicekraft.length > 0 && (
                  <StornoAggregat
                    eintraege={stornierungenProServicekraft}
                    className="mb-0 mt-0.5"
                  />
                )}
              </div>
              <CollapsibleTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="group/storno shrink-0"
                >
                  Details
                  <ChevronDown className="size-4 transition-transform group-data-[state=open]/storno:rotate-180" />
                </Button>
              </CollapsibleTrigger>
            </div>
            <CollapsibleContent>
              <div className="space-y-2 border-t px-4 py-3">
                {liveData.stornierungen.map((s) => (
                  <StornoItem
                    key={`${s.zeitpunkt}-${String(s.tischId)}-${String(s.userId)}`}
                    storno={s}
                  />
                ))}
              </div>
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}
    </div>
  )
}
