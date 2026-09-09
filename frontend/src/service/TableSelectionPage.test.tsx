import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { AktiverTischMitFavorit, TischSession } from './table/Tisch'
import { TableSelectionPage } from './TableSelectionPage'

const navigate = vi.fn()
vi.mock('react-router', () => ({
  useNavigate: () => navigate,
}))

let meineTische: TischSession[] = []
let alleTische: AktiverTischMitFavorit[] = []
// Je Query steuerbar: die drei Lesepfade der Seite scheitern unabhängig.
const fehler = { meineTische: false, alleTische: false, uebersicht: false }
const { reloadMeineTische, reloadAlleTische, reloadUebersicht } = vi.hoisted(
  () => ({
    reloadMeineTische: vi.fn(),
    reloadAlleTische: vi.fn(),
    reloadUebersicht: vi.fn(),
  }),
)

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
// der Seite, nicht das Rendern der Karten oder des Drawers. Die Übersichtskarten
// bleiben echt, damit der Fehlerfall ihre Null-Beträge nachweislich unterdrückt.
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

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  meineTische = []
  alleTische = []
  fehler.meineTische = false
  fehler.alleTische = false
  fehler.uebersicht = false
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

describe('TableSelectionPage bei Ladefehler', () => {
  it('zeigt einen Fehlerzustand statt der Leer-Defaults', () => {
    fehler.meineTische = true
    fehler.alleTische = true
    fehler.uebersicht = true
    render(<TableSelectionPage />)

    expect(
      screen.getByText('Tischübersicht konnte nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Übersicht 0,00 €) darf bei einem Fehler nicht
    // erscheinen — der Dienst wirkt sonst fälschlich abgerechnet.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
    // Der Alle-Tische-Drawer bleibt erreichbar.
    expect(
      screen.getByRole('button', { name: 'Alle Tische' }),
    ).toBeInTheDocument()
  })

  it('lädt über „Erneut versuchen" neu', async () => {
    fehler.meineTische = true
    fehler.uebersicht = true
    const user = userEvent.setup()
    render(<TableSelectionPage />)

    await user.click(screen.getByRole('button', { name: 'Erneut versuchen' }))

    expect(reloadMeineTische).toHaveBeenCalled()
    expect(reloadUebersicht).toHaveBeenCalled()
  })

  // Die Suche liest eine eigene Query. Scheitert nur sie, bleiben „Meine
  // Tische" und die Übersicht stehen — sonst wäre bei einem Suchlisten-Fehler
  // kein Tisch mehr erreichbar.
  it('lässt Meine Tische stehen, wenn nur die Suchliste scheitert', () => {
    fehler.alleTische = true
    meineTische = [tischSession(1, 'Stammtisch', true)]
    render(<TableSelectionPage />)

    expect(screen.getByText('Noch offen · 1')).toBeInTheDocument()
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.getByText('Tischsuche konnte nicht geladen werden'),
    ).toBeInTheDocument()
    // Kein stilles Suchfeld ohne Trefferliste.
    expect(
      screen.queryByPlaceholderText(/Tisch suchen/),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByText('Tischübersicht konnte nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('lädt nur die Suchliste nach, wenn nur sie scheitert', async () => {
    fehler.alleTische = true
    meineTische = [tischSession(1, 'Stammtisch', true)]
    const user = userEvent.setup()
    render(<TableSelectionPage />)

    await user.click(screen.getByRole('button', { name: 'Erneut versuchen' }))

    expect(reloadAlleTische).toHaveBeenCalled()
    expect(reloadMeineTische).not.toHaveBeenCalled()
  })
})

describe('TableSelectionPage im Vorgangs-Register', () => {
  it('meldet die Tischsuche als reine Anzeige nicht', async () => {
    meineTische = [tischSession(1, 'Stammtisch', true)]
    alleTische = [
      { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
      { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
    ]
    const user = userEvent.setup()
    render(<TableSelectionPage />)

    await user.type(screen.getByPlaceholderText(/Tisch suchen/), 'Bar')

    // Ein Suchbegriff filtert nur die Anzeige — es geht nichts verloren.
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
