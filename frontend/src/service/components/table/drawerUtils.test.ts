import { describe, expect, it } from 'vitest'

import type { Position } from '@/service/table/Bestellung'

import { calculateTotalPrice, selectPositionen } from './drawerUtils'

describe('selectPositionen', () => {
  const positionen: Position[] = [
    {
      positionId: 'aaa-001',
      varianteId: 1,
      produktName: 'Bratwurst',
      varianteName: 'Normal',
      kategorie: 'essen',
      einzelpreis: 350,
      menge: 5,
    },
    {
      positionId: 'aaa-002',
      varianteId: 4,
      produktName: 'Pommes',
      varianteName: 'Klein',
      kategorie: 'essen',
      einzelpreis: 250,
      menge: 3,
    },
    {
      positionId: 'aaa-003',
      varianteId: 31,
      produktName: 'Softdrinks',
      varianteName: 'Cola',
      kategorie: 'getraenk',
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
