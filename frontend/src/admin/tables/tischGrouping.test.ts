import { describe, expect, it } from 'vitest'

import { type Tisch } from './Tisch'
import { gruppiereTische } from './tischGrouping'

function tisch(id: number, name: string, saldoCents = 0): Tisch {
  return {
    id,
    name,
    status: 'active',
    saldoCents,
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
  }
}

describe('gruppiereTische', () => {
  it('groups tische by the prefix before the trailing number', () => {
    const gruppen = gruppiereTische([
      tisch(1, 'Zelt 1'),
      tisch(2, 'Zelt 2'),
      tisch(3, 'Biergarten 1'),
    ])

    expect(gruppen.map((g) => g.name)).toEqual(['Zelt', 'Biergarten'])
    expect(gruppen[0].tische.map((t) => t.name)).toEqual(['Zelt 1', 'Zelt 2'])
    expect(gruppen[1].tische.map((t) => t.name)).toEqual(['Biergarten 1'])
  })

  it('sorts tische within a group numerically, not lexically', () => {
    const gruppen = gruppiereTische([
      tisch(1, 'Tisch 10'),
      tisch(2, 'Tisch 2'),
      tisch(3, 'Tisch 1'),
    ])

    expect(gruppen).toHaveLength(1)
    expect(gruppen[0].tische.map((t) => t.name)).toEqual([
      'Tisch 1',
      'Tisch 2',
      'Tisch 10',
    ])
  })

  it('collects tische without a trailing number under "Weitere", always last', () => {
    const gruppen = gruppiereTische([
      tisch(1, 'Eingang'),
      tisch(2, 'Zelt 1'),
      tisch(3, 'Theke'),
    ])

    expect(gruppen.map((g) => g.name)).toEqual(['Zelt', 'Weitere'])
    expect(gruppen[1].tische.map((t) => t.name)).toEqual(['Eingang', 'Theke'])
  })

  it('treats a purely numeric name (no prefix) as "Weitere"', () => {
    const gruppen = gruppiereTische([tisch(1, '12'), tisch(2, 'Zelt 1')])

    expect(gruppen.map((g) => g.name)).toEqual(['Zelt', 'Weitere'])
    expect(gruppen[1].tische.map((t) => t.name)).toEqual(['12'])
  })

  it('keeps group order by first appearance in the input', () => {
    const gruppen = gruppiereTische([
      tisch(1, 'Halle 1'),
      tisch(2, 'Zelt 1'),
      tisch(3, 'Halle 2'),
    ])

    expect(gruppen.map((g) => g.name)).toEqual(['Halle', 'Zelt'])
    expect(gruppen[0].tische.map((t) => t.name)).toEqual(['Halle 1', 'Halle 2'])
  })

  it('returns an empty array for no tische', () => {
    expect(gruppiereTische([])).toEqual([])
  })
})
