import { describe, expect, it } from 'vitest'

import { formatCents } from './utils'

describe('formatCents', () => {
  it('formats standard amounts', () => {
    expect(formatCents(1250)).toBe('12,50')
  })

  it('formats zero', () => {
    expect(formatCents(0)).toBe('0,00')
  })

  it('formats single-digit cents with padding', () => {
    expect(formatCents(5)).toBe('0,05')
  })

  it('formats amounts under one euro', () => {
    expect(formatCents(99)).toBe('0,99')
  })

  it('formats exact euro amounts', () => {
    expect(formatCents(500)).toBe('5,00')
  })

  it('formats large values', () => {
    expect(formatCents(999999)).toBe('9999,99')
  })

  it('formats negative values', () => {
    expect(formatCents(-350)).toBe('-3,50')
  })

  it('formats ten cents', () => {
    expect(formatCents(10)).toBe('0,10')
  })
})
