import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  TSESignaturQueue,
  TSEStatus,
  TSEStoerung,
} from '../tse/TSEBackend'
import type { Betreiber, Kassenidentitaet } from './BetreiberBackend'
import { FinanzamtPage } from './FinanzamtPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

const setElsterMeldung = vi.fn(() => Promise.resolve())
const nimmElsterMeldungZurueck = vi.fn(() => Promise.resolve())
const saveBetreiber = vi.fn(() => Promise.resolve())

const refetchBetreiber = vi.fn(() => Promise.resolve())

const hookState = vi.hoisted(() => ({
  betreiber: null as Betreiber | null,
  betreiberLoading: false,
  betreiberError: false,
  kassenidentitaet: null as Kassenidentitaet | null,
  tseStatus: undefined as TSEStatus | undefined,
  tseLoading: false,
  queue: undefined as TSESignaturQueue | undefined,
  stoerungen: [] as TSEStoerung[],
}))

vi.mock('./hooks', () => ({
  useBetreiber: () => ({
    betreiber: hookState.betreiber,
    isPending: hookState.betreiberLoading,
    isError: hookState.betreiberError,
    error: hookState.betreiberError ? new Error('Netzfehler') : null,
    refetchBetreiber,
    saveBetreiber,
    setElsterMeldung,
    nimmElsterMeldungZurueck,
  }),
  useKassenidentitaet: () => ({
    data: hookState.kassenidentitaet,
    isPending: false,
    error: null,
  }),
}))

vi.mock('../tse/hooks', () => ({
  RUECKSTAND_WARN_SEKUNDEN: 60,
  useTSEStatus: () => ({
    tseStatus: hookState.tseStatus,
    isPending: hookState.tseLoading,
    error: null,
  }),
  useTSESignaturQueue: () => ({ queue: hookState.queue }),
  useTSEStoerungen: () => ({ stoerungen: hookState.stoerungen }),
}))

function makeBetreiber(overrides: Partial<Betreiber> = {}): Betreiber {
  return {
    vereinsname: 'Musterverein e.V.',
    strasse: 'Musterstraße 1',
    plz: '12345',
    ort: 'Musterstadt',
    steuernummer: null,
    ustId: null,
    elsterGemeldetAm: null,
    ...overrides,
  }
}

const kassenidentitaet: Kassenidentitaet = {
  seriennummer: 'a3f8c2e1-7b94-4d06-9e2a-51c8f0b7d3a9',
  angelegtAm: '2026-07-01',
}

function normaleQueue(): TSESignaturQueue {
  return {
    offeneAuftraege: 3,
    fehlgeschlageneAuftraege: 0,
    letzterFehler: '',
    rueckstandSekunden: 12,
    signaturenProMinute: 20,
    signierdauerP95Sekunden: 1.2,
  }
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  hookState.betreiber = null
  hookState.betreiberLoading = false
  hookState.betreiberError = false
  hookState.kassenidentitaet = null
  hookState.tseStatus = undefined
  hookState.tseLoading = false
  hookState.queue = undefined
  hookState.stoerungen = []
})

