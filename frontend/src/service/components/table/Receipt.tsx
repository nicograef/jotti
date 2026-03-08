import { ScrollArea } from '@/components/ui/scroll-area'
import { formatCents } from '@/lib/utils'
import type { Position } from '@/service/table/Bestellung'

export function Receipt({
  positionen,
  totalPrice,
}: {
  positionen: Position[]
  totalPrice?: number
}) {
  return (
    <>
      <ScrollArea
        className={`inset-shadow-sm h-${Math.min(positionen.length * 10, 50).toString()}`} // max height 50 units, 10 units per variant
      >
        <div className="px-4 pt-2 pb-0 space-y-2">
          {positionen.map((position) => {
            return (
              <div
                key={position.id}
                className="flex justify-between border-b pb-2 last:border-0"
              >
                <div>
                  {position.quantity} x {position.name}
                </div>
                <div>
                  € {formatCents(position.preisCents * position.quantity)}
                </div>
              </div>
            )
          })}
        </div>
      </ScrollArea>
      <div className="flex justify-between font-bold px-4 pt-2 pb-4 border-t-2">
        {totalPrice !== undefined && (
          <>
            <div>Gesamt</div>
            <div>€ {formatCents(totalPrice)}</div>
          </>
        )}
      </div>
    </>
  )
}
