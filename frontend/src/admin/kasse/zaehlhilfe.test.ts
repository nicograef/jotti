import { describe, expect, it } from 'vitest'

import { summeAusStueckzahlen } from './zaehlhilfe'

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
})
