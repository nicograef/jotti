import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { belegDruckenMitNachfassen } from './beleg'

describe('belegDruckenMitNachfassen', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('liefert eingereiht sofort ohne Nachfassen', async () => {
    const anfordern = vi.fn().mockResolvedValue('eingereiht')

    const status = await belegDruckenMitNachfassen(anfordern)

    expect(status).toBe('eingereiht')
    expect(anfordern).toHaveBeenCalledTimes(1)
  })

  it('fasst bei ausstehender Signatur im 1,5-Sekunden-Takt nach', async () => {
    const anfordern = vi
      .fn()
      .mockResolvedValueOnce('ausstehend')
      .mockResolvedValueOnce('ausstehend')
      .mockResolvedValue('eingereiht')

    const ergebnis = belegDruckenMitNachfassen(anfordern)
    await vi.advanceTimersByTimeAsync(3_000)

    expect(await ergebnis).toBe('eingereiht')
    expect(anfordern).toHaveBeenCalledTimes(3)
  })

  it('gibt nach dem Nachfass-Fenster ausstehend zurück', async () => {
    const anfordern = vi.fn().mockResolvedValue('ausstehend')

    const ergebnis = belegDruckenMitNachfassen(anfordern)
    await vi.advanceTimersByTimeAsync(10_000)

    expect(await ergebnis).toBe('ausstehend')
    // Erstanfrage plus sechs Nachfass-Versuche (~9 Sekunden).
    expect(anfordern).toHaveBeenCalledTimes(7)
  })
})
