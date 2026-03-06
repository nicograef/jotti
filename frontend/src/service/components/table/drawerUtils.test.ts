import { describe, expect, it } from 'vitest'

import type { LineItem } from '@/service/table/Order'

import { calculateTotalPrice, selectVariants } from './drawerUtils'

describe('selectVariants', () => {
  const variants: LineItem[] = [
    { id: 1, name: 'Bratwurst', priceCents: 350, quantity: 5 },
    { id: 2, name: 'Pommes', priceCents: 250, quantity: 3 },
    { id: 3, name: 'Cola 0,3l', priceCents: 200, quantity: 2 },
  ]

  it('returns only variants with selected quantity > 0', () => {
    const result = selectVariants(variants, { 1: 2, 3: 1 })

    expect(result).toEqual([
      { id: 1, name: 'Bratwurst', priceCents: 350, quantity: 2 },
      { id: 3, name: 'Cola 0,3l', priceCents: 200, quantity: 1 },
    ])
  })

  it('returns empty array when no variants are selected', () => {
    expect(selectVariants(variants, {})).toEqual([])
  })

  it('returns empty array for empty input', () => {
    expect(selectVariants([], {})).toEqual([])
  })

  it('handles single item selection', () => {
    const result = selectVariants(variants, { 2: 1 })

    expect(result).toEqual([
      { id: 2, name: 'Pommes', priceCents: 250, quantity: 1 },
    ])
  })

  it('ignores selection keys that do not match any variant', () => {
    const result = selectVariants(variants, { 999: 5 })

    expect(result).toEqual([])
  })

  it('filters out variants where selected quantity is 0', () => {
    const result = selectVariants(variants, { 1: 0, 2: 3 })

    expect(result).toEqual([
      { id: 2, name: 'Pommes', priceCents: 250, quantity: 3 },
    ])
  })
})

describe('calculateTotalPrice', () => {
  it('calculates total for multiple items', () => {
    const items: LineItem[] = [
      { id: 1, name: 'Bratwurst', priceCents: 350, quantity: 2 },
      { id: 2, name: 'Pommes', priceCents: 250, quantity: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(950)
  })

  it('returns 0 for empty array', () => {
    expect(calculateTotalPrice([])).toBe(0)
  })

  it('handles single item', () => {
    const items: LineItem[] = [
      { id: 1, name: 'Cola', priceCents: 200, quantity: 3 },
    ]

    expect(calculateTotalPrice(items)).toBe(600)
  })

  it('handles zero-cent variants', () => {
    const items: LineItem[] = [
      { id: 1, name: 'Wasser', priceCents: 0, quantity: 5 },
      { id: 2, name: 'Cola', priceCents: 200, quantity: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(200)
  })

  it('handles quantity of 1', () => {
    const items: LineItem[] = [
      { id: 1, name: 'Bier', priceCents: 300, quantity: 1 },
    ]

    expect(calculateTotalPrice(items)).toBe(300)
  })
})
