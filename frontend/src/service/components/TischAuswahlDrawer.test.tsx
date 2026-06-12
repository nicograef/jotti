import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  MEINE_TISCHE_STATE_KEY,
} from '../table/hooks'
import { TischAuswahlDrawer } from './TischAuswahlDrawer'

vi.mock('react-router', () => ({
  useNavigate: () => vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// vaul's Drawer depends on browser APIs unavailable in jsdom — render children inline.
vi.mock('@/components/ui/drawer', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return {
    Drawer: Passthrough,
    DrawerContent: Passthrough,
    DrawerHeader: Passthrough,
    DrawerTitle: Passthrough,
  }
})

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

vi.mock('../table/TischBackend', () => ({
  TischBackend: class {
    favoritHinzufuegen = vi.fn().mockResolvedValue(undefined)
    favoritEntfernen = vi.fn().mockResolvedValue(undefined)
  },
}))

vi.mock('../table/hooks', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../table/hooks')>()
  return {
    ...actual,
    useAktiveTischeMitFavoriten: () => ({
      tische: [{ id: 1, name: 'Stammtisch', istFavorit: false, saldoCents: 0 }],
    }),
  }
})

afterEach(() => {
  cleanup()
})

function renderDrawer() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: 0 }, mutations: { retry: 0 } },
  })
  const invalidate = vi.spyOn(queryClient, 'invalidateQueries')

  render(
    <QueryClientProvider client={queryClient}>
      <TischAuswahlDrawer open={true} onOpenChange={vi.fn()} />
    </QueryClientProvider>,
  )

  return { invalidate }
}

describe('TischAuswahlDrawer', () => {
  it('invalidiert nach Favoriten-Toggle beide Query-Caches', async () => {
    const user = userEvent.setup()
    const { invalidate } = renderDrawer()

    await user.click(
      screen.getByRole('button', {
        name: 'Stammtisch zu Favoriten hinzufügen',
      }),
    )

    await waitFor(() => {
      expect(invalidate).toHaveBeenCalledWith({
        queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
      })
      expect(invalidate).toHaveBeenCalledWith({
        queryKey: [MEINE_TISCHE_STATE_KEY],
      })
    })
  })
})
