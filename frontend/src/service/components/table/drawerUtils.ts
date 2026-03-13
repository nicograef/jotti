import type { Position, PositionRef } from '../../table/Bestellung'
import type { ReceiptPosition } from './Receipt'

export function selectPositionen(
  positionen: Position[],
  ausgewaehlteMengen: Record<string, number>,
): Position[] {
  return positionen
    .map((position) => ({
      ...position,
      menge: ausgewaehlteMengen[position.positionId] || 0,
    }))
    .filter((position) => position.menge > 0)
}

export function calculateTotalPrice(
  positionen: { einzelpreis: number; menge: number }[],
): number {
  return positionen.reduce(
    (total, position) => total + position.einzelpreis * position.menge,
    0,
  )
}

export function toReceiptItems(positionen: Position[]): ReceiptPosition[] {
  return positionen.map((p) => ({
    name: `${p.produktName} ${p.varianteName}`,
    einzelpreis: p.einzelpreis,
    menge: p.menge,
  }))
}

export function toPositionRefs(positionen: Position[]): PositionRef[] {
  return positionen.map((p) => ({
    positionId: p.positionId,
    menge: p.menge,
  }))
}
