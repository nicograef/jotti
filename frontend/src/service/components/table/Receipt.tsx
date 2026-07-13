import { formatCents } from '@/lib/utils'

export interface ReceiptPosition {
  name: string
  einzelpreisCents: number
  menge: number
}

export function Receipt({
  positionen,
  totalPrice,
}: {
  positionen: ReceiptPosition[]
  totalPrice?: number
}) {
  return (
    <>
      <div className="px-4 pt-2 pb-0 space-y-2">
        {positionen.map((position) => {
          return (
            <div
              key={`${position.name}-${position.einzelpreisCents.toString()}`}
              className="flex justify-between border-b pb-2 last:border-0"
            >
              <div>
                {position.menge} x {position.name}
              </div>
              <div>
                {formatCents(position.einzelpreisCents * position.menge)}
                &nbsp;€
              </div>
            </div>
          )
        })}
      </div>
      {totalPrice !== undefined && (
        <div className="flex justify-between font-bold px-4 pt-2 pb-4 border-t-2">
          <div>Gesamt</div>
          <div>{formatCents(totalPrice)}&nbsp;€</div>
        </div>
      )}
    </>
  )
}
