import {
  Ban,
  ChartBar,
  CheckCircle2,
  LayoutDashboard,
  TableIcon,
  Users,
} from 'lucide-react'
import { NavLink } from 'react-router'

import { Badge } from '@/components/ui/badge'
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
import { Progress } from '@/components/ui/progress'
import {
  ScrollableTabsList,
  Tabs,
  TabsContent,
  TabsTrigger,
} from '@/components/ui/tabs'
import { formatCents, formatPositionName } from '@/lib/utils'

import { SummaryCard } from './SummaryCard'
import type { LiveReportingData } from './types'
import {
  formatBediener,
  formatDatum,
  formatLocalTime,
  formatOffeneArbeit,
  pct,
} from './utils'

export function LiveReportingSection({
  liveData,
  loading,
}: {
  liveData: LiveReportingData | null
  loading: boolean
}) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Progress className="w-1/2" />
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
  const breakdowns = liveData.breakdowns

  return (
    <div data-testid="live-reporting-section" className="space-y-4">
      <div className="flex items-baseline gap-2">
        <h1 className="text-2xl font-bold">Live-Dashboard</h1>
        <p className="text-muted-foreground">
          <strong>{formatDatum(liveData.datum)}</strong> {liveData.bezeichnung}
        </p>
      </div>

      <Tabs defaultValue="uebersicht">
        <ScrollableTabsList variant="line">
          <TabsTrigger value="uebersicht">
            <ChartBar className="size-4" />
            Übersicht
          </TabsTrigger>
          <TabsTrigger value="servicekraefte">
            <Users className="size-4" />
            Servicekräfte
            {breakdowns.servicekraefte.length > 0 && (
              <Badge variant="secondary" className="ml-1">
                {breakdowns.servicekraefte.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="tische">
            <TableIcon className="size-4" />
            Tische
            {breakdowns.umsatzProTisch.length > 0 && (
              <Badge variant="secondary" className="ml-1">
                {breakdowns.umsatzProTisch.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="stornierungen">
            <Ban className="size-4" />
            Stornierungen
            {summary.anzahlStornierungen > 0 && (
              <Badge variant="destructive" className="ml-1">
                {summary.anzahlStornierungen}
              </Badge>
            )}
          </TabsTrigger>
        </ScrollableTabsList>

        {/* Übersicht */}
        <TabsContent value="uebersicht" className="mt-4 space-y-6">
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <SummaryCard
              title="Bestellungen"
              value={String(summary.anzahlBestellungen)}
              sub={`${formatCents(summary.gesamtBestellungenCents)} €`}
            />
            <SummaryCard
              title="Direktverkauf"
              value={String(summary.anzahlDirektverkaeufe)}
              sub={`${formatCents(summary.direktverkaufUmsatzCents)} €`}
            />
            <SummaryCard
              title="Gesamtumsatz"
              value={`${formatCents(summary.gesamtUmsatzCents)} €`}
              sub="Kassierungen − Warenrücknahmen"
            />
            <SummaryCard
              title="Offene Saldi"
              value={`${formatCents(liveData.offeneSaldiCents)} €`}
              sub={`${String(liveData.offeneTische.length)} offene Tische`}
            />
            <SummaryCard
              title="Stornierungen"
              value={String(summary.anzahlStornierungen)}
              sub={`${formatCents(summary.gesamtStornierungenCents)} €`}
            />
          </div>

          {liveData.offeneTische.length > 0 && (
            <div>
              <h3 className="mb-3 text-sm font-medium text-muted-foreground">
                Offene Tische
              </h3>
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
        </TabsContent>

        {/* Servicekräfte */}
        <TabsContent value="servicekraefte" className="mt-4">
          {breakdowns.servicekraefte.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Users />
                </EmptyMedia>
                <EmptyTitle>Keine Servicekräfte aktiv</EmptyTitle>
                <EmptyDescription>
                  Noch keine Zahlungen und keine offenen Tische.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <ItemGroup>
              {breakdowns.servicekraefte.map((sk) => (
                <Item key={sk.userId} variant="outline" size="sm">
                  <ItemContent>
                    <ItemTitle>
                      {formatBediener(sk.userName, sk.name)}
                    </ItemTitle>
                    <Progress
                      value={pct(sk.zahlungenCents, summary.gesamtUmsatzCents)}
                      className="mt-1 h-1.5"
                    />
                    {sk.erledigt ? (
                      <span className="mt-1.5 inline-flex items-center gap-1 text-xs font-medium text-green-700">
                        <CheckCircle2 className="size-3.5" />
                        Fertig
                      </span>
                    ) : (
                      <div className="mt-1.5 flex flex-wrap gap-1.5">
                        {sk.offeneTische.map((t) => (
                          <span
                            key={t.tischId}
                            className="rounded-md border px-2 py-0.5 text-xs"
                          >
                            <span className="font-medium">{t.tischName}</span>
                            <span className="text-muted-foreground">
                              {' · '}
                              {formatOffeneArbeit(t)}
                            </span>
                          </span>
                        ))}
                      </div>
                    )}
                  </ItemContent>
                  <ItemActions>
                    <Badge variant="secondary">
                      {sk.anzahlZahlungen} Zahlungen
                    </Badge>
                    <span className="min-w-24 text-right text-sm font-semibold">
                      {formatCents(sk.zahlungenCents)} €
                    </span>
                  </ItemActions>
                </Item>
              ))}
            </ItemGroup>
          )}
        </TabsContent>

        {/* Tische */}
        <TabsContent value="tische" className="mt-4">
          {breakdowns.umsatzProTisch.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <TableIcon />
                </EmptyMedia>
                <EmptyTitle>Keine Tischzahlungen</EmptyTitle>
                <EmptyDescription>
                  Keine Tischzahlungen im gewählten Zeitraum.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <ItemGroup>
              {breakdowns.umsatzProTisch.map((t) => (
                <Item key={t.tischId} variant="outline" size="sm">
                  <ItemContent>
                    <ItemTitle>{t.tischName}</ItemTitle>
                    <Progress
                      value={pct(t.zahlungenCents, summary.gesamtUmsatzCents)}
                      className="mt-1 h-1.5"
                    />
                  </ItemContent>
                  <ItemActions>
                    <Badge variant="secondary">
                      {t.anzahlZahlungen} Zahlungen
                    </Badge>
                    <span className="min-w-24 text-right text-sm font-semibold">
                      {formatCents(t.zahlungenCents)} €
                    </span>
                  </ItemActions>
                </Item>
              ))}
            </ItemGroup>
          )}
        </TabsContent>

        {/* Stornierungen */}
        <TabsContent value="stornierungen" className="mt-4">
          {liveData.stornierungen.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Ban />
                </EmptyMedia>
                <EmptyTitle>Keine Stornierungen</EmptyTitle>
                <EmptyDescription>
                  Keine Stornierungen in dieser Kassensitzung.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <ItemGroup>
              {liveData.stornierungen.map((s) => (
                <Item
                  key={`${s.zeitpunkt}-${String(s.tischId)}`}
                  variant="outline"
                  size="sm"
                >
                  <ItemContent>
                    <ItemTitle>
                      {s.quelle === 'direktverkauf'
                        ? 'Direktverkauf'
                        : s.tischName}{' '}
                      · {formatBediener(s.userName, s.name)}
                    </ItemTitle>
                    <p className="text-xs text-muted-foreground">
                      {formatLocalTime(s.zeitpunkt)}
                      {s.kommentar ? ` — ${s.kommentar}` : ''}
                    </p>
                    {s.positionen.length > 0 && (
                      <ul className="mt-1">
                        {s.positionen.map((pos) => (
                          <li
                            key={`${pos.produktName}-${pos.varianteName}`}
                            className="flex justify-between text-sm text-muted-foreground"
                          >
                            <span>
                              {pos.menge}×{' '}
                              {formatPositionName(
                                pos.produktName,
                                pos.varianteName,
                              )}
                            </span>
                            <span>{formatCents(pos.einzelpreisCents)} €</span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </ItemContent>
                  <ItemActions>
                    <Badge variant={s.barRueckgabe ? 'outline' : 'secondary'}>
                      {s.barRueckgabe ? 'Bar-Rückgabe' : 'Geldneutral'}
                    </Badge>
                    <span className="text-sm font-semibold">
                      {formatCents(s.betragCents)} €
                    </span>
                  </ItemActions>
                </Item>
              ))}
            </ItemGroup>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
