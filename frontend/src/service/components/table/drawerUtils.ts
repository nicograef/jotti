import type { Ausgabe } from '../../table/Ausgabe'
import type { Auszahlung } from '../../table/Auszahlung'
import type { Bestellung, Position, PositionRef } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Zahlung } from '../../table/Zahlung'
import type { ReceiptPosition } from './Receipt'

type HistorieEintrag = Bestellung | Zahlung | Stornierung | Ausgabe | Auszahlung

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

export function getStornierbarePositionen(
  bestellung: Bestellung,
  historie: HistorieEintrag[],
) {
  const stornierteMengen = new Map<string, number>()

  historie.forEach((item) => {
    if (!Object.prototype.hasOwnProperty.call(item, 'storniertAm')) {
      return
    }

    const stornierung = item as Stornierung
    stornierung.positionen.forEach((position) => {
      const bisherigeMenge = stornierteMengen.get(position.positionId) ?? 0
      stornierteMengen.set(position.positionId, bisherigeMenge + position.menge)
    })
  })

  return bestellung.positionen.flatMap((position) => {
    const verbleibendeMenge =
      position.menge - (stornierteMengen.get(position.positionId) ?? 0)
    if (verbleibendeMenge <= 0) {
      return []
    }

    return [{ ...position, menge: verbleibendeMenge }]
  })
}

export function getUmbuchbarePositionen(
  bestellung: Bestellung,
  historie: HistorieEintrag[],
) {
  const stornierteMengen = new Map<string, number>()
  const bezahlteMengen = new Map<string, number>()

  historie.forEach((item) => {
    if (Object.prototype.hasOwnProperty.call(item, 'storniertAm')) {
      const stornierung = item as Stornierung
      stornierung.positionen.forEach((position) => {
        const bisherigeMenge = stornierteMengen.get(position.positionId) ?? 0
        stornierteMengen.set(
          position.positionId,
          bisherigeMenge + position.menge,
        )
      })
      return
    }

    if (Object.prototype.hasOwnProperty.call(item, 'kassiertAm')) {
      const zahlung = item as Zahlung
      zahlung.positionen.forEach((position) => {
        const bisherigeMenge = bezahlteMengen.get(position.positionId) ?? 0
        bezahlteMengen.set(position.positionId, bisherigeMenge + position.menge)
      })
    }
  })

  return bestellung.positionen.flatMap((position) => {
    const storniert = stornierteMengen.get(position.positionId) ?? 0
    const bezahlt = bezahlteMengen.get(position.positionId) ?? 0
    const verbleibendeMenge = position.menge - storniert - bezahlt
    if (verbleibendeMenge <= 0) {
      return []
    }

    return [{ ...position, menge: verbleibendeMenge }]
  })
}
