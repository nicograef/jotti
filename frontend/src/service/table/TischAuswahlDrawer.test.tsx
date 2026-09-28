import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { MEINE_TISCHE_STATE_KEY } from './hooks'
import type { AktiverTischMitFavorit } from './Tisch'
import { TischAuswahlDrawer } from './TischAuswahlDrawer'

vi.mock('react-router', () => ({
  useNavigate: () => vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

const stammtisch: AktiverTischMitFavorit = {
  id: 1,
  name: 'Stammtisch',
  istFavorit: false,
  saldoCents: 0,
}

function backend(tische: AktiverTischMitFavorit[] = [stammtisch]): FakeBackend {
  return new FakeBackend().respond('service/get-aktive-tische-mit-favoriten', {
    tische,
  })
}

function renderDrawer(fake: FakeBackend = backend()) {
  return renderWithBackend(
    <TischAuswahlDrawer open={true} onOpenChange={vi.fn()} />,
    fake,
  )
}

describe('TischAuswahlDrawer', () => {
  it('zeigt die Tisch-Liste im DrawerBody ohne eigenes Suchfeld', async () => {
    renderDrawer()

    const stammtischZeile = await screen.findByText('Stammtisch')
    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    expect(body).not.toBeNull()
    expect(body).toContainElement(stammtischZeile)
    // Die Suche liegt auf der Hauptseite — der Drawer hat kein Suchfeld.
    expect(screen.queryByPlaceholderText('Tisch suchen...')).toBeNull()
  })

  it('invalidiert nach Favoriten-Toggle beide Query-Caches', async () => {
    let istFavorit = false
    const fake = new FakeBackend()
      .respond('service/get-aktive-tische-mit-favoriten', () => ({
        tische: [{ ...stammtisch, istFavorit }],
      }))
      .respond('service/favorit-hinzufuegen', () => {
        istFavorit = true
        return {}
      })
    const user = userEvent.setup()
    const { queryClient } = renderDrawer(fake)
    // The overview of the user's own tables shares the favourite flag.
    queryClient.setQueryData([MEINE_TISCHE_STATE_KEY], [])

    await user.click(
      await screen.findByRole('button', {
        name: 'Stammtisch zu Favoriten hinzufügen',
      }),
    )

    expect(
      await screen.findByRole('button', {
        name: 'Stammtisch aus Favoriten entfernen',
      }),
    ).toHaveTextContent('★')
    await waitFor(() => {
      expect(
        queryClient.getQueryState([MEINE_TISCHE_STATE_KEY])?.isInvalidated,
      ).toBe(true)
    })
  })

  it('sortiert durchgehend nach Tischname mit numerischem Vergleich, Favoriten nicht vorgezogen', async () => {
    renderDrawer(
      backend([
        { id: 1, name: 'Tisch 10', istFavorit: false, saldoCents: 500 },
        { id: 2, name: 'Tisch 2', istFavorit: true, saldoCents: 100 },
        { id: 3, name: 'Tisch 1', istFavorit: false, saldoCents: 0 },
      ]),
    )

    const namen = (await screen.findAllByText(/^Tisch \d+$/)).map(
      (el) => el.textContent,
    )
    // Numerischer Vergleich („Tisch 2" vor „Tisch 10"); Favoriten stehen nicht vorn.
    expect(namen).toEqual(['Tisch 1', 'Tisch 2', 'Tisch 10'])
  })

  it('zeigt Favoriten-Stern und Saldo pro Zeile weiterhin an', async () => {
    renderDrawer(
      backend([
        { id: 1, name: 'Tisch 2', istFavorit: true, saldoCents: 100 },
        { id: 2, name: 'Tisch 10', istFavorit: false, saldoCents: 500 },
      ]),
    )

    expect(
      await screen.findByRole('button', {
        name: 'Tisch 2 aus Favoriten entfernen',
      }),
    ).toHaveTextContent('★')
    expect(
      screen.getByRole('button', {
        name: 'Tisch 10 zu Favoriten hinzufügen',
      }),
    ).toHaveTextContent('☆')
    expect(screen.getByText(/1,00\s*€/)).toBeInTheDocument()
    expect(screen.getByText(/5,00\s*€/)).toBeInTheDocument()
  })
})

describe('TischAuswahlDrawer bei Ladefehler', () => {
  it('zeigt den Hinweis statt einer leeren Liste', async () => {
    renderDrawer(backend().fail('service/get-aktive-tische-mit-favoriten'))

    expect(
      await screen.findByText('Tische konnten nicht geladen werden'),
    ).toBeInTheDocument()
  })

  it('lädt die Tische über „Erneut versuchen" neu', async () => {
    const getTische = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(400, 'test_fehler')
      })
      .mockReturnValue({ tische: [stammtisch] })
    const user = userEvent.setup()
    renderDrawer(
      backend().respond('service/get-aktive-tische-mit-favoriten', getTische),
    )

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })
})
