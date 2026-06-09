import { describe, expect, it } from 'vitest'

import type { Bestellung, Position } from '@/service/table/Bestellung'
import type { Stornierung } from '@/service/table/Stornierung'
import type { Zahlung } from '@/service/table/Zahlung'

import {
  calculateTotalPrice,
  calculateZahlungsbetraege,
  getStornierbarePositionen,
  getUmbuchbarePositionen,
  selectPositionen,
} from './drawerUtils'

const positionA: Position = {
  positionId: 'aaa-001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'ermaessigt',
  einzelpreis: 350,
  menge: 5,
}

const positionB: Position = {
  positionId: 'aaa-002',
  varianteId: 4,
  produktName: 'Pommes',
  varianteName: 'Klein',
  kategorie: 'essen',
  steuersatz: 'ermaessigt',
  einzelpreis: 250,
  menge: 3,
}

function createBestellung(positionen: Position[]): Bestellung {
  return {
    id: 'bestellung-1',
    userId: 1,
    tischId: 11,
    positionen,
    gesamtPreisCents: 2500,
    kommentar: '',
    aufgenommenAm: '2026-06-09T10:00:00Z',
  }
}

function createStornierung(positionen: Position[]): Stornierung {
  return {
    id: 'storno-1',
    userId: 1,
    tischId: 11,
    positionen,
    gesamtStornierungCents: 500,
    kommentar: 'Storno',
    storniertAm: '2026-06-09T10:05:00Z',
  }
}

function createZahlung(positionen: Position[]): Zahlung {
  return {
    id: 'zahlung-1',
    userId: 1,
    tischId: 11,
    positionen,
    gesamtZahlungCents: 1200,
    kommentar: '',
    kassiertAm: '2026-06-09T10:06:00Z',
  }
}

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

describe('getStornierbarePositionen', () => {
  it('subtracts already cancelled amounts for each position', () => {
    const bestellung = createBestellung([{ ...positionA }, { ...positionB }])
    const historie = [
      bestellung,
      createStornierung([
        {
          ...positionA,
          menge: 2,
        },
      ]),
    ]

    expect(getStornierbarePositionen(bestellung, historie)).toEqual([
      { ...positionA, menge: 3 },
      { ...positionB, menge: 3 },
    ])
  })

  it('returns empty when all quantities are cancelled', () => {
    const bestellung = createBestellung([{ ...positionA, menge: 2 }])
    const historie = [
      bestellung,
      createStornierung([{ ...positionA, menge: 2 }]),
    ]

    expect(getStornierbarePositionen(bestellung, historie)).toEqual([])
  })
})

describe('getUmbuchbarePositionen', () => {
  it('subtracts both cancelled and paid partial quantities', () => {
    const bestellung = createBestellung([{ ...positionA }, { ...positionB }])
    const historie = [
      bestellung,
      createStornierung([
        {
          ...positionA,
          menge: 1,
        },
      ]),
      createZahlung([
        {
          ...positionA,
          menge: 2,
        },
        {
          ...positionB,
          menge: 1,
        },
      ]),
    ]

    expect(getUmbuchbarePositionen(bestellung, historie)).toEqual([
      { ...positionA, menge: 2 },
      { ...positionB, menge: 2 },
    ])
  })

  it('returns empty when all quantities are paid or cancelled', () => {
    const bestellung = createBestellung([{ ...positionA, menge: 3 }])
    const historie = [
      bestellung,
      createStornierung([{ ...positionA, menge: 1 }]),
      createZahlung([{ ...positionA, menge: 2 }]),
    ]

    expect(getUmbuchbarePositionen(bestellung, historie)).toEqual([])
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
