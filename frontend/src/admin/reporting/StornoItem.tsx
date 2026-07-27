import { Badge } from '@/components/ui/badge'
import { Item, ItemContent } from '@/components/ui/item'
import { formatEuro, formatPositionName } from '@/lib/utils'

import type { StornierungDetail } from './types'
import { formatLocalTime, formatServicekraft } from './utils'

// StornoItem rendert einen einzelnen Stornierungs-Eintrag einheitlich für das
// Live-Dashboard und die Kassenberichte: Tisch/Direktverkauf + die betroffenen
// Servicekräfte, Uhrzeit (HH:MM), Bar-Rückgabe-Status, Betrag und die
// stornierten Positionen. Genannt wird zuerst, wen der Storno betrifft (wessen
// Vorgang er rückgängig macht); wer ihn ausgelöst hat, folgt nur als gedämpfter
// Zusatz, wenn er nicht selbst betroffen ist.
export function StornoItem({ storno }: { storno: StornierungDetail }) {
  const betroffene = storno.betroffene
    .map((b) => formatServicekraft(b.userName, b.name))
    .join(', ')
  const akteurIstBetroffen = storno.betroffene.some(
    (b) => b.userId === storno.akteur.userId,
  )

  return (
    <Item variant="outline" size="sm">
      <ItemContent>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div>
            <p className="text-sm font-medium">
              {storno.quelle === 'direktverkauf'
                ? 'Direktverkauf'
                : storno.tischName}{' '}
              · {betroffene}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {formatLocalTime(storno.zeitpunkt)}
              {!akteurIstBetroffen &&
                ` · storniert von ${formatServicekraft(
                  storno.akteur.userName,
                  storno.akteur.name,
                )}`}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant={storno.barRueckgabe ? 'outline' : 'secondary'}>
              {storno.barRueckgabe ? 'Bar-Rückgabe' : 'Geldneutral'}
            </Badge>
            <span className="whitespace-nowrap text-sm font-semibold">
              {formatEuro(storno.betragCents)}
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
                  {formatEuro(pos.einzelpreisCents)}
                </span>
              </li>
            ))}
          </ul>
        )}
      </ItemContent>
    </Item>
  )
}
