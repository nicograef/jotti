import { describe, expect, it } from 'vitest'

import { OnetimePasswordSchema } from './identity'

describe('OnetimePasswordSchema', () => {
  it('akzeptiert genau 6 Ziffern', () => {
    expect(OnetimePasswordSchema.safeParse('123456').success).toBe(true)
  })

  it.each(['', '12345', '1234567', 'abcdef', '12ab56', ' 123456 '])(
    'lehnt %o ab (nicht genau 6 Ziffern)',
    (input) => {
      expect(OnetimePasswordSchema.safeParse(input).success).toBe(false)
    },
  )
})
