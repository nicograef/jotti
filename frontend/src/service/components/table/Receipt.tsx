import { ScrollArea } from '@/components/ui/scroll-area'
import { formatCents } from '@/lib/utils'
import type { LineItem } from '@/service/table/Order'

export function Receipt({
  variants,
  totalPrice,
}: {
  variants: LineItem[]
  totalPrice?: number
}) {
  return (
    <>
      <ScrollArea
        className={`inset-shadow-sm h-${Math.min(variants.length * 10, 50).toString()}`} // max height 50 units, 10 units per variant
      >
        <div className="px-4 pt-2 pb-0 space-y-2">
          {variants.map((variant) => {
            return (
              <div
                key={variant.id}
                className="flex justify-between border-b pb-2 last:border-0"
              >
                <div>
                  {variant.quantity} x {variant.name}
                </div>
                <div>
                  € {formatCents(variant.priceCents * variant.quantity)}
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
