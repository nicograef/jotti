import { Ban, ChartBar, InfoIcon, TableIcon, Users } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatCents } from '@/lib/utils'

import type { Reporting } from './types'

function SummaryCard({
  title,
  value,
  sub,
}: {
  title: string
  value: string
  sub?: string
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xl font-bold">{value}</p>
        {sub && <p className="mt-0.5 text-sm text-muted-foreground">{sub}</p>}
      </CardContent>
    </Card>
  )
}

function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0
}

function formatLocalTime(utcString: string): string {
  return new Date(utcString).toLocaleString('de-DE')
}

export function ReportingResults({ result }: { result: Reporting }) {
  const summary = result.summary
  const breakdowns = result.breakdowns

  return (
    <Tabs defaultValue="uebersicht">
      <div className="overflow-x-auto -mx-4 px-4">
        <TabsList variant="line">
          <TabsTrigger value="uebersicht">
            <ChartBar className="size-4" />
            Übersicht
          </TabsTrigger>
          <TabsTrigger value="servicekraefte">
            <Users className="size-4" />
            Servicekräfte
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <InfoIcon className="size-3 cursor-help opacity-60" />
                </TooltipTrigger>
                <TooltipContent>
                  Summe der registrierten Zahlungen pro Servicekraft.
                  <br />
                  Rückzahlungen werden aktuell nicht berücksichtigt.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
            {breakdowns.umsatzProServicekraft.length > 0 && (
              <Badge variant="secondary" className="ml-1">
                {breakdowns.umsatzProServicekraft.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="tische">
            <TableIcon className="size-4" />
            Tische
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <InfoIcon className="size-3 cursor-help opacity-60" />
                </TooltipTrigger>
                <TooltipContent>
                  Summe der registrierten Zahlungen pro Tisch.
                  <br />
                  Rückzahlungen werden aktuell nicht berücksichtigt.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
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
        </TabsList>
      </div>

      {/* Übersicht */}
      <TabsContent value="uebersicht" className="mt-4">
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-1 text-sm text-muted-foreground">
                Gesamtumsatz
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <InfoIcon className="size-3.5 cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      Summe aller registrierten Zahlungen im Zeitraum.
                      <br />
                      Rückzahlungen werden aktuell nicht berücksichtigt.
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xl font-bold">
                {formatCents(summary.gesamtUmsatzCents)} €
              </p>
            </CardContent>
          </Card>
          <SummaryCard
            title="Offene Tische"
            value={String(summary.anzahlOffeneTische)}
            sub="Aktueller Stand"
          />
          <SummaryCard
            title="Bestellungen"
            value={String(summary.anzahlBestellungen)}
            sub={`${formatCents(summary.gesamtBestellungenCents)} €`}
          />
          <SummaryCard
            title="Stornierungen"
            value={String(summary.anzahlStornierungen)}
            sub={`${formatCents(summary.gesamtStornierungenCents)} €`}
          />
          <SummaryCard
            title="Offene Saldi"
            value={`${formatCents(summary.offeneSaldiCents)} €`}
            sub="Aktueller Stand"
          />
        </div>
      </TabsContent>

      {/* Servicekräfte */}
      <TabsContent value="servicekraefte" className="mt-4">
        {breakdowns.umsatzProServicekraft.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users />
              </EmptyMedia>
              <EmptyTitle>Keine Zahlungen</EmptyTitle>
              <EmptyDescription>
                Keine Zahlungen im gewählten Zeitraum.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ItemGroup>
            {breakdowns.umsatzProServicekraft.map((sk) => (
              <Item key={sk.userId} variant="outline" size="sm">
                <ItemContent>
                  <ItemTitle>{sk.userName}</ItemTitle>
                  <Progress
                    value={pct(sk.zahlungenCents, summary.gesamtUmsatzCents)}
                    className="mt-1 h-1.5"
                  />
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
        {result.stornierungen.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Ban />
              </EmptyMedia>
              <EmptyTitle>Keine Stornierungen</EmptyTitle>
              <EmptyDescription>
                Keine Stornierungen im gewählten Zeitraum.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ItemGroup>
            {result.stornierungen.map((storno) => (
              <Item
                key={`${storno.zeitpunkt}-${String(storno.tischId)}-${String(storno.userId)}`}
                variant="outline"
              >
                <ItemContent>
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <ItemTitle>
                        {storno.tischName}
                        <Badge variant="secondary" className="ml-2 font-normal">
                          {storno.userName}
                        </Badge>
                      </ItemTitle>
                      <p className="mt-0.5 text-sm text-muted-foreground">
                        {formatLocalTime(storno.zeitpunkt)}
                      </p>
                    </div>
                    <Badge variant="destructive">
                      {formatCents(storno.betragCents)} €
                    </Badge>
                  </div>
                  {storno.kommentar && (
                    <p className="mt-1 text-sm italic text-muted-foreground">
                      {storno.kommentar}
                    </p>
                  )}
                  {storno.positionen.length > 0 && (
                    <ul className="mt-2 space-y-0.5">
                      {storno.positionen.map((pos) => (
                        <li
                          key={`${pos.produktName}-${pos.varianteName}`}
                          className="flex justify-between text-sm text-muted-foreground"
                        >
                          <span>
                            {pos.menge}× {pos.produktName}
                            {pos.varianteName ? ` (${pos.varianteName})` : ''}
                          </span>
                          <span>{formatCents(pos.einzelpreis)} €</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </ItemContent>
              </Item>
            ))}
          </ItemGroup>
        )}
      </TabsContent>
    </Tabs>
  )
}
