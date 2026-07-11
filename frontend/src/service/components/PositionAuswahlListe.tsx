import { formatCents } from '@/lib/utils'

import { Stepper } from './Stepper'

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

// PositionAuswahlListe rendert eine Positionsliste mit Mengen-Steppern
// (Minus/Anzahl/Plus). Sie ist controlled: die Mengenlogik (Grenzen,
// Voll-Vorauswahl) bleibt im jeweiligen Drawer, hier liegt nur die Darstellung.
// Die Liste scrollt nicht selbst — sie liegt im DrawerBody, dem einzigen
// Scrollbereich des Drawers. Lange Namen werden per truncate abgeschnitten,
// sodass die Stepper-Buttons an ihrem Platz bleiben.
export function PositionAuswahlListe({
  positionen,
  mengen,
  onAdd,
  onRemove,
}: PositionAuswahlListeProps) {
  return (
    <div className="px-4 space-y-2">
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
            <div className="ml-2">
              <Stepper
                menge={selected}
                onAdd={() => {
                  onAdd(position.id)
                }}
                onRemove={() => {
                  onRemove(position.id)
                }}
                addLabel={`${position.name} hinzufügen`}
                removeLabel={`${position.name} verringern`}
                addDisabled={selected >= position.maxMenge}
              />
            </div>
          </div>
        )
      })}
    </div>
  )
}
