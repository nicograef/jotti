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
// Fehlerzustand je Query der Seite, damit jeder der drei Pfade einzeln geprüft
// werden kann.
let fehler = {
  meineTische: false,
  alleTische: false,
  uebersicht: false,
}

const reloadMeineTische = vi.fn()
const reloadAlleTische = vi.fn()
const reloadUebersicht = vi.fn()

vi.mock('./table/hooks', () => ({
  useMeineTischeState: () => ({
    tische: meineTische,
    isPending: false,
    isError: fehler.meineTische,
    refetch: reloadMeineTische,
  }),
  useAktiveTischeMitFavoriten: () => ({
    tische: alleTische,
    isError: fehler.alleTische,
    refetch: reloadAlleTische,
  }),
  useEigeneUebersicht: () => ({
    uebersicht: {
      anzahlBestellungen: 0,
      bestellungenCents: 0,
      anzahlZahlungen: 0,
      zahlungenCents: 0,
      anzahlRuecknahmen: 0,
      ruecknahmenCents: 0,
      abzugebenCents: 0,
    },
    isPending: false,
    isError: fehler.uebersicht,
    refetch: reloadUebersicht,
  }),
}))

// Kindkomponenten auf Stubs reduzieren: der Test prüft die Such-/Favoriten-Logik
// der Seite, nicht das Rendern der Karten oder des Drawers. Der Karten-Stub
// trägt einen Marker, damit prüfbar bleibt, dass er bei einem Ladefehler nicht
// erscheint.
vi.mock('./components/EigeneUebersicht', () => ({
  EigeneUebersichtKarten: () => <div>Übersichtskarten</div>,
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
  fehler = { meineTische: false, alleTische: false, uebersicht: false }
})

describe('TableSelectionPage', () => {
  // Eine fehlgeschlagene Query darf nie als Leerzustand erscheinen: „Keine
  // Tische markiert" ist für einen Helfer mit Favoriten eine Falschaussage.
  it.each([
    ['meineTische' as const],
    ['alleTische' as const],
    ['uebersicht' as const],
  ])('zeigt bei Fehler der Query %s den Ladefehler', (query) => {
    fehler[query] = true
    render(<TableSelectionPage />)

    expect(
      screen.getByText('Tische konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Keine Tische markiert')).not.toBeInTheDocument()
    // Auch die Übersichtskarten bleiben aus — 0 · 0,00 € wäre als
    // Fehlerdarstellung genauso irreführend wie der leere Tischbereich.
    expect(screen.queryByText('Übersichtskarten')).not.toBeInTheDocument()
  })

  it('lädt über „Erneut versuchen" neu und zeigt danach die Tische', async () => {
    fehler.meineTische = true
    const user = userEvent.setup()
    const { rerender } = render(<TableSelectionPage />)

    await user.click(screen.getByRole('button', { name: 'Erneut versuchen' }))
    expect(reloadMeineTische).toHaveBeenCalledTimes(1)

    // Der Refetch war erfolgreich: die Query liefert wieder Daten.
    fehler.meineTische = false
    meineTische = [tischSession(1, 'Stammtisch', true)]
    rerender(<TableSelectionPage />)

    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt ohne Fehler und ohne Favoriten den Leerzustand', () => {
    render(<TableSelectionPage />)

    expect(screen.getByText('Keine Tische markiert')).toBeInTheDocument()
    expect(
      screen.queryByText('Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

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
