import { describe, expect, it } from 'vitest'

import { KassensitzungSchema, KassensitzungStatus } from './Kassensitzung'

const baseSitzung = {
  zNr: 1,
  datum: '2026-07-05',
  bezeichnung: 'Sommerfest',
}

describe('KassensitzungSchema', () => {
  it.each([
    KassensitzungStatus.OFFEN,
    KassensitzungStatus.WIRD_ABGESCHLOSSEN,
    KassensitzungStatus.ABGESCHLOSSEN,
  ])('akzeptiert den Status %s', (status) => {
    const result = KassensitzungSchema.safeParse({ ...baseSitzung, status })

    expect(result.success).toBe(true)
  })

  it('lehnt einen unbekannten Status ab', () => {
    const result = KassensitzungSchema.safeParse({
      ...baseSitzung,
      status: 'storniert',
    })

    expect(result.success).toBe(false)
  })
})
