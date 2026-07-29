import { beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from './VorgangsRegister'

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

describe('VorgangsRegister', () => {
  it('zählt Anmeldungen hoch und Abmeldungen wieder herunter', () => {
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    VorgangsRegisterSingleton.anmelden()
    VorgangsRegisterSingleton.anmelden()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(2)

    VorgangsRegisterSingleton.abmelden()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    VorgangsRegisterSingleton.abmelden()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('benachrichtigt Interessenten bei jeder Änderung', () => {
    const interessent = vi.fn()
    VorgangsRegisterSingleton.abonnieren(interessent)

    VorgangsRegisterSingleton.anmelden()
    VorgangsRegisterSingleton.abmelden()

    expect(interessent).toHaveBeenCalledTimes(2)
  })

  it('beendet das Abo über den Rückgabewert', () => {
    const interessent = vi.fn()
    const abbestellen = VorgangsRegisterSingleton.abonnieren(interessent)

    abbestellen()
    VorgangsRegisterSingleton.anmelden()

    expect(interessent).not.toHaveBeenCalled()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
  })

  // Ohne diesen Weg trüge ein Test die offenen Vorgänge des vorherigen in den
  // nächsten — der Zähler ist Modulzustand und lebt länger als ein Test.
  it('lässt sich zwischen Tests auf null zurücksetzen', () => {
    VorgangsRegisterSingleton.anmelden()
    VorgangsRegisterSingleton.anmelden()

    VorgangsRegisterSingleton.zuruecksetzen()

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
