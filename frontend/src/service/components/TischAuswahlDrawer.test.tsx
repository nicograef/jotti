import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

vi.mock('../table/TischBackend', () => ({
  TischBackend: class {
    favoritHinzufuegen = vi.fn().mockResolvedValue(undefined)
    favoritEntfernen = vi.fn().mockResolvedValue(undefined)
  },
}))

let mockTische = [
  { id: 1, name: 'Stammtisch', istFavorit: false, saldoCents: 0 },
]

vi.mock('../table/hooks', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../table/hooks')>()
  return {
    ...actual,
    useAktiveTischeMitFavoriten: () => ({ tische: mockTische }),
  }
})

afterEach(() => {
  cleanup()
  mockTische = [{ id: 1, name: 'Stammtisch', istFavorit: false, saldoCents: 0 }]
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
  it('zeigt die Tisch-Liste im DrawerBody ohne eigenes Suchfeld', () => {
    renderDrawer()

    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    expect(body).not.toBeNull()
    expect(body).toContainElement(screen.getByText('Stammtisch'))
    // Die Suche liegt jetzt auf der Hauptseite — der Drawer hat kein Suchfeld.
    expect(screen.queryByPlaceholderText('Tisch suchen...')).toBeNull()
  })

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

  it('sortiert Favoriten zuerst, dann Saldo absteigend, dann Name', () => {
    mockTische = [
      { id: 1, name: 'Bar', istFavorit: false, saldoCents: 500 },
      { id: 2, name: 'Zelt', istFavorit: true, saldoCents: 100 },
      { id: 3, name: 'Ausschank', istFavorit: false, saldoCents: 500 },
    ]
    renderDrawer()

    const namen = screen
      .getAllByText(/^(Bar|Zelt|Ausschank)$/)
      .map((el) => el.textContent)
    expect(namen).toEqual(['Zelt', 'Ausschank', 'Bar'])
  })
})
