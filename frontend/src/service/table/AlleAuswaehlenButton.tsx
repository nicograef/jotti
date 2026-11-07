import { CircleCheck } from 'lucide-react'

import type { MengenSteuerung } from '@/hooks/use-mengen'
import { formatEuro } from '@/lib/utils'

import { calculateTotalPrice, toMengen } from './drawerUtils'

/**
 * `'alle'` — Umbuchung (every position), `'meine'` — Kassieren (only the
 * caller's own). Both with the correct grammatical number and the selection sum.
 */
function formatAlleAuswaehlenLabel(
  anzahl: number,
  summeCents: number,
  variante: 'alle' | 'meine',
): string {
  let auswahl: string
  if (variante === 'meine') {
    auswahl =
      anzahl === 1
        ? 'Meine Position auswählen'
        : `Meine ${anzahl.toString()} Positionen auswählen`
  } else {
    auswahl =
      anzahl === 1
        ? '1 Position auswählen'
        : `Alle ${anzahl.toString()} Positionen auswählen`
  }
  return `${auswahl} · ${formatEuro(summeCents)}`
}

// Wählt die übergebenen Positionen voll aus; sind sie es schon, hebt derselbe
// Tap ihre Auswahl auf. Die Auswahl anderer Positionen bleibt in beiden
// Richtungen stehen.
export function AlleAuswaehlenButton({
  positionen,
  steuerung,
  variante = 'alle',
}: {
  positionen: { positionId: string; menge: number; einzelpreisCents: number }[]
  steuerung: Pick<MengenSteuerung<string>, 'mengen' | 'setAll'>
  variante?: 'alle' | 'meine'
}) {
  const alleVollAusgewaehlt = positionen.every(
    (position) =>
      (steuerung.mengen[position.positionId] || 0) === position.menge,
  )

  return (
    <button
      type="button"
      onClick={() => {
        if (alleVollAusgewaehlt) {
          const ids = new Set(positionen.map((p) => p.positionId))
          steuerung.setAll(
            Object.fromEntries(
              Object.entries(steuerung.mengen).filter(([id]) => !ids.has(id)),
            ),
          )
        } else {
          steuerung.setAll({ ...steuerung.mengen, ...toMengen(positionen) })
        }
      }}
      className="flex h-11 w-full items-center justify-center gap-2 rounded-lg border border-primary/50 bg-primary/5 px-4 text-sm font-medium text-primary"
    >
      <CircleCheck className="size-5" />
      {alleVollAusgewaehlt
        ? 'Auswahl aufheben'
        : formatAlleAuswaehlenLabel(
            positionen.length,
            calculateTotalPrice(positionen),
            variante,
          )}
    </button>
  )
}
