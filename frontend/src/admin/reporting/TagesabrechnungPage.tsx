import { ClipboardList, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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

      <div className="mt-4 flex flex-wrap items-end gap-4">
        <div>
          <label
            htmlFor="von"
            className="mb-1 block text-sm font-medium text-muted-foreground"
          >
            Von
          </label>
          <input
            id="von"
            type="datetime-local"
            value={von}
            onChange={(e) => {
              setVon(e.target.value)
            }}
            className="rounded-md border px-3 py-2 text-sm"
          />
        </div>
        <div>
          <label
            htmlFor="bis"
            className="mb-1 block text-sm font-medium text-muted-foreground"
          >
            Bis
          </label>
          <input
            id="bis"
            type="datetime-local"
            value={bis}
            onChange={(e) => {
              setBis(e.target.value)
            }}
            className="rounded-md border px-3 py-2 text-sm"
          />
        </div>
        <Button onClick={() => void auswerten()} disabled={loading}>
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <ClipboardList className="h-4 w-4" />
          )}
          Auswerten
        </Button>
      </div>

      {result && (
        <div className="mt-6 space-y-6">
          {/* Summary Cards */}
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            <SummaryCard
              title="Gesamtumsatz"
              value={`${formatCents(result.gesamtUmsatzCents)} €`}
            />
            <SummaryCard
              title="Bestellungen"
              value={`${String(result.anzahlBestellungen)} (${formatCents(result.gesamtBestellungenCents)} €)`}
            />
            <SummaryCard
              title="Stornierungen"
              value={`${String(result.anzahlStornierungen)} (${formatCents(result.gesamtStornierungenCents)} €)`}
            />
            <SummaryCard
              title="Offene Saldi"
              value={`${formatCents(result.offeneSaldiCents)} €`}
            />
          </div>

          {/* Umsatz pro Servicekraft */}
          <Card>
            <CardHeader>
              <CardTitle>Umsatz pro Servicekraft</CardTitle>
            </CardHeader>
            <CardContent>
              {result.umsatzProServicekraft.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  Keine Zahlungen im Zeitraum.
                </p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-left text-muted-foreground">
                        <th className="pb-2 pr-4">Servicekraft</th>
                        <th className="pb-2 pr-4 text-right">Zahlungen</th>
                        <th className="pb-2 text-right">Anzahl</th>
                      </tr>
                    </thead>
                    <tbody>
                      {result.umsatzProServicekraft.map((sk) => (
                        <tr key={sk.userId} className="border-b last:border-0">
                          <td className="py-2 pr-4">{sk.userName}</td>
                          <td className="py-2 pr-4 text-right">
                            {formatCents(sk.zahlungenCents)} €
                          </td>
                          <td className="py-2 text-right">
                            {sk.anzahlZahlungen}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Stornierungen */}
          <Card>
            <CardHeader>
              <CardTitle>Stornierungen</CardTitle>
            </CardHeader>
            <CardContent>
              {result.stornierungen.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  Keine Stornierungen im Zeitraum.
                </p>
              ) : (
                <div className="space-y-4">
                  {result.stornierungen.map((storno) => (
                    <div
                      key={`${storno.zeitpunkt}-${String(storno.tischId)}-${String(storno.userId)}`}
                      className="rounded-md border p-3"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-2">
                        <div>
                          <p className="font-medium">
                            {storno.tischName} — {storno.userName}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {formatLocalTime(storno.zeitpunkt)}
                          </p>
                        </div>
                        <p className="font-semibold">
                          {formatCents(storno.betragCents)} €
                        </p>
                      </div>
                      {storno.kommentar && (
                        <p className="mt-1 text-sm italic text-muted-foreground">
                          {storno.kommentar}
                        </p>
                      )}
                      {storno.positionen.length > 0 && (
                        <ul className="mt-2 space-y-1 text-sm">
                          {storno.positionen.map((pos) => (
                            <li
                              key={`${pos.produktName}-${pos.varianteName}`}
                              className="flex justify-between text-muted-foreground"
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
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </>
  )
}

function SummaryCard({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xl font-bold">{value}</p>
      </CardContent>
    </Card>
  )
}
