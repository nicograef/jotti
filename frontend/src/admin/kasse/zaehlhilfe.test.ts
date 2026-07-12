import { describe, expect, it } from 'vitest'

import { NENNWERTE_CENTS, summeAusStueckzahlen } from './zaehlhilfe'

describe('summeAusStueckzahlen', () => {
  it('summiert eine leere Erfassung zu 0', () => {
    expect(summeAusStueckzahlen({})).toBe(0)
  })

  it('multipliziert Nennwert mit Stückzahl und summiert über alle Nennwerte', () => {
    // 3×200 € (60000) + 2×50 € (10000) + 5×2 € (1000) + 7×1 ct (7) = 71007.
    expect(
      summeAusStueckzahlen({
        20000: 3,
        5000: 2,
        200: 5,
        1: 7,
      }),
    ).toBe(71007)
  })

  it('rechnet jeden einzelnen Nennwert korrekt um', () => {
    for (const nennwert of NENNWERTE_CENTS) {
      expect(summeAusStueckzahlen({ [nennwert]: 4 })).toBe(nennwert * 4)
    }
  })

  it('ignoriert negative, null- und nicht-ganze Stückzahlen', () => {
    // Negative, 0- und Bruch-Stückzahlen sind strukturell number, werden aber
    // als 0 gewertet; nur 4×10 ct (40) zählt.
    expect(
      summeAusStueckzahlen({
        200: -3,
        100: 0,
        50: 2.5,
        10: 4,
      }),
    ).toBe(40)
  })
})
