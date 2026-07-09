import { Minus, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { formatCents } from '@/lib/utils'

// AuswahlPosition ist die minimale Form, die PositionAuswahlListe braucht: ein
// bereits formatierter Name, der Einzelpreis und die Obergrenze der auswählbaren
// Menge (für die Anzeige „N Stück“).
export interface AuswahlPosition {
  id: string
  name: string
  einzelpreisCents: number
  maxMenge: number
}

interface PositionAuswahlListeProps {
  positionen: AuswahlPosition[]
  mengen: Record<string, number>
  onAdd: (id: string) => void
  onRemove: (id: string) => void
}

// PositionAuswahlListe rendert eine scrollbare Positionsliste mit Mengen-Steppern
// (Minus/Anzahl/Plus). Sie ist controlled: die Mengenlogik (Grenzen,
// Voll-Vorauswahl) bleibt im jeweiligen Drawer, hier liegen nur Darstellung und
// der native Scrollbereich. Nur diese Liste scrollt (natives overflow-y-auto,
// dvh-basierte Maximalhöhe, damit Kommentarfeld und Buttons auch mit geöffneter
// Bildschirmtastatur erreichbar bleiben); Header, Summe, Kommentarfeld und Footer
// liegen im Drawer außerhalb dieser Komponente. Lange Namen werden per truncate
// abgeschnitten, sodass die Stepper-Buttons an ihrem Platz bleiben.
export function PositionAuswahlListe({
  positionen,
  mengen,
  onAdd,
  onRemove,
}: PositionAuswahlListeProps) {
  return (
    <div className="px-4 space-y-2 overflow-y-auto max-h-[45dvh]">
      {positionen.map((position) => {
        const selected = mengen[position.id] || 0
        return (
          <div
            key={position.id}
            className="flex items-center justify-between border-b pb-2 last:border-0"
          >
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">
                {position.name}
              </div>
              <div className="text-xs text-muted-foreground">
                {formatCents(position.einzelpreisCents)}&nbsp;€ ·{' '}
                {position.maxMenge}&nbsp;Stück
              </div>
            </div>
            <div className="flex items-center gap-1 ml-2">
              <Button
                variant="secondary"
                size="icon"
                className="h-8 w-8"
                aria-label={`${position.name} verringern`}
                onClick={() => {
                  onRemove(position.id)
                }}
              >
                <Minus className={selected > 0 ? '' : 'opacity-50'} size={16} />
              </Button>
              <span className="font-bold tabular-nums text-center w-6 text-sm">
                {selected}
              </span>
              <Button
                variant="secondary"
                size="icon"
                className="h-8 w-8"
                aria-label={`${position.name} hinzufügen`}
                onClick={() => {
                  onAdd(position.id)
                }}
              >
                <Plus size={16} />
              </Button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
