import type { Produkt } from '@/lib/produktSchemas'
import { formatPositionName } from '@/lib/utils'

import type { PositionRef } from '../../schemas'
import type { BestellPositionInput, Bestellung } from '../../table/Bestellung'
import type { Umbuchung } from '../../table/Umbuchung'
import type { AuswahlPosition } from '../PositionAuswahlListe'
import type { ReceiptPosition } from './Receipt'

// Bei einem Umbuchungs-Zugang ist `kommentar` der Richtungs-Autotext
// („Umbuchung von Tisch X") und damit der Vorgangstitel.
export function quelleTitel(quelle: Bestellung | Umbuchung): string {
  return quelle.art === 'bestellung' ? 'Bestellung' : quelle.kommentar
}

export function quelleZeitpunkt(quelle: Bestellung | Umbuchung): string {
  return quelle.art === 'bestellung' ? quelle.aufgenommenAm : quelle.umgebuchtAm
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

// Die vorhandene Menge wird zur auswählbaren Obergrenze (maxMenge).
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

export function selectPositionen<
  T extends { positionId: string; menge: number },
>(positionen: T[], ausgewaehlteMengen: Record<string, number>): T[] {
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
 * Effective target: `zielbetragCents` when entered (> 0), else `gesamtCents`.
 * Both values are `null` (caller hides them) unless `erhaltenCents > 0` and
 * `gesamtCents <= effektiverZielbetrag <= erhaltenCents`; Trinkgeld stays
 * `null` without a Zielbetrag.
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

/**
 * Two round-up suggestions above `gesamtCents`: the next whole Euro and the
 * next 5 Euro; when both coincide the 5-Euro one advances so the chips differ.
 * Examples: 1230 → [1300, 1500]; 1300 → [1400, 1500]; 450 → [500, 1000].
 */
export function aufrundenVorschlaege(gesamtCents: number): number[] {
  const einEuro = Math.floor(gesamtCents / 100) * 100 + 100
  let fuenfEuro = Math.floor(gesamtCents / 500) * 500 + 500
  if (fuenfEuro === einEuro) fuenfEuro += 500
  return [einEuro, fuenfEuro]
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

export function toPositionRefs(
  positionen: { positionId: string; menge: number }[],
): PositionRef[] {
  return positionen.map((p) => ({
    positionId: p.positionId,
    menge: p.menge,
  }))
}
