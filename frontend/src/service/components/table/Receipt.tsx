import { ScrollArea } from '@/components/ui/scroll-area'
import type { OrderProduct } from '@/service/table/Order'

export function Receipt({
  products,
  totalPrice,
}: {
  products: OrderProduct[]
  totalPrice?: number
}) {
  return (
    <>
      <ScrollArea
        className={`inset-shadow-sm h-${Math.min(products.length * 10, 50).toString()}`} // max height 50 units, 10 units per product
      >
        <div className="px-4 pt-2 pb-0 space-y-2">
          {products.map((product) => {
            return (
              <div
                key={product.id}
                className="flex justify-between border-b pb-2 last:border-0"
              >
                <div>
                  {product.quantity} x {product.name}
                </div>
                <div>
                  €{' '}
                  {((product.netPriceCents / 100) * product.quantity).toFixed(
                    2,
                  )}
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
            <div>€ {(totalPrice / 100).toFixed(2)}</div>
          </>
        )}
      </div>
    </>
  )
}
