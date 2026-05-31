import { describe, expect, it } from 'vitest'

import { formatCents, parseCents } from './utils'

describe('formatCents', () => {
  it.each([
    [1250, '12,50'],
    [0, '0,00'],
    [5, '0,05'],
    [99, '0,99'],
    [500, '5,00'],
    [999999, '9999,99'],
    [-350, '-3,50'],
    [10, '0,10'],
  ])('formatCents(%i) = %s', (input, expected) => {
    expect(formatCents(input)).toBe(expected)
  })
})

describe('parseCents', () => {
  it.each([
    ['12,50', 1250],
    ['12.50', 1250],
    ['0,05', 5],
    ['0', 0],
    ['-3,50', -350],
    ['', 0],
    ['abc', 0],
    ['10,', 1000],
  ])('parseCents(%s) = %i', (input, expected) => {
    expect(parseCents(input)).toBe(expected)
  })
})
