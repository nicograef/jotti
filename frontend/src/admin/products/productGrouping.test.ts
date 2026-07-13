import { describe, expect, it } from 'vitest'

import type { DruckstationConfig } from '../settings/DruckstationBackend'
import {
  gemeinsamerSteuersatz,
  groupProdukteByKategorie,
  kategorieZusatz,
  produktUnterzeile,
} from './productGrouping'
import type { Produkt } from './Produkt'

function produkt(overrides: Partial<Produkt>): Produkt {
  return {
    id: 1,
    name: 'Produkt',
    kategorie: 'essen',
    steuersatz: 'ermaessigt',
    status: 'active',
    varianten: [],
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
    ...overrides,
  }
}

describe('groupProdukteByKategorie', () => {
  it('orders sections Essen, Getränke, Sonstiges and drops empty categories', () => {
    const produkte = [
      produkt({ id: 1, name: 'Cola', kategorie: 'getraenk' }),
      produkt({ id: 2, name: 'Pommes', kategorie: 'essen' }),
    ]

    const gruppen = groupProdukteByKategorie(produkte)

    expect(gruppen.map((g) => g.kategorie)).toEqual(['essen', 'getraenk'])
    expect(gruppen[0].produkte[0].name).toBe('Pommes')
  })
})

describe('gemeinsamerSteuersatz', () => {
  it('returns the shared Steuersatz when all products agree', () => {
    expect(
      gemeinsamerSteuersatz([
        produkt({ steuersatz: 'regel' }),
        produkt({ steuersatz: 'regel' }),
      ]),
    ).toBe('regel')
  })

  it('returns null when products disagree', () => {
    expect(
      gemeinsamerSteuersatz([
        produkt({ steuersatz: 'regel' }),
        produkt({ steuersatz: 'ermaessigt' }),
      ]),
    ).toBeNull()
  })
})

describe('kategorieZusatz', () => {
  const stationen: DruckstationConfig[] = [
    { kategorie: 'essen', druckerIp: '192.168.0.5', bonmodus: 'pro_position' },
    { kategorie: 'getraenk', druckerIp: '', bonmodus: 'pro_position' },
  ]

  it('includes Steuersatz and station hint when a printer is configured', () => {
    const zusatz = kategorieZusatz(
      'essen',
      [produkt({ kategorie: 'essen', steuersatz: 'ermaessigt' })],
      stationen,
    )
    expect(zusatz).toContain('Ermäßigter Steuersatz')
    expect(zusatz).toContain('Bons an Station „Essen"')
  })

  it('omits the station hint when no printer is configured', () => {
    const zusatz = kategorieZusatz(
      'getraenk',
      [produkt({ kategorie: 'getraenk', steuersatz: 'regel' })],
      stationen,
    )
    expect(zusatz).not.toContain('Bons an Station')
    expect(zusatz).toContain('Regelsteuersatz')
  })
})

describe('produktUnterzeile', () => {
  it('counts products and variants with correct plural forms', () => {
    const produkte = [
      produkt({
        id: 1,
        varianten: [
          {
            id: 1,
            name: 'klein',
            preisCents: 100,
            status: 'active',
            createdAt: '2026-07-01T10:00:00Z',
            updatedAt: '2026-07-01T10:00:00Z',
          },
        ],
      }),
    ]
    expect(produktUnterzeile(produkte)).toBe(
      '1 Produkt · 1 Variante · Änderungen wirken sofort auf allen Service-Handys',
    )
  })
})
