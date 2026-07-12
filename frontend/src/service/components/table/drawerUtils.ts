import { formatPositionName } from '@/lib/utils'

import type { Produkt } from '../../product/Produkt'
import type {
  BestellPositionInput,
  Bestellung,
  Position,
  PositionRef,
} from '../../table/Bestellung'
import type { Umbuchung } from '../../table/Umbuchung'
import type { AuswahlPosition } from '../PositionAuswahlListe'
import type { ReceiptPosition } from './Receipt'

// quelleTitel liefert den menschenlesbaren Vorgangstitel für die Folge-Drawer
// (Stornieren/Umbuchen): „Bestellung" bzw. für einen Umbuchungs-Zugang den
// Richtungs-Autotext („Umbuchung von Tisch X"), wie in der Historien-Zeile.
export function quelleTitel(quelle: Bestellung | Umbuchung): string {
  return quelle.art === 'bestellung' ? 'Bestellung' : quelle.kommentar
}

// Minimale Positionsform, die toAuswahlPositionen benötigt (Position und
// VerkaufPosition erfüllen sie beide).
interface AuswaehlbarePosition {
  positionId: string
  produktName: string
  varianteName: string
  einzelpreisCents: number
  menge: number
}

// toAuswahlPositionen bringt fachliche Positionen in die von PositionAuswahlListe
// erwartete Form: die vorhandene Menge wird zur auswählbaren Obergrenze.
export function toAuswahlPositionen(
  positionen: AuswaehlbarePosition[],
): AuswahlPosition[] {
  return positionen.map((position) => ({
    id: position.positionId,
    name: formatPositionName(position.produktName, position.varianteName),
    einzelpreisCents: position.einzelpreisCents,
    maxMenge: position.menge,
  }))
}

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
  positionen: { einzelpreisCents: number; menge: number }[],
): number {
  return positionen.reduce(
    (total, position) => total + position.einzelpreisCents * position.menge,
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

export function toBestellungData(
  products: Produkt[],
  ausgewaehlteMengen: Record<number, number>,
): { receiptItems: ReceiptPosition[]; inputItems: BestellPositionInput[] } {
  const items = products.flatMap((p) =>
    p.varianten
      .filter((v) => (ausgewaehlteMengen[v.id] || 0) > 0)
      .map((v) => ({
        produktId: p.id,
        varianteId: v.id,
        name: formatPositionName(p.name, v.name),
        einzelpreisCents: v.preisCents,
        menge: ausgewaehlteMengen[v.id],
      })),
  )

  return {
    receiptItems: items.map((i) => ({
      name: i.name,
      einzelpreisCents: i.einzelpreisCents,
      menge: i.menge,
    })),
    inputItems: items.map((i) => ({
      produktId: i.produktId,
      varianteId: i.varianteId,
      menge: i.menge,
    })),
  }
}

export function toReceiptItems(
  positionen: {
    produktName: string
    varianteName: string
    einzelpreisCents: number
    menge: number
  }[],
): ReceiptPosition[] {
  return positionen.map((p) => ({
    name: formatPositionName(p.produktName, p.varianteName),
    einzelpreisCents: p.einzelpreisCents,
    menge: p.menge,
  }))
}

export function toPositionRefs(positionen: Position[]): PositionRef[] {
  return positionen.map((p) => ({
    positionId: p.positionId,
    menge: p.menge,
  }))
}
