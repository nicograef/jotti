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

/**
 * Derives the cash change (Rückgeld) and tip (Trinkgeld) for a payment.
 *
 * The effective target amount is `zielbetragCents` when a Zielbetrag was
 * entered (> 0), otherwise the order total `gesamtCents`. Both values are only
 * returned when `gesamtCents <= effektiverZielbetrag <= erhaltenCents`;
 * otherwise (negative tip or too little cash) both are `null` so the caller
 * hides them. Trinkgeld is only reported when a Zielbetrag was entered.
 */
export function calculateZahlungsbetraege(
  gesamtCents: number,
  erhaltenCents: number,
  zielbetragCents: number,
): { rueckgeldCents: number | null; trinkgeldCents: number | null } {
  const hasZielbetrag = zielbetragCents > 0
  const effektiverZielbetrag = hasZielbetrag ? zielbetragCents : gesamtCents

  const gueltig =
    erhaltenCents > 0 &&
    gesamtCents <= effektiverZielbetrag &&
    effektiverZielbetrag <= erhaltenCents

  if (!gueltig) {
    return { rueckgeldCents: null, trinkgeldCents: null }
  }

  return {
    rueckgeldCents: erhaltenCents - effektiverZielbetrag,
    trinkgeldCents: hasZielbetrag ? effektiverZielbetrag - gesamtCents : null,
  }
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
