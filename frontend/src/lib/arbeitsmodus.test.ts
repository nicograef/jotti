import { beforeEach, describe, expect, it } from 'vitest'

import { getArbeitsmodus, setArbeitsmodus } from './arbeitsmodus'

describe('arbeitsmodus', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('liefert Tischservice ohne gespeicherte Präferenz', () => {
    expect(getArbeitsmodus()).toBe('tischservice')
  })

  it('liefert Tischservice bei ungültigem gespeichertem Wert', () => {
    localStorage.setItem('jotti-arbeitsmodus', 'unfug')

    expect(getArbeitsmodus()).toBe('tischservice')
  })

  it('liefert den gesetzten Modus zurück und persistiert ihn', () => {
    setArbeitsmodus('direktverkauf')

    // Von außen beobachtbar: der Wert überlebt ein erneutes Lesen.
    expect(getArbeitsmodus()).toBe('direktverkauf')
    expect(localStorage.getItem('jotti-arbeitsmodus')).toBe('direktverkauf')
  })
})
