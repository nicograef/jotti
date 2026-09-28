import { act, cleanup, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Seite } from '@/lib/reload'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { renderHookWithBackend } from '@/test/render'

import { useVersion } from './use-version'
import {
  RELOAD_VERMERK_SCHLUESSEL,
  useVersionsGuard,
} from './use-versions-guard'

// The client runs as a real release; /health reports `serverVersion`.
const CLIENT = 'v1.2.3'
let serverVersion = CLIENT

// Waits until the server version has arrived, so every assertion sees the
// guard's answer to it rather than the state before the first response.
async function renderGuard() {
  const view = renderHookWithBackend(
    () => ({ zustand: useVersionsGuard(CLIENT), version: useVersion() }),
    new FakeBackend().respond('health', () => ({ version: serverVersion })),
  )
  await waitFor(() => {
    expect(view.result.current.version).toBe(serverVersion)
  })
  return view
}

// The next poll reports the new server version.
async function meldeServerVersion(
  view: Awaited<ReturnType<typeof renderGuard>>,
  version: string,
) {
  serverVersion = version
  await act(() => view.queryClient.refetchQueries())
  await waitFor(() => {
    expect(view.result.current.version).toBe(version)
  })
}

let neuLaden: ReturnType<typeof vi.spyOn>

beforeEach(() => {
  serverVersion = CLIENT
  sessionStorage.clear()
  VorgangsRegisterSingleton.zuruecksetzen()
  neuLaden = vi.spyOn(Seite, 'neuLaden').mockImplementation(() => undefined)
})

// Ohne `globals: true` registriert Testing Library kein Auto-Cleanup. Ein
// stehen gebliebener Hook bliebe am Vorgangs-Register abonniert.
afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('useVersionsGuard', () => {
  it('lädt bei einer Abweichung mit leerem Register sofort neu', async () => {
    serverVersion = 'v1.2.4'

    await renderGuard()

    expect(neuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.4')
  })

  it('lädt bei gleicher Version nicht neu', async () => {
    serverVersion = 'v1.2.3'

    const { result } = await renderGuard()

    expect(neuLaden).not.toHaveBeenCalled()
    expect(result.current.zustand).toBe('aus')
  })

  // Ein Ausfall (useVersion liefert undefined) ist kein Versionswechsel.
  it('lädt ohne erfolgreich beantwortete Abfrage nicht neu', async () => {
    // The query cache logs every failed query; that log marks the settled failure.
    const fehlerLog = vi
      .spyOn(console, 'error')
      .mockImplementation(() => undefined)
    const { result } = renderHookWithBackend(
      () => useVersionsGuard(CLIENT),
      new FakeBackend().fail('health'),
    )
    await waitFor(() => {
      expect(fehlerLog).toHaveBeenCalled()
    })

    expect(neuLaden).not.toHaveBeenCalled()
    expect(result.current).toBe('aus')
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()
  })

  it('hält den Reload zurück, solange ein Vorgang offen ist', async () => {
    VorgangsRegisterSingleton.anmelden()
    serverVersion = 'v1.2.4'

    const { result } = await renderGuard()

    expect(neuLaden).not.toHaveBeenCalled()
    expect(result.current.zustand).toBe('wartet')
  })

  it('lädt von selbst neu, sobald der letzte Vorgang abgeschlossen ist', async () => {
    VorgangsRegisterSingleton.anmelden()
    VorgangsRegisterSingleton.anmelden()
    serverVersion = 'v1.2.4'

    await renderGuard()

    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })
    expect(neuLaden).not.toHaveBeenCalled()

    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })
    expect(neuLaden).toHaveBeenCalledTimes(1)
  })

  // Zwischen Auslösen und Entladen läuft die Anwendung weiter: Weitere
  // Abfragen melden dieselbe Abweichung. Ein zweiter Reload darf nicht folgen.
  it('löst innerhalb eines Seitenlebens genau einen Reload aus', async () => {
    serverVersion = 'v1.2.4'

    const { rerender } = await renderGuard()
    expect(neuLaden).toHaveBeenCalledTimes(1)

    rerender()
    act(() => {
      VorgangsRegisterSingleton.anmelden()
    })
    act(() => {
      VorgangsRegisterSingleton.abmelden()
    })

    expect(neuLaden).toHaveBeenCalledTimes(1)
  })

  it('lädt nach einem wirkungslosen Reload nicht erneut, sondern meldet die Bremse', async () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    serverVersion = 'v1.2.4'

    const { result } = await renderGuard()

    expect(neuLaden).not.toHaveBeenCalled()
    expect(result.current.zustand).toBe('gebremst')
    // Erst die Einigkeit mit dem Server löst den Vermerk ein.
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.4')
  })

  it('bleibt gebremst, auch wenn der Vermerk auf eine andere als die gemeldete Version zeigt', async () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    serverVersion = 'v1.2.5'

    const { result } = await renderGuard()

    expect(neuLaden).not.toHaveBeenCalled()
    expect(result.current.zustand).toBe('gebremst')
  })

  // Nach Rollback oder Vorwärts-Korrektur trägt der Client eine andere als die
  // vermerkte Zielversion. Bliebe der Vermerk liegen, bliebe die Bremse gezogen.
  it('löst den Vermerk auch bei anderer Version ein, sobald Client und Server einig sind', async () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.4')
    serverVersion = 'v1.2.3'

    const view = await renderGuard()
    const { result } = view

    expect(result.current.zustand).toBe('aus')
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()

    // Ohne Rücknahme des eingefrorenen Flags bliebe es hier bei `gebremst` —
    // im wochenlang offenen Tab der installierten App für immer.
    await meldeServerVersion(view, 'v1.2.5')

    expect(neuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.5')
  })

  it('löst einen erreichten Vermerk ein und ist danach wieder scharf', async () => {
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, 'v1.2.3')
    serverVersion = 'v1.2.3'

    const view = await renderGuard()

    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBeNull()
    expect(neuLaden).not.toHaveBeenCalled()

    await meldeServerVersion(view, 'v1.2.5')

    expect(neuLaden).toHaveBeenCalledTimes(1)
    expect(sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)).toBe('v1.2.5')
  })
})
