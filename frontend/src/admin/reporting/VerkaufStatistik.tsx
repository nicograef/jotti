import { Fragment } from 'react'

import { formatEuro, formatPositionName } from '@/lib/utils'

import type { ProduktStatistik } from './types'

const KATEGORIE_LABEL: Record<string, string> = {
  essen: 'Essen',
  getraenk: 'Getränke',
  sonstiges: 'Sonstiges',
}

function kategorieLabel(kategorie: string): string {
  return KATEGORIE_LABEL[kategorie] ?? kategorie
}

// StatistikZeile ist eine Tabellenzeile des Verkaufsabschnitts: Beschriftung,
// ausgegebene Menge (ganze Portionen) und Umsatz. `bold` hebt die
// Produkt-Zwischensumme hervor, `indent` rückt Variantenzeilen darunter ein.
function StatistikZeile({
  label,
  ausgegebeneMenge,
  umsatzCents,
  bold = false,
  indent = false,
}: {
  label: string
  ausgegebeneMenge: number
  umsatzCents: number
  bold?: boolean
  indent?: boolean
}) {
  const betonung = bold ? 'font-semibold' : ''
  return (
    <div className="grid grid-cols-[1.6fr_1fr_1fr] gap-x-3 border-t px-3 py-2 text-sm">
      <span
        className={`${indent ? 'pl-4 text-muted-foreground' : ''} ${betonung}`}
      >
        {label}
      </span>
      <span className={`text-right tabular-nums ${betonung}`}>
        {ausgegebeneMenge}
      </span>
      <span className={`text-right tabular-nums ${betonung}`}>
        {formatEuro(umsatzCents)}
      </span>
    </div>
  )
}

// VerkaufStatistik zeigt die Verkäufe je Produkt und Variante einer
// Kassensitzung, in Kategorie-Abschnitte (Essen → Getränke → Sonstiges)
// gegliedert. Zwei bewusst getrennte Zahlen: ausgegebene Menge (Produktion) und
// Umsatz (Einnahmen) — nicht ineinander umrechenbar. Ein-Varianten-Produkte
// erscheinen als eine Zeile; sonst Produkt-Zwischensumme mit eingerückten
// Varianten. Das Backend liefert die Liste fertig gruppiert und sortiert.
export function VerkaufStatistik({
  produktStatistik,
}: {
  produktStatistik: ProduktStatistik[]
}) {
  return (
    <div>
      <div className="mb-2 text-sm font-semibold">Verkäufe pro Produkt</div>
      <p className="mb-2 text-xs text-muted-foreground">
        „Ausgegeben" zählt zubereitete Portionen (bestellt minus Korrekturen,
        inklusive Direktverkauf); „Umsatz" zählt die Einnahmen (kassiert und
        Direktverkauf minus Storno). Die beiden Zahlen ruhen auf
        unterschiedlichen Grundlagen und sind nicht ineinander umrechenbar.
      </p>
      {produktStatistik.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Keine Verkäufe in dieser Kassensitzung.
        </p>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <div className="grid grid-cols-[1.6fr_1fr_1fr] gap-x-3 bg-sidebar px-3 py-2 text-xs font-semibold text-muted-foreground">
            <span>Produkt</span>
            <span className="text-right">Ausgegeben</span>
            <span className="text-right">Umsatz</span>
          </div>
          {produktStatistik.map((produkt, index) => {
            const neuerAbschnitt =
              index === 0 ||
              produktStatistik[index - 1].kategorie !== produkt.kategorie
            const einVariante = produkt.varianten.length === 1

            return (
              <Fragment key={`${produkt.kategorie}-${produkt.produktName}`}>
                {neuerAbschnitt && (
                  <div className="border-t bg-muted/40 px-3 py-1.5 text-xs font-semibold">
                    {kategorieLabel(produkt.kategorie)}
                  </div>
                )}
                {einVariante ? (
                  <StatistikZeile
                    label={formatPositionName(
                      produkt.produktName,
                      produkt.varianten[0].varianteName,
                    )}
                    ausgegebeneMenge={produkt.ausgegebeneMenge}
                    umsatzCents={produkt.umsatzCents}
                  />
                ) : (
                  <>
                    <StatistikZeile
                      label={produkt.produktName}
                      ausgegebeneMenge={produkt.ausgegebeneMenge}
                      umsatzCents={produkt.umsatzCents}
                      bold
                    />
                    {produkt.varianten.map((variante) => (
                      <StatistikZeile
                        key={variante.varianteId}
                        label={variante.varianteName}
                        ausgegebeneMenge={variante.ausgegebeneMenge}
                        umsatzCents={variante.umsatzCents}
                        indent
                      />
                    ))}
                  </>
                )}
              </Fragment>
            )
          })}
        </div>
      )}
    </div>
  )
}
