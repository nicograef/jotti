import { act, waitFor } from '@testing-library/react'
import { toast } from 'sonner'
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  type MockInstance,
  vi,
} from 'vitest'

import type { DownloadResult } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { renderHookWithBackend } from '@/test/render'

import {
  useAbgeschlosseneKassensitzungen,
  useDsfinvkExport,
  useLiveReporting,
} from './hooks'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

let klick: MockInstance<() => void>

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
  // jsdom implements neither object URLs nor navigation on an anchor click.
  URL.createObjectURL = vi.fn(() => 'blob:dsfinvk')
  URL.revokeObjectURL = vi.fn()
  klick = vi
    .spyOn(HTMLAnchorElement.prototype, 'click')
    .mockImplementation(() => undefined)
})

afterEach(() => {
  vi.clearAllMocks()
  vi.restoreAllMocks()
})

// The fake answers downloads at once; a held promise keeps the export running
// until the test settles it.
function haltenderDownload(fake: FakeBackend, archiv: Promise<DownloadResult>) {
  vi.spyOn(fake, 'download').mockReturnValue(archiv)
}

describe('useDsfinvkExport im Vorgangs-Register', () => {
  it('meldet den laufenden Export und gibt ihn nach dem Download frei', async () => {
    const fake = new FakeBackend()
    let liefern!: (archiv: DownloadResult) => void
    haltenderDownload(
      fake,
      new Promise((resolve) => {
        liefern = resolve
      }),
    )
    const { result } = renderHookWithBackend(() => useDsfinvkExport(), fake)

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    act(() => {
      result.current.exportieren(1)
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
    })

    await act(async () => {
      liefern({ blob: new Blob(), filename: 'dsfinvk.zip' })
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
    })
    expect(klick).toHaveBeenCalledOnce()
    expect(klick.mock.contexts[0]).toHaveProperty('download', 'dsfinvk.zip')
  })

  it('gibt den Export auch nach einem Fehlschlag frei', async () => {
    const fake = new FakeBackend()
    let scheitern!: (fehler: Error) => void
    haltenderDownload(
      fake,
      new Promise((_, reject) => {
        scheitern = reject
      }),
    )
    const { result } = renderHookWithBackend(() => useDsfinvkExport(), fake)

    act(() => {
      result.current.exportieren(1)
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
    })

    await act(async () => {
      scheitern(new Error('Netzabbruch'))
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
    })
  })
})

// Ohne eigenen Fehlerzustand ist der Toast das einzige Signal: Die Seite sähe
// sonst aus wie „keine Kassensitzung offen". Die Live-Abfrage läuft im
// Intervall wie use-version, darf deren Toast-Unterdrückung aber nicht erben.
describe('Reporting-Abfragen bei Ladefehler', () => {
  it.each<[string, () => unknown]>([
    ['admin/get-live-reporting', useLiveReporting],
    [
      'admin/get-abgeschlossene-kassensitzungen',
      useAbgeschlosseneKassensitzungen,
    ],
  ])('%s meldet den Fehler per Toast', async (endpoint, hook) => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    renderHookWithBackend(hook, new FakeBackend().fail(endpoint))

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.',
        { id: 'query-fehler' },
      )
    })
  })
})
