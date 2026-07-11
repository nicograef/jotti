import { Badge } from '@/components/ui/badge'
import { Item, ItemContent } from '@/components/ui/item'
import { formatCents, formatPositionName } from '@/lib/utils'

import type { StornierungDetail } from './types'
import { formatBediener, formatLocalTime } from './utils'

// StornoItem rendert einen einzelnen Stornierungs-Eintrag einheitlich für das
// Live-Dashboard und die Kassenberichte: Tisch/Direktverkauf + Bediener, Uhrzeit
// (HH:MM), Bar-Rückgabe-Status, Betrag und die stornierten Positionen.
export function StornoItem({ storno }: { storno: StornierungDetail }) {
  return (
    <Item variant="outline" size="sm">
      <ItemContent>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div>
            <p className="text-sm font-medium">
              {storno.quelle === 'direktverkauf'
                ? 'Direktverkauf'
                : storno.tischName}{' '}
              · {formatBediener(storno.userName, storno.name)}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {formatLocalTime(storno.zeitpunkt)}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant={storno.barRueckgabe ? 'outline' : 'secondary'}>
              {storno.barRueckgabe ? 'Bar-Rückgabe' : 'Geldneutral'}
            </Badge>
            <span className="whitespace-nowrap text-sm font-semibold">
              {formatCents(storno.betragCents)} €
            </span>
          </div>
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
                  {pos.menge}×{' '}
                  {formatPositionName(pos.produktName, pos.varianteName)}
                </span>
                <span className="whitespace-nowrap">
                  {formatCents(pos.einzelpreisCents)} €
                </span>
              </li>
            ))}
          </ul>
        )}
      </ItemContent>
    </Item>
  )
}
