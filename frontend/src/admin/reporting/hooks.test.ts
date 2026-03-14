import { describe, expect, it } from 'vitest'

import { combineDateTime } from './hooks'

describe('combineDateTime', () => {
  it('returns null when date is undefined', () => {
    expect(combineDateTime(undefined, '10:30')).toBeNull()
  })

  it('combines date and time correctly', () => {
    const date = new Date(2026, 2, 14) // March 14, 2026
    const result = combineDateTime(date, '10:30')

    expect(result).not.toBeNull()
    if (result === null) return

    expect(result.getHours()).toBe(10)
    expect(result.getMinutes()).toBe(30)
    expect(result.getSeconds()).toBe(0)
    expect(result.getMilliseconds()).toBe(0)
  })

  it('sets midnight for time 00:00', () => {
    const date = new Date(2026, 2, 14)
    const result = combineDateTime(date, '00:00')

    expect(result).not.toBeNull()
    if (result === null) return

    expect(result.getHours()).toBe(0)
    expect(result.getMinutes()).toBe(0)
  })

  it('preserves date components', () => {
    const date = new Date(2026, 2, 14) // March 14, 2026
    const result = combineDateTime(date, '08:00')

    expect(result).not.toBeNull()
    if (result === null) return

    expect(result.getFullYear()).toBe(2026)
    expect(result.getMonth()).toBe(2) // March (0-indexed)
    expect(result.getDate()).toBe(14)
  })
})
