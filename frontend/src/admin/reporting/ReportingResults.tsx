import { Ban, ChartBar, Loader2, Users } from 'lucide-react'

import { STEUERSATZ_LABEL } from '@/admin/products/Produkt'
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
import {
  ScrollableTabsList,
  Tabs,
  TabsContent,
  TabsTrigger,
} from '@/components/ui/tabs'
import { formatCents } from '@/lib/utils'

import { StornoItem } from './StornoItem'
import { StornoAggregat, StornoMarker } from './StornoServicekraft'
import { SummaryCard } from './SummaryCard'
import type { ReportingData } from './types'
import { formatBediener } from './utils'

export function ReportingResults({
  result,
  loading,
}: {
  result: ReportingData
  loading: boolean
}) {
  const summary = result.summary
  const breakdowns = result.breakdowns
  const stornoAnzahlByUserId = new Map(
    breakdowns.stornierungenProServicekraft.map((s) => [
      s.userId,
      s.anzahlStornierungen,
    ]),
  )

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <Tabs defaultValue="uebersicht">
      <ScrollableTabsList variant="line">
        <TabsTrigger value="uebersicht">
          <ChartBar className="size-4" />
          Übersicht
        </TabsTrigger>
        <TabsTrigger value="servicekraefte">
          <Users className="size-4" />
          Servicekräfte
          {breakdowns.umsatzProServicekraft.length > 0 && (
            <Badge variant="secondary" className="ml-1">
              {breakdowns.umsatzProServicekraft.length}
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
      <TabsContent value="uebersicht" className="mt-4">
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <SummaryCard
            title="Gesamtumsatz"
            value={`${formatCents(summary.gesamtUmsatzCents)} €`}
            sub="kassiert, abzüglich Warenrücknahmen"
          />
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
            title="Stornierungen"
            value={String(summary.anzahlStornierungen)}
            sub={`${formatCents(summary.gesamtStornierungenCents)} €`}
          />
        </div>

        <Card className="mt-4">
          <CardHeader>
            <CardTitle className="text-base">Umsatz nach Steuersatz</CardTitle>
          </CardHeader>
          <CardContent>
            {result.umsatzProSteuersatz.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                Keine steuerrelevanten Umsätze in dieser Kassensitzung.
              </p>
            ) : (
              <div className="space-y-3">
                {result.umsatzProSteuersatz.map((umsatz) => (
                  <div key={umsatz.satz} className="rounded-md border p-3">
                    <p className="font-medium">
                      {STEUERSATZ_LABEL[umsatz.satz]}
                    </p>
                    <dl className="mt-2 space-y-1 text-sm">
                      <div className="flex justify-between gap-2">
                        <dt className="text-muted-foreground">Brutto</dt>
                        <dd className="whitespace-nowrap font-medium">
                          {formatCents(umsatz.bruttoCents)} €
                        </dd>
                      </div>
                      <div className="flex justify-between gap-2">
                        <dt className="text-muted-foreground">Netto</dt>
                        <dd className="whitespace-nowrap">
                          {formatCents(umsatz.nettoCents)} €
                        </dd>
                      </div>
                      <div className="flex justify-between gap-2">
                        <dt className="text-muted-foreground">Steuer</dt>
                        <dd className="whitespace-nowrap">
                          {formatCents(umsatz.steuerCents)} €
                        </dd>
                      </div>
                    </dl>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
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
                Keine Zahlungen in dieser Kassensitzung.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ItemGroup>
            {breakdowns.umsatzProServicekraft.map((sk) => {
              const stornoAnzahl = stornoAnzahlByUserId.get(sk.userId) ?? 0
              return (
                <Item key={sk.userId} variant="outline" size="sm">
                  <ItemContent>
                    <ItemTitle>
                      {formatBediener(sk.userName, sk.name)}
                    </ItemTitle>
                    {stornoAnzahl > 0 && <StornoMarker anzahl={stornoAnzahl} />}
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
                Keine Stornierungen in dieser Kassensitzung.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <>
            {breakdowns.stornierungenProServicekraft.length > 0 && (
              <StornoAggregat
                eintraege={breakdowns.stornierungenProServicekraft}
              />
            )}
            <ItemGroup>
              {result.stornierungen.map((storno) => (
                <StornoItem
                  key={`${storno.zeitpunkt}-${String(storno.tischId)}-${String(storno.userId)}`}
                  storno={storno}
                />
              ))}
            </ItemGroup>
          </>
        )}
      </TabsContent>
    </Tabs>
  )
}