describe('FinanzamtPage — Einrichtungs-Checkliste', () => {
  it('zeigt „0 von 3" ohne Vereinsdaten, TSE und Meldung', () => {
    hookState.betreiber = makeBetreiber({
      vereinsname: '',
      strasse: '',
      plz: '',
      ort: '',
    })
    hookState.tseStatus = { umgebung: '', istKonfiguriert: false }
    hookState.kassenidentitaet = kassenidentitaet
    render(<FinanzamtPage />)

    expect(
      screen.getByText('Einrichtung — 0 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    // Schritt 2 unerledigt bietet den Wizard-Link an.
    expect(
      screen.getByRole('link', { name: 'TSE einrichten' }),
    ).toHaveAttribute('href', '/admin/tse-einrichtung')
  })

  it('zeigt einen Ladefehler statt der leeren Checkliste, wenn die Betreiber-Query fehlschlägt', async () => {
    hookState.betreiberError = true
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    render(<FinanzamtPage />)

    expect(
      screen.getByText('Vereinsdaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Kein irreführender „0 von 3"-Leerstand bei einem Ladefehler.
    expect(
      screen.queryByText(/von 3 Schritten erledigt/),
    ).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: 'Erneut versuchen' }),
    )
    expect(refetchBetreiber).toHaveBeenCalledOnce()
  })

  it('zeigt „2 von 3" mit Vereinsdaten und TSE, aber offener Meldung', () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.kassenidentitaet = kassenidentitaet
    render(<FinanzamtPage />)

    expect(
      screen.getByText('Einrichtung — 2 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    // Offene Meldung: Fristtext mit Paragraf und Seriennummer-Pill.
    expect(screen.getByText(/§ 146a Abs\. 4 AO/)).toBeInTheDocument()
    expect(screen.getByText(kassenidentitaet.seriennummer)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Als erledigt markieren' }),
    ).toBeInTheDocument()
  })

  it('zeigt „3 von 3" und „Gemeldet am {Datum}" nach erfolgter Meldung', () => {
    hookState.betreiber = makeBetreiber({ elsterGemeldetAm: '2026-07-12' })
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.kassenidentitaet = kassenidentitaet
    render(<FinanzamtPage />)

    expect(
      screen.getByText('Einrichtung — 3 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    expect(screen.getByText(/Gemeldet am/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Zurücknehmen' }),
    ).toBeInTheDocument()
    // Kein Fristtext mehr, wenn erledigt.
    expect(screen.queryByText(/Noch offen — Frist/)).not.toBeInTheDocument()
  })

  it('ruft setElsterMeldung beim Abhaken der Kassenmeldung', async () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.kassenidentitaet = kassenidentitaet
    render(<FinanzamtPage />)

    await userEvent.click(
      screen.getByRole('button', { name: 'Als erledigt markieren' }),
    )

    await waitFor(() => {
      expect(setElsterMeldung).toHaveBeenCalledOnce()
    })
  })
})

describe('FinanzamtPage — Läuft-alles-Ampel', () => {
  it('zeigt den grünen Normalzustand als Klartext', () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.queue = normaleQueue()
    render(<FinanzamtPage />)

    expect(screen.getByText('Ja — TSE signiert normal')).toBeInTheDocument()
  })

  it('zeigt den roten Fehlerzustand bei fehlgeschlagenen Signaturen', () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.queue = { ...normaleQueue(), fehlgeschlageneAuftraege: 2 }
    render(<FinanzamtPage />)

    expect(screen.getByText('TSE braucht Aufmerksamkeit')).toBeInTheDocument()
  })

  it('zeigt den roten Fehlerzustand bei Rückstand über der 60-s-Schwelle', () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.queue = { ...normaleQueue(), rueckstandSekunden: 90 }
    render(<FinanzamtPage />)

    expect(screen.getByText('TSE braucht Aufmerksamkeit')).toBeInTheDocument()
  })
})

describe('FinanzamtPage — Collapsibles', () => {
  it('blendet die Roh-Metriken erst nach Klick auf „Technische Details" ein', async () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.queue = normaleQueue()
    render(<FinanzamtPage />)

    // Vor dem Aufklappen ist die Roh-Metrik „Signaturen/Minute" nicht sichtbar.
    expect(screen.queryByText('Signaturen/Minute')).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: /Technische Details/ }),
    )

    expect(screen.getByText('Signaturen/Minute')).toBeInTheDocument()
  })

  it('bietet das Störungsprotokoll aufklappbar an, wenn Störungen vorliegen', async () => {
    hookState.betreiber = makeBetreiber()
    hookState.tseStatus = { umgebung: 'LIVE', istKonfiguriert: true }
    hookState.queue = normaleQueue()
    hookState.stoerungen = [
      {
        id: 1,
        beginn: '2026-07-05T14:00:00Z',
        ende: '2026-07-05T14:04:00Z',
        grundArt: 'rueckstand',
        fehlertext: 'Nachsigniert nach kurzem Rückstand',
      },
    ]
    render(<FinanzamtPage />)

    // Vor dem Aufklappen ist die Detailzeile (Fehlertext) nicht sichtbar.
    expect(screen.getByText(/1 dokumentierte Störung/)).toBeInTheDocument()
    expect(
      screen.queryByText('Nachsigniert nach kurzem Rückstand'),
    ).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: /Protokoll ansehen/ }),
    )
    expect(
      screen.getByText('Nachsigniert nach kurzem Rückstand'),
    ).toBeInTheDocument()
  })
})
