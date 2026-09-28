import { waitFor } from '@testing-library/react'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderHookWithBackend } from '@/test/render'

import { useVersion, VERSIONSABFRAGE_INTERVALL_MS } from './use-version'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

// Der Hook läuft gegen den echten QueryClient der Anwendung — nur so ist
// belegt, dass sein meta-Flag den globalen Fehler-Toast wirklich unterdrückt.
function renderUseVersion(backend: FakeBackend) {
  return renderHookWithBackend(() => useVersion(), backend)
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => undefined)
})

afterEach(() => {
  vi.clearAllMocks()
  vi.restoreAllMocks()
  vi.useRealTimers()
})

describe('useVersion', () => {
  it('fragt die Version alle 30 Sekunden erneut ab', async () => {
    const backend = new FakeBackend().respond('health', { version: 'v1.2.3' })
    vi.useFakeTimers()

    const { result } = renderUseVersion(backend)

    await vi.advanceTimersByTimeAsync(0)
    expect(result.current).toBe('v1.2.3')
    expect(backend.bodies('health')).toHaveLength(1)

    await vi.advanceTimersByTimeAsync(VERSIONSABFRAGE_INTERVALL_MS)
    expect(backend.bodies('health')).toHaveLength(2)

    await vi.advanceTimersByTimeAsync(VERSIONSABFRAGE_INTERVALL_MS)
    expect(backend.bodies('health')).toHaveLength(3)
  })

  // Im Funkloch schlägt die Abfrage dauerhaft fehl; ein Toast alle 30 Sekunden
  // wäre eine Verschlechterung, niemand kann darauf reagieren.
  it('erzeugt bei einem Fehlschlag keinen Fehler-Toast', async () => {
    const backend = new FakeBackend().fail('health')

    const { result } = renderUseVersion(backend)

    await waitFor(() => {
      expect(backend.bodies('health')).toHaveLength(1)
    })
    await waitFor(() => {
      expect(console.error).toHaveBeenCalled()
    })
    expect(result.current).toBeUndefined()
    expect(toast.error).not.toHaveBeenCalled()
  })
})
