import { describe, expect, it } from 'vitest'
import { z } from 'zod'

import { kuerzeAufZeichen, maxZeichen } from './zeichen'

describe('zeichen', () => {
  it('zählt Umlaute und Emoji als je ein Zeichen', () => {
    const schema = maxZeichen(z.string(), 100)

    expect(schema.safeParse('ä'.repeat(100)).success).toBe(true)
    expect(schema.safeParse('ä'.repeat(101)).success).toBe(false)
    expect(kuerzeAufZeichen('a😀b', 2)).toBe('a😀')
  })
})
