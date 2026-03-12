import type { Position } from '../../table/Bestellung'

export function selectPositionen(
  positionen: Position[],
  selectedQuantity: Record<number, number>,
): Position[] {
  return positionen
    .map((position) => ({
      ...position,
      menge: selectedQuantity[position.id] || 0,
    }))
    .filter((position) => position.menge > 0)
}

export function calculateTotalPrice(positionen: Position[]): number {
  return positionen.reduce(
    (total, position) => total + position.preisCents * position.menge,
    0,
  )
}
