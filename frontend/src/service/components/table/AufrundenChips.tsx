import { EuroInput } from '@/components/common/EuroInput'
import { Label } from '@/components/ui/label'
import { cn, formatCents, formatEuro, parseCents } from '@/lib/utils'

import { aufrundenVorschlaege } from './drawerUtils'

interface AufrundenChipsProps {
  // Gesamtbetrag des Vorgangs; Basis für den „genau"-Chip und die Vorschläge.
  gesamtCents: number
  // Aktueller Zielbetrag als Euro-String ('' = kein Trinkgeld, „genau").
  zielbetragEuro: string
  onZielbetragEuroChange: (euro: string) => void
  // Ob das freie Euro-Feld hinter „Anderer …" eingeblendet ist.
  andererAktiv: boolean
  onAndererAktivChange: (aktiv: boolean) => void
}

// chipKlassen liefert die Pill-Optik eines Chips (mindestens 44 px hoch,
// tabular-nums für die Beträge); der aktive Chip trägt die Primärfläche.
function chipKlassen(aktiv: boolean): string {
  return cn(
    'inline-flex min-h-11 items-center justify-center rounded-full border px-4 text-sm font-medium tabular-nums transition-colors outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50',
    aktiv
      ? 'border-primary bg-primary text-primary-foreground'
      : 'border-border bg-background text-foreground hover:bg-muted',
  )
}

// Aufrunden-Chips fürs Trinkgeld: „{Betrag} € genau" für den exakten Betrag,
// zwei glatte Aufrunden-Vorschläge und „Anderer …" für einen freien Zielbetrag.
// Presentation-neutral und Variant-agnostisch (Sheet wie Spalte); der
// Zielbetrag-State liegt beim Aufrufer, damit dessen warLeer-Reset ihn miträumt.
export function AufrundenChips(props: AufrundenChipsProps) {
  const zielCents = parseCents(props.zielbetragEuro)
  const [ersterVorschlag, zweiterVorschlag] = aufrundenVorschlaege(
    props.gesamtCents,
  )

  const genauAktiv = !props.andererAktiv && zielCents === 0

  const waehleGenau = () => {
    props.onAndererAktivChange(false)
    props.onZielbetragEuroChange('')
  }

  const waehleVorschlag = (cents: number) => {
    props.onAndererAktivChange(false)
    // Erneuter Tap auf den aktiven Vorschlag wählt ab, zurück zu „genau".
    if (!props.andererAktiv && zielCents === cents) {
      props.onZielbetragEuroChange('')
    } else {
      props.onZielbetragEuroChange(formatCents(cents))
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <div role="group" aria-label="Aufrunden fürs Trinkgeld">
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            aria-pressed={genauAktiv}
            className={chipKlassen(genauAktiv)}
            onClick={waehleGenau}
          >
            {formatEuro(props.gesamtCents)} genau
          </button>
          {[ersterVorschlag, zweiterVorschlag].map((cents) => {
            const aktiv = !props.andererAktiv && zielCents === cents
            return (
              <button
                key={cents}
                type="button"
                aria-pressed={aktiv}
                className={chipKlassen(aktiv)}
                onClick={() => {
                  waehleVorschlag(cents)
                }}
              >
                {formatEuro(cents)}
              </button>
            )
          })}
          <button
            type="button"
            aria-pressed={props.andererAktiv}
            className={chipKlassen(props.andererAktiv)}
            onClick={() => {
              props.onAndererAktivChange(true)
            }}
          >
            Anderer …
          </button>
        </div>
      </div>
      {props.andererAktiv && (
        <div className="flex items-center justify-between gap-3">
          <Label htmlFor="zielbetrag">Zahlbetrag inkl. Trinkgeld</Label>
          <EuroInput
            id="zielbetrag"
            value={props.zielbetragEuro}
            onValueChange={props.onZielbetragEuroChange}
            className="w-28"
          />
        </div>
      )}
    </div>
  )
}
