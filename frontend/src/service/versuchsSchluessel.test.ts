import { describe, expect, it } from 'vitest'

import { mitVersuchsSchluessel } from './versuchsSchluessel'

async function versuch(inhalt: unknown, gelingt: boolean): Promise<string> {
  let schluessel = ''
  await mitVersuchsSchluessel(inhalt, (s) => {
    schluessel = s
    return gelingt ? Promise.resolve() : Promise.reject(new Error('verloren'))
  }).catch(() => undefined)
  return schluessel
}

describe('mitVersuchsSchluessel', () => {
  it('behält den Schlüssel eines Inhalts über andere Versuche hinweg, bis einer gelingt', async () => {
    const erster = await versuch(['a', 1], false)
    const anderer = await versuch(['b', 1], false)
    const wiederholung = await versuch(['a', 1], true)
    const danach = await versuch(['a', 1], true)

    expect(anderer).not.toBe(erster)
    expect(wiederholung).toBe(erster)
    expect(danach).not.toBe(erster)
  })
})
