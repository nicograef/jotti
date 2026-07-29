import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import { useAnzahlOffeneVorgaenge } from './use-anzahl-offene-vorgaenge'
import { useOffenerVorgang } from './use-offener-vorgang'

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

describe('useOffenerVorgang', () => {
  it('hält die Anmeldung genau so lange, wie der Vorgang offen ist', () => {
    const { rerender } = renderHook(
      ({ offen }: { offen: boolean }) => {
        useOffenerVorgang(offen)
      },
      { initialProps: { offen: false } },
    )

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    rerender({ offen: true })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Erneutes Rendern mit unverändertem Zustand meldet nicht ein zweites Mal.
    rerender({ offen: true })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    rerender({ offen: false })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  it('gibt den Vorgang beim Aushängen frei', () => {
    const { unmount } = renderHook(() => {
      useOffenerVorgang(true)
    })

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    unmount()

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})

describe('useAnzahlOffeneVorgaenge', () => {
  it('rendert bei jeder Änderung des Zählers neu', () => {
    const { result } = renderHook(() => useAnzahlOffeneVorgaenge())

    expect(result.current).toBe(0)

    act(() => {
      VorgangsRegisterSingleton.anmelden()
    })
    expect(result.current).toBe(1)

    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })
    expect(result.current).toBe(0)
  })
})
