import { CheckCircle2, LayoutDashboard, Loader2, RefreshCw } from 'lucide-react'
import { NavLink } from 'react-router'

import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { formatCents } from '@/lib/utils'

import { StornoItem } from './StornoItem'
import { StornoAggregat, StornoMarker } from './StornoServicekraft'
import { SummaryCard } from './SummaryCard'
import type { LiveReportingData } from './types'
import { formatBediener, formatDatum, formatStand } from './utils'

export function LiveReportingSection({
  liveData,
  loading,
  dataUpdatedAt,
  onRefresh,
}: {
  liveData: LiveReportingData | null
  loading: boolean
  dataUpdatedAt: number
  onRefresh: () => void
}) {
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

  return (
    <div data-testid="live-reporting-section" className="space-y-6">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h1 className="text-2xl font-bold">Live-Dashboard</h1>
          <p className="text-muted-foreground">
            <strong>{formatDatum(liveData.datum)}</strong>{' '}
            {liveData.bezeichnung}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <span className="whitespace-nowrap text-xs text-muted-foreground">
            Stand {formatStand(dataUpdatedAt)}
          </span>
          <Button
            variant="outline"
            size="icon"
            onClick={onRefresh}
            aria-label="Aktualisieren"
          >
            <RefreshCw className="size-4" />
          </Button>
        </div>
      </div>

      {/* Kennzahlen (kanonische Reihenfolge) */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <SummaryCard
          title="Gesamtumsatz"
          value={`${formatCents(summary.gesamtUmsatzCents)} €`}
          sub="kassiert, abzüglich Warenrücknahmen"
        />
        <SummaryCard
          title="Offene Saldi"
          value={`${formatCents(liveData.offeneSaldiCents)} €`}
          sub={`${String(liveData.offeneTische.length)} offene Tische`}
        />
        <SummaryCard
          title="Bestellungen"
          value={`${formatCents(summary.gesamtBestellungenCents)} €`}
          sub="Bestellwert, inkl. noch nicht kassiert"
        />
        <SummaryCard
          title="Direktverkauf"
          value={String(summary.anzahlDirektverkaeufe)}
          sub={`${formatCents(summary.direktverkaufUmsatzCents)} €`}
        />
        <SummaryCard
          title="Stornierungen"
          value={String(summary.anzahlStornierungen)}
          sub={`${formatCents(summary.gesamtStornierungenCents)} €`}
        />
      </div>

      {/* Offene Tische (Backend-Sortierung nach Saldo absteigend) */}
      {liveData.offeneTische.length > 0 && (
        <div>
          <h2 className="mb-3 text-sm font-medium text-muted-foreground">
            Offene Tische
          </h2>
          <div className="flex flex-wrap gap-2">
            {liveData.offeneTische.map((t) => (
              <div
                key={t.tischId}
                className="flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm"
              >
                <span>{t.tischName}</span>
                <span className="font-semibold">
                  {formatCents(t.saldoCents)} €
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Stornierungen */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-muted-foreground">
          Stornierungen
        </h2>
        {liveData.stornierungen.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Keine Stornierungen in dieser Kassensitzung.
          </p>
        ) : (
          <>
            {stornierungenProServicekraft.length > 0 && (
              <StornoAggregat eintraege={stornierungenProServicekraft} />
            )}
            <ItemGroup>
              {liveData.stornierungen.map((s) => (
                <StornoItem
                  key={`${s.zeitpunkt}-${String(s.tischId)}-${String(s.userId)}`}
                  storno={s}
                />
              ))}
            </ItemGroup>
          </>
        )}
      </div>

      {/* Servicekräfte */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-muted-foreground">
          Servicekräfte
        </h2>
        {servicekraefte.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Noch keine Zahlungen und keine offenen Tische.
          </p>
        ) : (
          <ItemGroup>
            {servicekraefte.map((sk) => {
              const offenCents = sk.offeneTische.reduce(
                (summe, t) => summe + t.offenCents,
                0,
              )
              const tischNamen = sk.offeneTische
                .map((t) => t.tischName)
                .join(', ')
              const stornoAnzahl = stornoAnzahlByUserId.get(sk.userId) ?? 0
              return (
                <Item key={sk.userId} variant="outline" size="sm">
                  <ItemContent>
                    <ItemTitle>
                      {formatBediener(sk.userName, sk.name)}
                    </ItemTitle>
                    {stornoAnzahl > 0 && <StornoMarker anzahl={stornoAnzahl} />}
                    {sk.erledigt ? (
                      <span className="mt-1.5 inline-flex items-center gap-1 text-xs font-medium text-primary">
                        <CheckCircle2 className="size-3.5" />
                        Fertig
                      </span>
                    ) : (
                      <span className="mt-1.5 text-xs text-muted-foreground">
                        Offen:{' '}
                        <span className="whitespace-nowrap font-medium">
                          {formatCents(offenCents)} €
                        </span>
                        {' · '}
                        {tischNamen}
                      </span>
                    )}
                  </ItemContent>
                  <ItemActions>
                    <span className="min-w-24 text-right text-sm font-semibold">
                      {formatCents(sk.zahlungenCents)} €
                    </span>
                  </ItemActions>
                </Item>
              )
            })}
          </ItemGroup>
        )}
      </div>
    </div>
  )
}
