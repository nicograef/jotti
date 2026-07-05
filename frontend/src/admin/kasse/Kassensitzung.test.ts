import { describe, expect, it } from 'vitest'

import {
  KassensitzungSchema,
  KassensitzungStatus,
  kassensitzungStatusLabel,
} from './Kassensitzung'

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

describe('kassensitzungStatusLabel', () => {
  it('gibt für jeden Status ein eigenes Symbol und einen Text', () => {
    expect(kassensitzungStatusLabel(KassensitzungStatus.OFFEN)).toEqual({
      symbol: '🟢',
      text: 'offen',
    })
    expect(
      kassensitzungStatusLabel(KassensitzungStatus.WIRD_ABGESCHLOSSEN),
    ).toEqual({ symbol: '🟡', text: 'wird abgeschlossen…' })
    expect(kassensitzungStatusLabel(KassensitzungStatus.ABGESCHLOSSEN)).toEqual(
      {
        symbol: '🔴',
        text: 'abgeschlossen',
      },
    )
  })

  it('zeigt den Zwischenstatus nicht wie abgeschlossen an', () => {
    const zwischen = kassensitzungStatusLabel(
      KassensitzungStatus.WIRD_ABGESCHLOSSEN,
    )
    const abgeschlossen = kassensitzungStatusLabel(
      KassensitzungStatus.ABGESCHLOSSEN,
    )

    expect(zwischen.symbol).not.toBe(abgeschlossen.symbol)
  })
})
