import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AktiverTischMitFavorit, TischSession } from './table/Tisch'
import { TableSelectionPage } from './TableSelectionPage'

const navigate = vi.fn()
vi.mock('react-router', () => ({
  useNavigate: () => navigate,
}))

let meineTische: TischSession[] = []
let alleTische: AktiverTischMitFavorit[] = []

vi.mock('./table/hooks', () => ({
  useMeineTischeState: () => ({ tische: meineTische, isPending: false }),
  useAktiveTischeMitFavoriten: () => ({ tische: alleTische }),
  useEigeneUebersicht: () => ({
    uebersicht: {
      anzahlBestellungen: 0,
      bestellungenCents: 0,
      anzahlZahlungen: 0,
      zahlungenCents: 0,
    },
    isPending: false,
  }),
}))

// Kindkomponenten auf Stubs reduzieren: der Test prüft die Such-/Favoriten-Logik
// der Seite, nicht das Rendern der Karten oder des Drawers.
vi.mock('./components/EigeneUebersicht', () => ({
  EigeneUebersichtKarten: () => null,
}))
vi.mock('./components/MeinTischCard', () => ({
  MeinTischCard: ({ state }: { state: TischSession }) => (
    <div>{state.tischName}</div>
  ),
}))
vi.mock('./components/TischAuswahlDrawer', () => ({
  TischAuswahlDrawer: () => null,
}))

function tischSession(
  tischId: number,
  tischName: string,
  offen: boolean,
): TischSession {
  return {
    tischId,
    tischName,
    saldoCents: offen ? 500 : 0,
    unbezahltePositionen: offen
      ? // Nur die Länge zählt für die Offen/Erledigt-Gruppierung.
        ([{ positionId: 'p1' }] as TischSession['unbezahltePositionen'])
      : [],
    fuerMichErledigt: !offen,
  }
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  meineTische = []
  alleTische = []
})

describe('TableSelectionPage', () => {
  it('zeigt bei leerem Suchfeld die Favoriten („Meine Tische")', () => {
    meineTische = [tischSession(1, 'Stammtisch', true)]
    alleTische = [
      { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
      { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
    ]
    render(<TableSelectionPage />)

    expect(screen.getByText('Noch offen · 1')).toBeInTheDocument()
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    // Nicht favorisierte Tische erscheinen ohne Suche nicht.
    expect(screen.queryByText('Bar')).not.toBeInTheDocument()
  })

  it('findet einen nicht favorisierten aktiven Tisch über die Hauptsuche', async () => {
    meineTische = [tischSession(1, 'Stammtisch', true)]
    alleTische = [
      { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
      { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
    ]
    const user = userEvent.setup()
    render(<TableSelectionPage />)

    // Suchfeld und Treffer werden genau so angesprochen wie im e2e-Helper
    // (support/servicekraft.ts oeffneTisch): Platzhalter-Teilstring plus
    // Button-Name „<Name> … <Saldo> €".
    await user.type(screen.getByPlaceholderText(/Tisch suchen/), 'Bar')

    // Der nicht favorisierte Tisch erscheint als Treffer …
    const treffer = screen.getByRole('button', { name: /^Bar\b.*€/ })
    expect(treffer).toBeInTheDocument()
    // … und ein Treffer öffnet den Tisch direkt.
    await user.click(treffer)
    expect(navigate).toHaveBeenCalledWith('/service/tische/2')
  })

  it('meldet, wenn kein aktiver Tisch zur Suche passt', async () => {
    meineTische = [tischSession(1, 'Stammtisch', true)]
    alleTische = [
      { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
    ]
    const user = userEvent.setup()
    render(<TableSelectionPage />)

    await user.type(
      screen.getByPlaceholderText('Tisch suchen — Name oder Nummer'),
      'Zelt',
    )

    expect(screen.getByText(/Kein aktiver Tisch passt zu/)).toBeInTheDocument()
  })
})
