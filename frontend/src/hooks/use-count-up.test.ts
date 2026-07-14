import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useCountUp } from './use-count-up'

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('useCountUp', () => {
  it('liefert beim ersten Rendern den Ausgangswert ohne Animation', () => {
    const { result } = renderHook(({ ziel }) => useCountUp(ziel), {
      initialProps: { ziel: 1250 },
    })

    expect(result.current).toBe(1250)
  })

  it('liefert ohne Animationsumgebung bei Änderung sofort den Endwert', () => {
    // jsdom kennt kein window.matchMedia — der Hook stuft die Umgebung als
    // nicht animierbar ein und springt direkt auf den Zielwert.
    const { result, rerender } = renderHook(({ ziel }) => useCountUp(ziel), {
      initialProps: { ziel: 1250 },
    })

    act(() => {
      rerender({ ziel: 2000 })
    })

    expect(result.current).toBe(2000)
  })

  it('endet bei Änderung in einer Animationsumgebung exakt am Zielwert', () => {
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({ matches: false }))
    vi.spyOn(performance, 'now').mockReturnValue(0)
    let frames: FrameRequestCallback[] = []
    vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => {
      frames.push(cb)
      return frames.length
    })
    vi.stubGlobal('cancelAnimationFrame', vi.fn())

    const flush = (t: number) => {
      const anstehend = frames
      frames = []
      act(() => {
        anstehend.forEach((cb) => {
          cb(t)
        })
      })
    }

    const { result, rerender } = renderHook(({ ziel }) => useCountUp(ziel), {
      initialProps: { ziel: 1000 },
    })

    act(() => {
      rerender({ ziel: 2001 })
    })

    // Auf halber Strecke zählt der Hook, hat den Zielwert aber noch nicht erreicht.
    flush(350)
    expect(result.current).toBeGreaterThan(1000)
    expect(result.current).toBeLessThan(2001)

    // Nach Ablauf der Dauer steht exakt der Zielwert (auch bei ungeradem Delta).
    flush(700)
    expect(result.current).toBe(2001)
  })
})
