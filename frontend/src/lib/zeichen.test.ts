import { describe, expect, it } from 'vitest'
import { z } from 'zod'

import { maxZeichen, minZeichen } from './zeichen'

describe('zeichen', () => {
  it('zählt Umlaute und Emoji als je ein Zeichen', () => {
    const max = maxZeichen(z.string(), 100)
    const min = minZeichen(z.string(), 3)

    expect(max.safeParse('ä'.repeat(100)).success).toBe(true)
    expect(max.safeParse('ä'.repeat(101)).success).toBe(false)
    expect(min.safeParse('😀😀').success).toBe(false)
    expect(min.safeParse('😀😀😀').success).toBe(true)
  })
})
