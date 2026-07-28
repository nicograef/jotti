import { QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { createQueryClient } from '@/lib/queryClient'

import type { DruckstationConfig } from './DruckstationBackend'
import { useDruckstationen } from './hooks'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

// Der Hook läuft gegen die echte Query-Verdrahtung; nur seine Backend-Klasse
// ist ersetzt. Nur so ist prüfbar, dass ein gescheitertes Erstladen und ein
// gescheiterter Hintergrund-Refetch verschieden aussehen.
const { getDruckstationen } = vi.hoisted(() => ({
  getDruckstationen: vi.fn<() => Promise<DruckstationConfig[]>>(),
}))

vi.mock('./DruckstationBackend', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./DruckstationBackend')>()),
  DruckstationBackend: class {
    getDruckstationen = getDruckstationen
  },
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  vi.restoreAllMocks()
})

const station: DruckstationConfig = {
  kategorie: 'essen',
  druckerIp: '192.168.1.50',
  bonmodus: 'pro_position',
}

function erzeugeWrapper() {
  const queryClient = createQueryClient()
  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
  return { queryClient, wrapper }
}

describe('useDruckstationen', () => {
  it('meldet ein gescheitertes Erstladen als Ladefehler', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    getDruckstationen.mockRejectedValue(new Error('Netzabbruch'))
    const { wrapper } = erzeugeWrapper()

    const { result } = renderHook(() => useDruckstationen(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoadingError).toBe(true)
    })
    expect(result.current.druckstationen).toEqual([])
  })

  // Ein gescheiterter Hintergrund-Refetch — etwa der nach dem Speichern einer
  // Drucker-IP — darf die Bondrucker-Seite nicht wegreißen: Mit ihr
  // verschwänden die fehlgeschlagenen Druckaufträge samt „Nochmal drucken" und
  // „Verwerfen". `isError` wäre hier true, `isLoadingError` bleibt false.
  it('meldet einen gescheiterten Refetch bei gefülltem Cache nicht als Ladefehler', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    getDruckstationen
      .mockResolvedValueOnce([station])
      .mockRejectedValue(new Error('Netzabbruch'))
    const { queryClient, wrapper } = erzeugeWrapper()

    const { result } = renderHook(() => useDruckstationen(), { wrapper })
    await waitFor(() => {
      expect(result.current.druckstationen).toEqual([station])
    })

    await act(async () => {
      await queryClient.refetchQueries()
    })

    await waitFor(() => {
      expect(getDruckstationen).toHaveBeenCalledTimes(2)
    })
    expect(result.current.isLoadingError).toBe(false)
    expect(result.current.druckstationen).toEqual([station])
  })
})
