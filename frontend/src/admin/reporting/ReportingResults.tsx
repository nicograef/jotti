import { Loader2, Printer } from 'lucide-react'

import { STEUERSATZ_LABEL } from '@/admin/products/Produkt'
import { Button } from '@/components/ui/button'
import { formatCents } from '@/lib/utils'

import { StornoItem } from './StornoItem'
import { StornoMarker } from './StornoServicekraft'
import type { AbgeschlosseneSitzung, ReportingData } from './types'
import { formatBediener, formatDatumLang, formatLocalTime } from './utils'

// Berichtskopf-Zeile: Datum, Eröffnungs-/Abschlusszeit, abschließender Benutzer
// und Kassensturz-Differenz — rein aus den vom Backend projizierten Metadaten.
function BerichtsMeta({
  datum,
  result,
}: {
  datum: string
  result: ReportingData
}) {
  const { metadaten } = result
  const teile: string[] = [formatDatumLang(datum)]
  if (metadaten.eroeffnetAm) {
    teile.push(`eröffnet ${formatLocalTime(metadaten.eroeffnetAm)}`)
  }
  if (metadaten.abgeschlossenAm) {
    const von = metadaten.abgeschlossenVon
      ? ` von ${metadaten.abgeschlossenVon}`
      : ''
    teile.push(
      `abgeschlossen ${formatLocalTime(metadaten.abgeschlossenAm)}${von}`,
    )
  }
  if (metadaten.kassensturzDifferenzCents !== null) {
    teile.push(
      `Kassensturz-Differenz ${formatCents(metadaten.kassensturzDifferenzCents)} €`,
    )
  }
  return (
    <p className="mt-1 text-sm text-muted-foreground">{teile.join(' · ')}</p>
  )
}

function Kennzahl({
  label,
  wert,
  destructive,
}: {
  label: string
  wert: string
  destructive?: boolean
}) {
  return (
    <div className="rounded-lg bg-sidebar p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div
        className={`mt-0.5 text-lg font-bold ${destructive ? 'text-destructive' : ''}`}
      >
        {wert}
      </div>
    </div>
  )
}

// ReportingResults ist der vollständige Tagesbericht ohne Tabs: formaler
// Berichtskopf mit Metadaten und Drucken-Knopf, vier Kennzahl-Kacheln, die
// Steuersatz-Tabelle und die zwei Mini-Listen (Umsatz pro Servicekraft,
// Stornierungen). Per Tailwind-print:-Klassen druckt nur diese Berichtsspalte.
export function ReportingResults({
  result,
  sitzung,
  loading,
}: {
  result: ReportingData
  sitzung: AbgeschlosseneSitzung
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
    <div className="flex flex-col gap-4 rounded-xl border p-5">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-bold">
            Tagesbericht Nr. {sitzung.zNr} — {sitzung.bezeichnung}
          </h2>
          <BerichtsMeta datum={sitzung.datum} result={result} />
        </div>
        <Button
          variant="outline"
          size="sm"
          className="shrink-0 print:hidden"
          onClick={() => {
            window.print()
          }}
        >
          <Printer className="size-4" />
          Drucken
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Kennzahl
          label="Kassierter Umsatz"
          wert={`${formatCents(summary.gesamtUmsatzCents)} €`}
        />
        <Kennzahl
          label="Bestellungen"
          wert={String(summary.anzahlBestellungen)}
        />
        <Kennzahl
          label="Direktverkauf"
          wert={`${formatCents(summary.direktverkaufUmsatzCents)} €`}
        />
        <Kennzahl
          label="Storniert"
          wert={`${formatCents(summary.gesamtStornierungenCents)} €`}
          destructive={summary.gesamtStornierungenCents > 0}
        />
      </div>

      <div>
        <div className="mb-2 text-sm font-semibold">Umsatz nach Steuersatz</div>
        {result.umsatzProSteuersatz.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Keine steuerrelevanten Umsätze in dieser Kassensitzung.
          </p>
        ) : (
          <div className="overflow-hidden rounded-lg border">
            <div className="grid grid-cols-[1.6fr_1fr_1fr_1fr] gap-x-3 bg-sidebar px-3 py-2 text-xs font-semibold text-muted-foreground">
              <span>Steuersatz</span>
              <span className="text-right">Brutto</span>
              <span className="text-right">Netto</span>
              <span className="text-right">Steuer</span>
            </div>
            {result.umsatzProSteuersatz.map((umsatz) => (
              <div
                key={umsatz.satz}
                className="grid grid-cols-[1.6fr_1fr_1fr_1fr] gap-x-3 border-t px-3 py-2 text-sm"
              >
                <span>{STEUERSATZ_LABEL[umsatz.satz]}</span>
                <span className="text-right font-semibold">
                  {formatCents(umsatz.bruttoCents)} €
                </span>
                <span className="text-right">
                  {formatCents(umsatz.nettoCents)} €
                </span>
                <span className="text-right">
                  {formatCents(umsatz.steuerCents)} €
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <div>
          <div className="mb-2 text-sm font-semibold">
            Umsatz pro Servicekraft
          </div>
          {breakdowns.umsatzProServicekraft.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Keine Zahlungen in dieser Kassensitzung.
            </p>
          ) : (
            <div className="flex flex-col text-sm">
              {breakdowns.umsatzProServicekraft.map((sk) => {
                const stornoAnzahl = stornoAnzahlByUserId.get(sk.userId) ?? 0
                return (
                  <div
                    key={sk.userId}
                    className="flex items-center justify-between gap-2 border-b py-2 last:border-b-0"
                  >
                    <span className="flex flex-col">
                      {formatBediener(sk.userName, sk.name)}
                      {stornoAnzahl > 0 && (
                        <StornoMarker anzahl={stornoAnzahl} />
                      )}
                    </span>
                    <span className="whitespace-nowrap font-semibold">
                      {formatCents(sk.zahlungenCents)} €
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>

        <div>
          <div className="mb-2 text-sm font-semibold">
            Stornierungen{' '}
            <span className="font-normal text-muted-foreground">
              ({summary.anzahlStornierungen} ·{' '}
              {formatCents(summary.gesamtStornierungenCents)} €)
            </span>
          </div>
          {result.stornierungen.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Keine Stornierungen in dieser Kassensitzung.
            </p>
          ) : (
            <div className="flex flex-col gap-2">
              {result.stornierungen.map((storno) => (
                <StornoItem
                  key={`${storno.zeitpunkt}-${String(storno.tischId)}-${String(storno.userId)}`}
                  storno={storno}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
