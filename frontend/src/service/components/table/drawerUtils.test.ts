import { describe, expect, it } from 'vitest'

import type { Produkt } from '@/service/product/Produkt'
import type { Position } from '@/service/table/Bestellung'

import {
  calculateTotalPrice,
  calculateZahlungsbetraege,
  selectPositionen,
  toBestellungData,
} from './drawerUtils'

describe('selectPositionen', () => {
  const positionen: Position[] = [
    {
      positionId: 'aaa-001',
      varianteId: 1,
      produktName: 'Bratwurst',
      varianteName: 'Normal',
      kategorie: 'essen',
      steuersatz: 'ermaessigt',
      einzelpreis: 350,
      menge: 5,
      bestellerUserId: 1,
      bestellerName: 'Anna',
    },
    {
      positionId: 'aaa-002',
      varianteId: 4,
      produktName: 'Pommes',
      varianteName: 'Klein',
      kategorie: 'essen',
      steuersatz: 'ermaessigt',
      einzelpreis: 250,
      menge: 3,
      bestellerUserId: 1,
      bestellerName: 'Anna',
    },
    {
      positionId: 'aaa-003',
      varianteId: 31,
      produktName: 'Softdrinks',
      varianteName: 'Cola',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreis: 200,
      menge: 2,
      bestellerUserId: 1,
      bestellerName: 'Anna',
    },
  ]

  it('returns only positionen with selected quantity > 0', () => {
    const result = selectPositionen(positionen, {
      'aaa-001': 2,
      'aaa-003': 1,
    })

    expect(result).toEqual([
      { ...positionen[0], menge: 2 },
      { ...positionen[2], menge: 1 },
    ])
  })

  it('returns empty array when no positionen are selected', () => {
    expect(selectPositionen(positionen, {})).toEqual([])
  })

  it('returns empty array for empty input', () => {
    expect(selectPositionen([], {})).toEqual([])
  })

  it('handles single item selection', () => {
    const result = selectPositionen(positionen, { 'aaa-002': 1 })

    expect(result).toEqual([{ ...positionen[1], menge: 1 }])
  })

  it('ignores selection keys that do not match any position', () => {
    const result = selectPositionen(positionen, { 'zzz-999': 5 })

    expect(result).toEqual([])
  })

  it('filters out positionen where selected quantity is 0', () => {
    const result = selectPositionen(positionen, {
      'aaa-001': 0,
      'aaa-002': 3,
    })

    expect(result).toEqual([{ ...positionen[1], menge: 3 }])
  })
})

describe('calculateTotalPrice', () => {
  it('calculates total for multiple items', () => {
    const items = [
      { einzelpreis: 350, menge: 2 },
      { einzelpreis: 250, menge: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(950)
  })

  it('returns 0 for empty array', () => {
    expect(calculateTotalPrice([])).toBe(0)
  })

  it('handles single item', () => {
    expect(calculateTotalPrice([{ einzelpreis: 200, menge: 3 }])).toBe(600)
    expect(calculateTotalPrice([{ einzelpreis: 300, menge: 1 }])).toBe(300)
  })

  it('handles zero-cent positionen', () => {
    const items = [
      { einzelpreis: 0, menge: 5 },
      { einzelpreis: 200, menge: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(200)
  })
})

describe('calculateZahlungsbetraege', () => {
  describe('without a Zielbetrag (Rückgeld only)', () => {
    it('returns the change when more than the total is received', () => {
      expect(calculateZahlungsbetraege(1350, 2000, 0)).toEqual({
        rueckgeldCents: 650,
        trinkgeldCents: null,
      })
    })

    it('returns 0 change when the exact amount is received', () => {
      expect(calculateZahlungsbetraege(1350, 1350, 0)).toEqual({
        rueckgeldCents: 0,
        trinkgeldCents: null,
      })
    })

    it('returns null when too little is received', () => {
      expect(calculateZahlungsbetraege(1350, 1000, 0)).toEqual({
        rueckgeldCents: null,
        trinkgeldCents: null,
      })
    })

    it('returns null when nothing is received (empty field)', () => {
      expect(calculateZahlungsbetraege(1350, 0, 0)).toEqual({
        rueckgeldCents: null,
        trinkgeldCents: null,
      })
    })
  })

  describe('with a Zielbetrag (Trinkgeld + Rückgeld)', () => {
    it('derives Trinkgeld and Rückgeld from the target amount', () => {
      expect(calculateZahlungsbetraege(1350, 2000, 1500)).toEqual({
        rueckgeldCents: 500,
        trinkgeldCents: 150,
      })
    })

    it('reports zero Trinkgeld when the target equals the total', () => {
      expect(calculateZahlungsbetraege(1350, 2000, 1350)).toEqual({
        rueckgeldCents: 650,
        trinkgeldCents: 0,
      })
    })

    it('returns null when the target is below the total (negative Trinkgeld)', () => {
      expect(calculateZahlungsbetraege(1350, 2000, 1300)).toEqual({
        rueckgeldCents: null,
        trinkgeldCents: null,
      })
    })

    it('returns null when the target exceeds the received cash (negative Rückgeld)', () => {
      expect(calculateZahlungsbetraege(1350, 1400, 1500)).toEqual({
        rueckgeldCents: null,
        trinkgeldCents: null,
      })
    })
  })
})

describe('toBestellungData', () => {
  const pommes: Produkt = {
    id: 1,
    name: 'Pommes',
    kategorie: 'essen',
    status: 'active',
    varianten: [
      {
        id: 7,
        name: 'mit Ketchup',
        preisCents: 300,
        status: 'active',
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
      },
    ],
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  }

  it('sets the receipt preview name to {Produkt} {Variante}', () => {
    const { receiptItems } = toBestellungData([pommes], { 7: 2 })

    expect(receiptItems).toEqual([
      { name: 'Pommes mit Ketchup', einzelpreis: 300, menge: 2 },
    ])
  })

  it('ignores variants without a selected quantity', () => {
    expect(toBestellungData([pommes], {}).receiptItems).toEqual([])
  })
})
