import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { TSE_KONFIGURATION_KEY } from './hooks'
import type { TSEKonfiguration } from './TSEBackend'
import { TSEKonfigurationSection } from './TSEKonfigurationSection'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Backend', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/Backend')>()),
  BackendSingleton: {},
}))

// Die Sektion läuft gegen den echten Query-Hook; nur das Backend ist ersetzt.
// Nur so ist prüfbar, dass ein gescheitertes Erstladen (leerer Cache) und ein
// gescheiterter Hintergrund-Refetch (gefüllter Cache) verschieden aussehen.
const { getTSEKonfiguration } = vi.hoisted(() => ({
  getTSEKonfiguration: vi.fn<() => Promise<TSEKonfiguration>>(),
}))

vi.mock('./TSEBackend', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./TSEBackend')>()),
  TSEBackend: class {
    getTSEKonfiguration = getTSEKonfiguration
  },
}))

const konfiguration: TSEKonfiguration = {
  apiKeyGesetzt: true,
  apiSecretGesetzt: true,
  tssId: '123e4567-e89b-12d3-a456-426614174000',
  clientId: 'KASSE-1',
  istKonfiguriert: true,
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

function renderSection() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <TSEKonfigurationSection />
    </QueryClientProvider>,
  )
  return { queryClient }
}

describe('TSEKonfigurationSection', () => {
  it('zeigt bei gescheitertem Erstladen eine Meldung statt des Formulars', async () => {
    getTSEKonfiguration.mockRejectedValue(new Error('Netzabbruch'))
    renderSection()

    expect(
      await screen.findByText('Fehler beim Laden der TSE-Konfiguration.'),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText('API-Key *')).not.toBeInTheDocument()
  })

  // Der Fokus-Refetch (Rückkehr aus dem fiskaly-Portal) scheitert; die
  // eingetippten Zugangsdaten liegen in lokalem State und wären mit dem Unmount
  // des Formulars verloren. Die Meldung trägt der zentrale Fehler-Toast.
  it('lässt Formular und Eingaben stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getTSEKonfiguration
      .mockResolvedValueOnce(konfiguration)
      .mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    const { queryClient } = renderSection()

    await user.type(await screen.findByLabelText('API-Key *'), 'geheim')

    await act(async () => {
      await queryClient.refetchQueries()
    })
    await waitFor(() => {
      expect(queryClient.getQueryState([TSE_KONFIGURATION_KEY])?.status).toBe(
        'error',
      )
    })

    expect(getTSEKonfiguration).toHaveBeenCalledTimes(2)
    expect(
      screen.queryByText('Fehler beim Laden der TSE-Konfiguration.'),
    ).not.toBeInTheDocument()
    expect(screen.getByLabelText('API-Key *')).toHaveValue('geheim')
  })
})
