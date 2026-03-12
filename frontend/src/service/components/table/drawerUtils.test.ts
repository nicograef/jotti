import { describe, expect, it } from 'vitest'

import type { Position } from '@/service/table/Bestellung'

import { calculateTotalPrice, selectPositionen } from './drawerUtils'

describe('selectPositionen', () => {
  const positionen: Position[] = [
    { id: 1, name: 'Bratwurst', preisCents: 350, menge: 5 },
    { id: 2, name: 'Pommes', preisCents: 250, menge: 3 },
    { id: 3, name: 'Cola 0,3l', preisCents: 200, menge: 2 },
  ]

  it('returns only positionen with selected quantity > 0', () => {
    const result = selectPositionen(positionen, { 1: 2, 3: 1 })

    expect(result).toEqual([
      { id: 1, name: 'Bratwurst', preisCents: 350, menge: 2 },
      { id: 3, name: 'Cola 0,3l', preisCents: 200, menge: 1 },
    ])
  })

  it('returns empty array when no positionen are selected', () => {
    expect(selectPositionen(positionen, {})).toEqual([])
  })

  it('returns empty array for empty input', () => {
    expect(selectPositionen([], {})).toEqual([])
  })

  it('handles single item selection', () => {
    const result = selectPositionen(positionen, { 2: 1 })

    expect(result).toEqual([
      { id: 2, name: 'Pommes', preisCents: 250, menge: 1 },
    ])
  })

  it('ignores selection keys that do not match any position', () => {
    const result = selectPositionen(positionen, { 999: 5 })

    expect(result).toEqual([])
  })

  it('filters out positionen where selected quantity is 0', () => {
    const result = selectPositionen(positionen, { 1: 0, 2: 3 })

    expect(result).toEqual([
      { id: 2, name: 'Pommes', preisCents: 250, menge: 3 },
    ])
  })
})

describe('calculateTotalPrice', () => {
  it('calculates total for multiple items', () => {
    const items: Position[] = [
      { id: 1, name: 'Bratwurst', preisCents: 350, menge: 2 },
      { id: 2, name: 'Pommes', preisCents: 250, menge: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(950)
  })

  it('returns 0 for empty array', () => {
    expect(calculateTotalPrice([])).toBe(0)
  })

  it('handles single item', () => {
    const items: Position[] = [
      { id: 1, name: 'Cola', preisCents: 200, menge: 3 },
    ]

    expect(calculateTotalPrice(items)).toBe(600)
  })

  it('handles zero-cent positionen', () => {
    const items: Position[] = [
      { id: 1, name: 'Wasser', preisCents: 0, menge: 5 },
      { id: 2, name: 'Cola', preisCents: 200, menge: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(200)
  })

  it('handles menge of 1', () => {
    const items: Position[] = [
      { id: 1, name: 'Bier', preisCents: 300, menge: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(300)
  })
})
