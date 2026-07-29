import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { useMengen } from './use-mengen'

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

describe('useMengen im Vorgangs-Register', () => {
  it('meldet eine getroffene Auswahl als einen offenen Vorgang', () => {
    const { result } = renderHook(() => useMengen<number>())

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    act(() => {
      result.current.add(1)
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Eine zweite Menge ist derselbe Vorgang, kein zweiter.
    act(() => {
      result.current.add(2)
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
  })

  it('gibt den Vorgang frei, wenn die Auswahl wieder leer ist', () => {
    const { result } = renderHook(() => useMengen<number>())

    act(() => {
      result.current.add(1)
    })
    act(() => {
      result.current.remove(1)
    })

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('gibt den Vorgang beim Zurücksetzen der Auswahl frei', () => {
    const { result } = renderHook(() => useMengen<number>())

    act(() => {
      result.current.add(1)
    })
    act(() => {
      result.current.reset()
    })

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('gibt den Vorgang beim Aushängen frei', () => {
    const { result, unmount } = renderHook(() => useMengen<number>())

    act(() => {
      result.current.add(1)
    })
    unmount()

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
