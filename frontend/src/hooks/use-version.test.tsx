import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { createQueryClient } from '@/lib/queryClient'

import { useVersion, VERSIONSABFRAGE_INTERVALL_MS } from './use-version'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

const health = vi.hoisted(() => ({ getVersion: vi.fn() }))

vi.mock('@/lib/HealthBackend', () => ({
  HealthBackend: class {
    getVersion = health.getVersion
  },
}))

// Der Hook läuft gegen den echten QueryClient der Anwendung — nur so ist
// belegt, dass sein meta-Flag den globalen Fehler-Toast wirklich unterdrückt.
function renderUseVersion() {
  const queryClient = createQueryClient()
  return renderHook(() => useVersion(), {
    wrapper: ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
  })
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
    health.getVersion.mockResolvedValue('v1.2.3')
    vi.useFakeTimers()

    const { result } = renderUseVersion()

    await vi.advanceTimersByTimeAsync(0)
    expect(result.current).toBe('v1.2.3')
    expect(health.getVersion).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(VERSIONSABFRAGE_INTERVALL_MS)
    expect(health.getVersion).toHaveBeenCalledTimes(2)

    await vi.advanceTimersByTimeAsync(VERSIONSABFRAGE_INTERVALL_MS)
    expect(health.getVersion).toHaveBeenCalledTimes(3)
  })

  // Im Funkloch schlägt die Abfrage dauerhaft fehl. Ein Toast alle 30 Sekunden
  // wäre eine Verschlechterung: Niemand kann darauf reagieren. Der 4xx-Fehler
  // steht schon beim ersten Versuch fest — die Wiederholungspolitik ist hier
  // nicht der Gegenstand.
  it('erzeugt bei einem Fehlschlag keinen Fehler-Toast', async () => {
    health.getVersion.mockRejectedValue(new BackendError(400, 'bad_request'))

    const { result } = renderUseVersion()

    await waitFor(() => {
      expect(health.getVersion).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(console.error).toHaveBeenCalled()
    })
    expect(result.current).toBeUndefined()
    expect(toast.error).not.toHaveBeenCalled()
  })
})
