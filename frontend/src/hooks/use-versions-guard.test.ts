import { act, cleanup, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { seiteNeuLaden } from '@/lib/reload'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import {
  RELOAD_VERMERK_SCHLUESSEL,
  useVersionsGuard,
} from './use-versions-guard'

const versionState = vi.hoisted<{ version: string | undefined }>(() => ({
  version: undefined,
}))

vi.mock('@/hooks/use-version', () => ({
  useVersion: () => versionState.version,
}))

vi.mock('@/lib/reload', () => ({ seiteNeuLaden: vi.fn() }))

// Die Clientversion ist im Test der Default `dev` (define in vitest.config.ts)
// und damit kein Release — ohne echte Release-Version auf beiden Seiten meldet
// istVersionsabweichung grundsätzlich nichts.
vi.mock('@/lib/version', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/version')>()),
  CLIENT_VERSION: 'v1.2.3',
}))

beforeEach(() => {
  versionState.version = undefined
  sessionStorage.clear()
  VorgangsRegisterSingleton.zuruecksetzen()
  vi.mocked(seiteNeuLaden).mockClear()
})

// Ohne `globals: true` registriert Testing Library sein Auto-Cleanup nicht.
// Ein stehen gebliebener Hook bliebe am Vorgangs-Register abonniert und würde
// beim nächsten Test mitentscheiden — samt eigenem Reload.
afterEach(() => {
  cleanup()
})

describe('useVersionsGuard', () => {
  it('lädt bei einer Abweichung mit leerem Register sofort neu', () => {
    versionState.version = 'v1.2.4'

    renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.4')
  })

  it('lädt bei gleicher Version nicht neu', () => {
    versionState.version = 'v1.2.3'

    const { result } = renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('aus')
  })

  // Ein Serverneustart lässt die Abfrage scheitern; useVersion liefert dann
  // undefined. Ein Ausfall ist kein Versionswechsel und darf nichts erzwingen.
  it('lädt ohne erfolgreich beantwortete Abfrage nicht neu', () => {
    const { result } = renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('aus')
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()
  })

  it('hält den Reload zurück, solange ein Vorgang offen ist', () => {
    VorgangsRegisterSingleton.anmelden()
    versionState.version = 'v1.2.4'

    const { result } = renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('wartet')
  })

  it('lädt von selbst neu, sobald der letzte Vorgang abgeschlossen ist', () => {
    VorgangsRegisterSingleton.anmelden()
    VorgangsRegisterSingleton.anmelden()
    versionState.version = 'v1.2.4'

    renderHook(() => useVersionsGuard())

    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })
    expect(seiteNeuLaden).not.toHaveBeenCalled()

    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })
    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
  })

  // Zwischen dem Auslösen und dem tatsächlichen Entladen der Seite
  // läuft die Anwendung weiter: Weitere Abfragen melden dieselbe Abweichung,
  // und ein Vorgang kann noch aufgehen und wieder zugehen. Ein zweiter Reload
  // darf daraus nicht folgen.
  it('löst innerhalb eines Seitenlebens genau einen Reload aus', () => {
    versionState.version = 'v1.2.4'

    const { rerender } = renderHook(() => useVersionsGuard())
    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)

    rerender()
    act(() => {
      VorgangsRegisterSingleton.anmelden()
    })
    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
  })

  it('lädt nach einem wirkungslosen Reload nicht erneut, sondern meldet die Bremse', () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    versionState.version = 'v1.2.4'

    const { result } = renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('gebremst')
    // Der Vermerk bleibt stehen: Erst die Einigkeit mit dem Server löst ihn
    // ein, und die gibt es hier gerade nicht.
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.4')
  })

  it('bleibt gebremst, auch wenn der Vermerk auf eine andere als die gemeldete Version zeigt', () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    versionState.version = 'v1.2.5'

    const { result } = renderHook(() => useVersionsGuard())

    expect(seiteNeuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('gebremst')
  })

  // Nach einem misslungenen Update rollt der Betreiber zurück oder korrigiert
  // vorwärts — der Client trägt dann eine andere als die vermerkte
  // Zielversion, ist mit dem Server aber einig. Bliebe der Vermerk liegen,
  // wäre die Bremse für die restliche Lebensdauer des Tabs gezogen.
  it('löst den Vermerk auch bei anderer Version ein, sobald Client und Server einig sind', () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    versionState.version = 'v1.2.3'

    const { result, rerender } = renderHook(() => useVersionsGuard())

    expect(result.current).toBe('aus')
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()

    // Ohne Rücknahme des eingefrorenen Flags bliebe es hier bei `gebremst` —
    // im wochenlang offenen Tab der installierten App für immer.
    versionState.version = 'v1.2.5'
    rerender()

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.5')
  })

  it('löst einen erreichten Vermerk ein und ist danach wieder scharf', () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.3')
    versionState.version = 'v1.2.3'

    const { rerender } = renderHook(() => useVersionsGuard())

    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()
    expect(seiteNeuLaden).not.toHaveBeenCalled()

    versionState.version = 'v1.2.5'
    rerender()

    expect(seiteNeuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.5')
  })
})
