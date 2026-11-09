import { describe, expect, it } from 'vitest'

import { createNameSchema } from './nameSchema'

describe('createNameSchema', () => {
  const schema = createNameSchema(100)

  it('akzeptiert einen Namen an beiden Grenzen', () => {
    expect(schema.safeParse('abc').success).toBe(true)
    expect(schema.safeParse('a'.repeat(100)).success).toBe(true)
  })

  it('lehnt einen Namen unter und über den Grenzen ab', () => {
    expect(schema.safeParse('ab').success).toBe(false)
    expect(schema.safeParse('a'.repeat(101)).success).toBe(false)
  })

  it.each([
    [2, false],
    [3, true],
    [100, true],
    [101, false],
  ])(
    'zählt %i Emojis als Codepoints wie das Backend, gültig: %s',
    (anzahl, gueltig) => {
      expect(schema.safeParse('🍺'.repeat(anzahl)).success).toBe(gueltig)
    },
  )

  it('meldet die Grenzen mit den Namensmeldungen', () => {
    expect(schema.safeParse('ab').error?.issues[0]?.message).toBe(
      'Das sieht nicht nach einem echten Namen aus.',
    )
    expect(schema.safeParse('a'.repeat(101)).error?.issues[0]?.message).toBe(
      'Der Name ist zu lang.',
    )
  })

  it('trimmt vor der Prüfung', () => {
    expect(schema.parse('  Zelt 3  ')).toBe('Zelt 3')
    expect(schema.safeParse('   ').success).toBe(false)
    expect(schema.safeParse(` ${'a'.repeat(100)} `).success).toBe(true)
  })

  it('nimmt die Obergrenze als Argument', () => {
    const kurz = createNameSchema(50)

    expect(kurz.safeParse('a'.repeat(50)).success).toBe(true)
    expect(kurz.safeParse('a'.repeat(51)).success).toBe(false)
  })
})
