import { QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { createQueryClient } from '@/lib/queryClient'

import { useOffeneKassensitzung } from './hooks'
import type { OffeneKassensitzung } from './KasseBackend'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

// Der Hook läuft gegen die echte Query-Verdrahtung; nur seine Backend-Klasse
// ist ersetzt. Nur so ist prüfbar, dass ein gescheitertes Erstladen und ein
// gescheiterter Hintergrund-Refetch verschieden aussehen.
const { getOffeneKassensitzung } = vi.hoisted(() => ({
  getOffeneKassensitzung: vi.fn<() => Promise<OffeneKassensitzung | null>>(),
}))

vi.mock('./KasseBackend', () => ({
  KasseBackend: class {
    getOffeneKassensitzung = getOffeneKassensitzung
  },
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  vi.restoreAllMocks()
})

const sitzung: OffeneKassensitzung = {
  zNr: 12,
  datum: '2026-07-11',
  bezeichnung: 'Sommerfest Tag 2',
  status: 'offen',
  eroeffnetAm: '2026-07-11T08:02:00Z',
}

function erzeugeWrapper() {
  const queryClient = createQueryClient()
  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
  return { queryClient, wrapper }
}

describe('useOffeneKassensitzung', () => {
  it('meldet ein gescheitertes Erstladen als Ladefehler', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    getOffeneKassensitzung.mockRejectedValue(new Error('Netzabbruch'))
    const { wrapper } = erzeugeWrapper()

    const { result } = renderHook(() => useOffeneKassensitzung(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoadingError).toBe(true)
    })
    expect(result.current.kassensitzung).toBeNull()
  })

  // Ein gescheiterter Hintergrund-Refetch (react-query refetcht bei jedem
  // Fokuswechsel) darf die Kassentag-Seite nicht wegreißen — samt eingetipptem
  // Ist-Bestand. `isError` wäre hier true, `isLoadingError` bleibt false.
  it('meldet einen gescheiterten Refetch bei gefülltem Cache nicht als Ladefehler', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    getOffeneKassensitzung
      .mockResolvedValueOnce(sitzung)
      .mockRejectedValue(new Error('Netzabbruch'))
    const { queryClient, wrapper } = erzeugeWrapper()

    const { result } = renderHook(() => useOffeneKassensitzung(), { wrapper })
    await waitFor(() => {
      expect(result.current.kassensitzung).toEqual(sitzung)
    })

    await act(async () => {
      await queryClient.refetchQueries()
    })

    await waitFor(() => {
      expect(getOffeneKassensitzung).toHaveBeenCalledTimes(2)
    })
    expect(result.current.isLoadingError).toBe(false)
    expect(result.current.kassensitzung).toEqual(sitzung)
  })
})
