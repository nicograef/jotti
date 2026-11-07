import { describe, expect, it } from 'vitest'

import type { Produkt } from '@/lib/produktSchemas'
import { testPosition } from '@/test/fixtures'

import {
  aufrundenVorschlaege,
  calculateTotalPrice,
  calculateZahlungsbetraege,
  selectPositionen,
  toBestellungData,
} from './drawerUtils'

describe('selectPositionen', () => {
  const positionen = [
    testPosition({ positionId: 'aaa-001', menge: 5 }),
    testPosition({ positionId: 'aaa-002', menge: 3 }),
    testPosition({ positionId: 'aaa-003', menge: 2 }),
  ]

  it('returns only positionen with selected quantity > 0, at that quantity', () => {
    const result = selectPositionen(positionen, {
      'aaa-001': 2,
      'aaa-002': 0,
      'aaa-003': 1,
    })

    expect(result).toEqual([
      { ...positionen[0], menge: 2 },
      { ...positionen[2], menge: 1 },
    ])
  })

  it('ignores selection keys that do not match any position', () => {
    expect(selectPositionen(positionen, { 'zzz-999': 5 })).toEqual([])
  })
})

describe('calculateTotalPrice', () => {
  it.each([
    [[], 0],
    [[{ einzelpreisCents: 200, menge: 3 }], 600],
    [
      [
        { einzelpreisCents: 350, menge: 2 },
        { einzelpreisCents: 250, menge: 1 },
      ],
      950,
    ],
  ])('sums %j to %i cents', (items, expected) => {
    expect(calculateTotalPrice(items)).toBe(expected)
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

describe('aufrundenVorschlaege', () => {
  it('liefert 1-€- und 5-€-Vorschlag echt über einem Zwischenbetrag', () => {
    expect(aufrundenVorschlaege(1230)).toEqual([1300, 1500])
  })

  it('rundet einen glatten Euro-Betrag auf den nächsten 1-€- und 5-€-Schritt', () => {
    expect(aufrundenVorschlaege(1300)).toEqual([1400, 1500])
  })

  it('ersetzt den Doppelgänger, wenn 1-€- und 5-€-Vorschlag zusammenfallen', () => {
    expect(aufrundenVorschlaege(450)).toEqual([500, 1000])
  })
})

describe('toBestellungData', () => {
  const pommes: Produkt = {
    id: 1,
    name: 'Pommes',
    kategorie: 'essen',
    steuersatz: 'ermaessigt',
    pfand: false,
    status: 'active',
    varianten: [
      {
        id: 7,
        name: 'mit Ketchup',
        preisCents: 300,
        status: 'active',
        pfandVarianteId: null,
      },
    ],
  }

  it('sets the receipt preview name to {Produkt} {Variante}', () => {
    const { receiptItems } = toBestellungData([pommes], { 7: 2 })

    expect(receiptItems).toEqual([
      { id: '7', name: 'Pommes mit Ketchup', einzelpreisCents: 300, menge: 2 },
    ])
  })

  it('ignores variants without a selected quantity', () => {
    expect(toBestellungData([pommes], {}).receiptItems).toEqual([])
  })
})
