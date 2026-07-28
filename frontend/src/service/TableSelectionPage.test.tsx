import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { MEINE_TISCHE_STATE_KEY } from './table/hooks'
import type {
  AktiverTischMitFavorit,
  EigeneUebersicht,
  TischSession,
} from './table/Tisch'
import { TableSelectionPage } from './TableSelectionPage'

const navigate = vi.fn()
vi.mock('react-router', () => ({
  useNavigate: () => navigate,
}))

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

// Die Seite läuft gegen die echten Query-Hooks; nur das Backend ist ersetzt.
// Nur so ist prüfbar, dass ein gescheitertes Erstladen (leerer Cache) und ein
// gescheiterter Hintergrund-Refetch (gefüllter Cache) verschieden aussehen.
const {
  getMeineTischeState,
  getAktiveTischeMitFavoriten,
  getEigeneUebersicht,
} = vi.hoisted(() => ({
  getMeineTischeState: vi.fn<() => Promise<TischSession[]>>(),
  getAktiveTischeMitFavoriten: vi.fn<() => Promise<AktiverTischMitFavorit[]>>(),
  getEigeneUebersicht: vi.fn<() => Promise<EigeneUebersicht>>(),
}))

vi.mock('./table/TischBackend', () => ({
  TischBackend: class {
    getMeineTischeState = getMeineTischeState
    getAktiveTischeMitFavoriten = getAktiveTischeMitFavoriten
    getEigeneUebersicht = getEigeneUebersicht
  },
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

const leereUebersicht: EigeneUebersicht = {
  anzahlBestellungen: 0,
  bestellungenCents: 0,
  anzahlZahlungen: 0,
  zahlungenCents: 0,
  anzahlRuecknahmen: 0,
  ruecknahmenCents: 0,
  abzugebenCents: 0,
}

const stammtisch = tischSession(1, 'Stammtisch', true)
const alleTische: AktiverTischMitFavorit[] = [
  { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
  { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
]

beforeEach(() => {
  getMeineTischeState.mockResolvedValue([])
  getAktiveTischeMitFavoriten.mockResolvedValue([])
  getEigeneUebersicht.mockResolvedValue(leereUebersicht)
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <TableSelectionPage />
    </QueryClientProvider>,
  )
  return { queryClient }
}

describe('TableSelectionPage', () => {
  // Eine fehlgeschlagene Query darf nie als Leerzustand erscheinen: „Keine
  // Tische markiert" ist für einen Helfer mit Favoriten eine Falschaussage.
  it('zeigt bei gescheitertem Erstladen der eigenen Tische den Ladefehler statt des Leerzustands', async () => {
    getMeineTischeState.mockRejectedValue(new Error('Netzabbruch'))
    renderPage()

    expect(
      await screen.findByText('Meine Tische konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Keine Tische markiert')).not.toBeInTheDocument()
    // Die intakte Übersicht bleibt stehen — der Fehler ersetzt nur seinen
    // eigenen Bereich.
    expect(screen.getByText('Übersichtskarten')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Alle Tische' }),
    ).toBeInTheDocument()
  })

  // Die eigene Übersicht ist rein informativ; ihr Ausfall darf keinen Einstieg
  // in einen Tisch blockieren.
  it('lässt bei gescheiterter eigener Übersicht den Einstieg in einen Tisch offen', async () => {
    getEigeneUebersicht.mockRejectedValue(new Error('Netzabbruch'))
    getMeineTischeState.mockResolvedValue([stammtisch])
    getAktiveTischeMitFavoriten.mockResolvedValue(alleTische)
    const user = userEvent.setup()
    renderPage()

    expect(
      await screen.findByText('Eigene Übersicht konnte nicht geladen werden'),
    ).toBeInTheDocument()
    // 0 · 0,00 € wäre als Fehlerdarstellung irreführend.
    expect(screen.queryByText('Übersichtskarten')).not.toBeInTheDocument()

    // Favoritenliste, Suchfeld und Fußzeilen-Button bleiben bedienbar …
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Alle Tische' }),
    ).toBeInTheDocument()

    // … und ein Suchtreffer öffnet den Tisch weiterhin.
    await user.type(screen.getByPlaceholderText(/Tisch suchen/), 'Bar')
    await user.click(screen.getByRole('button', { name: /^Bar\b.*€/ }))
    expect(navigate).toHaveBeenCalledWith('/service/tische/2')
  })

  it('zeigt bei gescheiterter Tischsuche den Ladefehler statt „kein Treffer"', async () => {
    getAktiveTischeMitFavoriten.mockRejectedValue(new Error('Netzabbruch'))
    getMeineTischeState.mockResolvedValue([stammtisch])
    const user = userEvent.setup()
    renderPage()

    // Die Favoritenliste steht, obwohl die Suche fehlt.
    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()

    await user.type(screen.getByPlaceholderText(/Tisch suchen/), 'Bar')

    expect(
      screen.getByText('Tischsuche konnte nicht geladen werden'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(/Kein aktiver Tisch passt zu/),
    ).not.toBeInTheDocument()
  })

  // Der Kern der Unterscheidung: Mit gefülltem Cache ist ein gescheiterter
  // Refetch kein Grund, die Ansicht wegzureißen — die Meldung trägt der
  // zentrale Fehler-Toast.
  it('lässt die Tische stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getMeineTischeState
      .mockResolvedValueOnce([stammtisch])
      .mockRejectedValue(new Error('Netzabbruch'))
    const { queryClient } = renderPage()

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()

    await act(async () => {
      await queryClient.refetchQueries({ queryKey: [MEINE_TISCHE_STATE_KEY] })
    })
    // Die Query steht auf „error", die Seite hat die Aktualisierung verarbeitet
    // — erst danach ist die Anzeige aussagekräftig.
    await waitFor(() => {
      expect(queryClient.getQueryState([MEINE_TISCHE_STATE_KEY])?.status).toBe(
        'error',
      )
    })

    expect(getMeineTischeState).toHaveBeenCalledTimes(2)
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Meine Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('lädt über „Erneut versuchen" alle drei Queries neu und zeigt danach die Tische', async () => {
    getMeineTischeState
      .mockRejectedValueOnce(new Error('Netzabbruch'))
      .mockResolvedValue([stammtisch])
    const user = userEvent.setup()
    renderPage()

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    // Ein Netzabbruch trifft alle drei Queries; der Wiederholversuch lädt
    // deshalb auch die beiden Bereiche neu, die gerade keinen Fehler zeigen.
    expect(getMeineTischeState).toHaveBeenCalledTimes(2)
    expect(getAktiveTischeMitFavoriten).toHaveBeenCalledTimes(2)
    expect(getEigeneUebersicht).toHaveBeenCalledTimes(2)

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Meine Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt ohne Fehler und ohne Favoriten den Leerzustand', async () => {
    renderPage()

    expect(await screen.findByText('Keine Tische markiert')).toBeInTheDocument()
    expect(
      screen.queryByText('Meine Tische konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei leerem Suchfeld die Favoriten („Meine Tische")', async () => {
    getMeineTischeState.mockResolvedValue([stammtisch])
    getAktiveTischeMitFavoriten.mockResolvedValue(alleTische)
    renderPage()

    expect(await screen.findByText('Noch offen · 1')).toBeInTheDocument()
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    // Nicht favorisierte Tische erscheinen ohne Suche nicht.
    expect(screen.queryByText('Bar')).not.toBeInTheDocument()
  })

  it('findet einen nicht favorisierten aktiven Tisch über die Hauptsuche', async () => {
    getMeineTischeState.mockResolvedValue([stammtisch])
    getAktiveTischeMitFavoriten.mockResolvedValue(alleTische)
    const user = userEvent.setup()
    renderPage()

    // Suchfeld und Treffer werden genau so angesprochen wie im e2e-Helper
    // (support/servicekraft.ts oeffneTisch): Platzhalter-Teilstring plus
    // Button-Name „<Name> … <Saldo> €".
    await user.type(await screen.findByPlaceholderText(/Tisch suchen/), 'Bar')

    // Der nicht favorisierte Tisch erscheint als Treffer …
    const treffer = screen.getByRole('button', { name: /^Bar\b.*€/ })
    expect(treffer).toBeInTheDocument()
    // … und ein Treffer öffnet den Tisch direkt.
    await user.click(treffer)
    expect(navigate).toHaveBeenCalledWith('/service/tische/2')
  })

  it('meldet, wenn kein aktiver Tisch zur Suche passt', async () => {
    getMeineTischeState.mockResolvedValue([stammtisch])
    getAktiveTischeMitFavoriten.mockResolvedValue([alleTische[0]])
    const user = userEvent.setup()
    renderPage()

    await user.type(
      await screen.findByPlaceholderText('Tisch suchen — Name oder Nummer'),
      'Zelt',
    )

    expect(screen.getByText(/Kein aktiver Tisch passt zu/)).toBeInTheDocument()
  })
})
