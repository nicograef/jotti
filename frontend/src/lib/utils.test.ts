import { describe, expect, it } from 'vitest'

import { formatCents } from './utils'

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
