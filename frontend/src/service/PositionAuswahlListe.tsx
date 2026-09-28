import { formatEuro } from '@/lib/utils'

import { Stepper } from './Stepper'

// maxMenge ist die Obergrenze der auswählbaren Menge (Anzeige „N Stück“).
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

// Controlled: quantity logic stays in each drawer, and the list lives in DrawerBody, the drawer's only scroll area.
// Long names wrap instead of truncating: two truncated variants look alike, and picking the wrong position is costly.
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
              <div className="text-sm font-medium break-words">
                {position.name}
              </div>
              <div className="text-xs text-muted-foreground">
                {formatEuro(position.einzelpreisCents)} · {position.maxMenge}
                &nbsp;Stück
              </div>
            </div>
            <div className="ml-2 shrink-0">
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
