import {
  Ban,
  Calendar,
  ChartBar,
  ClipboardList,
  Loader2,
  TableIcon,
  Users,
} from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import { Field, FieldLabel } from '@/components/ui/field'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { ReportingBackend } from './ReportingBackend'
import type { Tagesabrechnung } from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

function todayStart(): string {
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return toDatetimeLocalString(now)
}

function nowLocal(): string {
  return toDatetimeLocalString(new Date())
}

function toDatetimeLocalString(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${String(date.getFullYear())}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function formatLocalTime(utcString: string): string {
  return new Date(utcString).toLocaleString('de-DE')
}

function pct(part: number, total: number): number {
  return total > 0 ? Math.round((part / total) * 100) : 0
}

export function TagesabrechnungPage() {
  const [von, setVon] = useState(todayStart)
  const [bis, setBis] = useState(nowLocal)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<Tagesabrechnung | null>(null)

  const auswerten = async () => {
    const vonUTC = new Date(von).toISOString()
    const bisUTC = new Date(bis).toISOString()

    if (vonUTC >= bisUTC) {
      toast.error('"Von" muss vor "Bis" liegen.')
      return
    }

    setLoading(true)
    try {
      const data = await reportingBackend.getTagesabrechnung(vonUTC, bisUTC)
      setResult(data)
    } catch {
      toast.error('Fehler beim Laden der Tagesabrechnung.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <h1 className="text-2xl font-bold">Tagesabrechnung</h1>

      {/* Filter bar */}
      <div className="mt-4 flex flex-wrap items-end gap-4">
        <Field>
          <FieldLabel htmlFor="von">Von</FieldLabel>
          <InputGroup>
            <InputGroupAddon align="inline-start">
              <Calendar className="size-4" />
            </InputGroupAddon>
            <InputGroupInput
              id="von"
              type="datetime-local"
              value={von}
              onChange={(e) => {
                setVon(e.target.value)
              }}
            />
          </InputGroup>
        </Field>

        <Field>
          <FieldLabel htmlFor="bis">Bis</FieldLabel>
          <InputGroup>
            <InputGroupAddon align="inline-start">
              <Calendar className="size-4" />
            </InputGroupAddon>
            <InputGroupInput
              id="bis"
              type="datetime-local"
              value={bis}
              onChange={(e) => {
                setBis(e.target.value)
              }}
            />
          </InputGroup>
        </Field>

        <Button onClick={() => void auswerten()} disabled={loading}>
          {loading ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <ClipboardList className="size-4" />
          )}
          Auswerten
        </Button>
      </div>

      {/* Results */}
      {result && (
        <div className="mt-6">
          <Tabs defaultValue="uebersicht">
            <TabsList variant="line">
              <TabsTrigger value="uebersicht">
                <ChartBar className="size-4" />
                Übersicht
              </TabsTrigger>
              <TabsTrigger value="servicekraefte">
                <Users className="size-4" />
                Servicekräfte
                {result.umsatzProServicekraft.length > 0 && (
                  <Badge variant="secondary" className="ml-1">
                    {result.umsatzProServicekraft.length}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value="tische">
                <TableIcon className="size-4" />
                Tische
                {result.umsatzProTisch.length > 0 && (
                  <Badge variant="secondary" className="ml-1">
                    {result.umsatzProTisch.length}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value="stornierungen">
                <Ban className="size-4" />
                Stornierungen
                {result.anzahlStornierungen > 0 && (
                  <Badge variant="destructive" className="ml-1">
                    {result.anzahlStornierungen}
                  </Badge>
                )}
              </TabsTrigger>
            </TabsList>

            {/* Übersicht */}
            <TabsContent value="uebersicht" className="mt-4">
              <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
                <SummaryCard
                  title="Gesamtumsatz"
                  value={`${formatCents(result.gesamtUmsatzCents)} €`}
                />
                <SummaryCard
                  title="Bestellungen"
                  value={String(result.anzahlBestellungen)}
                  sub={`${formatCents(result.gesamtBestellungenCents)} €`}
                />
                <SummaryCard
                  title="Stornierungen"
                  value={String(result.anzahlStornierungen)}
                  sub={`${formatCents(result.gesamtStornierungenCents)} €`}
                />
                <SummaryCard
                  title="Offene Saldi"
                  value={`${formatCents(result.offeneSaldiCents)} €`}
                />
              </div>
            </TabsContent>

            {/* Servicekräfte */}
            <TabsContent value="servicekraefte" className="mt-4">
              {result.umsatzProServicekraft.length === 0 ? (
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
                  {result.umsatzProServicekraft.map((sk) => (
                    <Item key={sk.userId} variant="outline" size="sm">
                      <ItemContent>
                        <ItemTitle>{sk.userName}</ItemTitle>
                        <Progress
                          value={pct(
                            sk.zahlungenCents,
                            result.gesamtUmsatzCents,
                          )}
                          className="mt-1 h-1.5"
                        />
                      </ItemContent>
                      <ItemActions>
                        <Badge variant="secondary">
                          {sk.anzahlZahlungen} Zahlungen
                        </Badge>
                        <span className="min-w-[6rem] text-right text-sm font-semibold">
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
              {result.umsatzProTisch.length === 0 ? (
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
                  {result.umsatzProTisch.map((t) => (
                    <Item key={t.tischId} variant="outline" size="sm">
                      <ItemContent>
                        <ItemTitle>{t.tischName}</ItemTitle>
                        <Progress
                          value={pct(
                            t.zahlungenCents,
                            result.gesamtUmsatzCents,
                          )}
                          className="mt-1 h-1.5"
                        />
                      </ItemContent>
                      <ItemActions>
                        <Badge variant="secondary">
                          {t.anzahlZahlungen} Zahlungen
                        </Badge>
                        <span className="min-w-[6rem] text-right text-sm font-semibold">
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
                              <Badge
                                variant="secondary"
                                className="ml-2 font-normal"
                              >
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
                                  {pos.varianteName
                                    ? ` (${pos.varianteName})`
                                    : ''}
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
        </div>
      )}
    </>
  )
}

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
