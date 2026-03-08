import type { Position } from '../../table/Bestellung'

export function selectPositionen(
  positionen: Position[],
  selectedQuantity: Record<number, number>,
): Position[] {
  return positionen
    .map((position) => ({
      ...position,
      quantity: selectedQuantity[position.id] || 0,
    }))
    .filter((position) => position.quantity > 0)
}

export function calculateTotalPrice(positionen: Position[]): number {
  return positionen.reduce(
    (total, position) => total + position.preisCents * position.quantity,
    0,
  )
}
