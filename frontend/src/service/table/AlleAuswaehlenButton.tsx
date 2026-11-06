import { CircleCheck } from 'lucide-react'

import type { MengenSteuerung } from '@/hooks/use-mengen'
import { formatAlleAuswaehlenLabel } from '@/lib/utils'

import { calculateTotalPrice, toMengen } from './drawerUtils'

// Wählt alle Positionen voll aus; sind sie es schon, hebt derselbe Tap die
// Auswahl auf.
export function AlleAuswaehlenButton({
  positionen,
  steuerung,
  variante,
}: {
  positionen: { positionId: string; menge: number; einzelpreisCents: number }[]
  steuerung: Pick<MengenSteuerung<string>, 'mengen' | 'reset' | 'setAll'>
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
          steuerung.reset()
        } else {
          steuerung.setAll(toMengen(positionen))
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
