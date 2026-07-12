import { describe, expect, it } from 'vitest'

import { tseAmpel } from './tseAmpel'
import type { TSESignaturQueue, TSEStatus } from './TSEBackend'

function queue(overrides: Partial<TSESignaturQueue> = {}): TSESignaturQueue {
  return {
    offeneAuftraege: 3,
    fehlgeschlageneAuftraege: 0,
    letzterFehler: '',
    rueckstandSekunden: 12,
    signaturenProMinute: 20,
    signierdauerP95Sekunden: 1.2,
    ...overrides,
  }
}

const konfiguriert: TSEStatus = { umgebung: 'LIVE', istKonfiguriert: true }

describe('tseAmpel', () => {
  it('meldet den grünen Normalzustand bei konfigurierter TSE ohne Rückstand', () => {
    const ampel = tseAmpel(konfiguriert, false, queue())
    expect(ampel.fehler).toBe(false)
    expect(ampel.ueberschrift).toBe('Ja — TSE signiert normal')
  })

  it('meldet einen Fehler, wenn die TSE nicht konfiguriert ist', () => {
    const ampel = tseAmpel(
      { umgebung: '', istKonfiguriert: false },
      false,
      queue(),
    )
    expect(ampel.fehler).toBe(true)
    expect(ampel.ueberschrift).toBe('TSE ist nicht eingerichtet')
  })

  it('meldet einen Fehler bei fehlgeschlagenen Signaturen', () => {
    const ampel = tseAmpel(
      konfiguriert,
      false,
      queue({ fehlgeschlageneAuftraege: 2 }),
    )
    expect(ampel.fehler).toBe(true)
    expect(ampel.ueberschrift).toBe('TSE braucht Aufmerksamkeit')
  })

  it('meldet einen Fehler bei Rückstand ab der 60-s-Schwelle, nicht darunter', () => {
    expect(
      tseAmpel(konfiguriert, false, queue({ rueckstandSekunden: 59 })).fehler,
    ).toBe(false)
    expect(
      tseAmpel(konfiguriert, false, queue({ rueckstandSekunden: 60 })).fehler,
    ).toBe(true)
  })

  it('behandelt eine noch ladende TSE nicht als Fehler', () => {
    const ampel = tseAmpel(undefined, true, undefined)
    expect(ampel.fehler).toBe(false)
  })
})
