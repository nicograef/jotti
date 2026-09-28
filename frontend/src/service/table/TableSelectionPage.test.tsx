import { cleanup, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type { Position } from './Bestellung'
import { TableSelectionPage } from './TableSelectionPage'
import type { AktiverTischMitFavorit, TischSession } from './Tisch'

const navigate = vi.fn()
vi.mock('react-router', () => ({
  useNavigate: () => navigate,
}))

const offenePosition: Position = {
  positionId: '00000000-0000-4000-8000-000000000001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 500,
  menge: 1,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

function tischSession(
  tischId: number,
  tischName: string,
  offen: boolean,
): TischSession {
  return {
    tischId,
    tischName,
    saldoCents: offen ? 500 : 0,
    unbezahltePositionen: offen ? [offenePosition] : [],
    fuerMichErledigt: !offen,
  }
}

const leereUebersicht = {
  anzahlBestellungen: 0,
  bestellungenCents: 0,
  anzahlZahlungen: 0,
  zahlungenCents: 0,
  anzahlRuecknahmen: 0,
  ruecknahmenCents: 0,
  abzugebenCents: 0,
}

// A function answer lets a test fail the first load and succeed on the retry.
function backend({
  meineTische = [],
  alleTische = [],
}: {
  meineTische?: TischSession[] | (() => unknown)
  alleTische?: AktiverTischMitFavorit[] | (() => unknown)
} = {}): FakeBackend {
  return new FakeBackend()
    .respond(
      'service/get-meine-tische-state',
      typeof meineTische === 'function' ? meineTische : { tische: meineTische },
    )
    .respond(
      'service/get-aktive-tische-mit-favoriten',
      typeof alleTische === 'function' ? alleTische : { tische: alleTische },
    )
    .respond('service/get-eigene-uebersicht', leereUebersicht)
}

function renderPage(fake: FakeBackend) {
  return renderWithBackend(<TableSelectionPage />, fake)
}

function einmalFehler(danach: unknown) {
  return vi
    .fn()
    .mockImplementationOnce(() => {
      throw new BackendError(400, 'test_fehler')
    })
    .mockReturnValue(danach)
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
  signIn({ userId: 1 })
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  signOut()
})

describe('TableSelectionPage', () => {
  it('zeigt bei leerem Suchfeld die Favoriten („Meine Tische")', async () => {
    renderPage(
      backend({
        meineTische: [tischSession(1, 'Stammtisch', true)],
        alleTische: [
          { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
          { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
        ],
      }),
    )

    // The search field appears once all tables loaded, so „Bar" is known.
    await screen.findByPlaceholderText(/Tisch suchen/)
    expect(await screen.findByText('Noch offen · 1')).toBeInTheDocument()
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(screen.queryByText('Bar')).not.toBeInTheDocument()
  })

  it('findet einen nicht favorisierten aktiven Tisch über die Hauptsuche', async () => {
    const user = userEvent.setup()
    renderPage(
      backend({
        meineTische: [tischSession(1, 'Stammtisch', true)],
        alleTische: [
          { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
          { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
        ],
      }),
    )

    // Suchfeld und Treffer werden wie im e2e-Helper angesprochen
    // (e2e/support/servicekraft.ts, oeffneTisch): Platzhalter-Teilstring plus
    // Button-Name „<Name> … <Saldo> €".
    await user.type(await screen.findByPlaceholderText(/Tisch suchen/), 'Bar')

    const treffer = screen.getByRole('button', { name: /^Bar\b.*€/ })
    expect(treffer).toBeInTheDocument()
    await user.click(treffer)
    expect(navigate).toHaveBeenCalledWith('/service/tische/2')
  })

  it('meldet, wenn kein aktiver Tisch zur Suche passt', async () => {
    const user = userEvent.setup()
    renderPage(
      backend({
        meineTische: [tischSession(1, 'Stammtisch', true)],
        alleTische: [
          { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
        ],
      }),
    )

    await user.type(
      await screen.findByPlaceholderText('Tisch suchen — Name oder Nummer'),
      'Zelt',
    )

    expect(screen.getByText(/Kein aktiver Tisch passt zu/)).toBeInTheDocument()
  })
})

describe('TableSelectionPage bei Ladefehler', () => {
  it('zeigt einen Fehlerzustand statt der Leer-Defaults', async () => {
    renderPage(
      backend()
        .fail('service/get-meine-tische-state')
        .fail('service/get-aktive-tische-mit-favoriten')
        .fail('service/get-eigene-uebersicht'),
    )

    expect(
      await screen.findByText('Tischübersicht konnte nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Übersicht 0,00 €) darf bei einem Fehler nicht
    // erscheinen — der Dienst wirkt sonst fälschlich abgerechnet.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Alle Tische' }),
    ).toBeInTheDocument()
  })

  it('lädt über „Erneut versuchen" neu', async () => {
    const user = userEvent.setup()
    renderPage(
      backend({
        meineTische: einmalFehler({
          tische: [tischSession(1, 'Stammtisch', true)],
        }),
      }).respond(
        'service/get-eigene-uebersicht',
        einmalFehler(leereUebersicht),
      ),
    )

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    // The error only clears once both the tables and the overview reloaded.
    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tischübersicht konnte nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  // Scheitert nur die Suchliste, wäre sonst kein Tisch mehr erreichbar.
  it('lässt Meine Tische stehen, wenn nur die Suchliste scheitert', async () => {
    renderPage(
      backend({ meineTische: [tischSession(1, 'Stammtisch', true)] }).fail(
        'service/get-aktive-tische-mit-favoriten',
      ),
    )

    expect(
      await screen.findByText('Tischsuche konnte nicht geladen werden'),
    ).toBeInTheDocument()
    expect(await screen.findByText('Noch offen · 1')).toBeInTheDocument()
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByPlaceholderText(/Tisch suchen/),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByText('Tischübersicht konnte nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('lädt nur die Suchliste nach, wenn nur sie scheitert', async () => {
    const user = userEvent.setup()
    const fake = backend({
      meineTische: [tischSession(1, 'Stammtisch', true)],
      alleTische: einmalFehler({
        tische: [
          { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
        ],
      }),
    })
    renderPage(fake)
    await screen.findByText('Noch offen · 1')

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(
      await screen.findByPlaceholderText(/Tisch suchen/),
    ).toBeInTheDocument()
    expect(fake.bodies('service/get-meine-tische-state')).toHaveLength(1)
  })
})

describe('TableSelectionPage im Vorgangs-Register', () => {
  it('meldet die Tischsuche als reine Anzeige nicht', async () => {
    const user = userEvent.setup()
    renderPage(
      backend({
        meineTische: [tischSession(1, 'Stammtisch', true)],
        alleTische: [
          { id: 1, name: 'Stammtisch', istFavorit: true, saldoCents: 500 },
          { id: 2, name: 'Bar', istFavorit: false, saldoCents: 300 },
        ],
      }),
    )

    await user.type(await screen.findByPlaceholderText(/Tisch suchen/), 'Bar')

    // Ein Suchbegriff filtert nur die Anzeige — es geht nichts verloren.
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
