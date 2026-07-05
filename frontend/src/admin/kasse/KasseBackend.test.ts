import { describe, expect, it } from 'vitest'

import { KassensitzungEroeffnenSchema } from './KasseBackend'

describe('KassensitzungEroeffnenSchema', () => {
  it('akzeptiert 0 Cent als Anfangsbestand', () => {
    const result = KassensitzungEroeffnenSchema.safeParse({
      bezeichnung: 'Sommerfest',
      betragCents: 0,
    })

    expect(result.success).toBe(true)
  })

  it('akzeptiert positive Betraege', () => {
    const result = KassensitzungEroeffnenSchema.safeParse({
      bezeichnung: 'Sommerfest',
      betragCents: 15000,
    })

    expect(result.success).toBe(true)
  })

  it('lehnt negative Betraege ab', () => {
    const result = KassensitzungEroeffnenSchema.safeParse({
      bezeichnung: 'Sommerfest',
      betragCents: -1,
    })

    expect(result.success).toBe(false)
  })

  it('lehnt fehlenden Betrag (leeres Feld) ab', () => {
    const result = KassensitzungEroeffnenSchema.safeParse({
      bezeichnung: 'Sommerfest',
    })

    expect(result.success).toBe(false)
  })
})
