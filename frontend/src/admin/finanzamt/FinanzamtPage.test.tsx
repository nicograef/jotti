import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

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

const liveTse: TSEStatus = { umgebung: 'LIVE', istKonfiguriert: true }

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

function backend({
  betreiber = makeBetreiber(),
  tseStatus = liveTse,
  queue = normaleQueue(),
  stoerungen = [],
}: {
  betreiber?: Betreiber
  tseStatus?: TSEStatus
  queue?: TSESignaturQueue
  stoerungen?: TSEStoerung[]
} = {}): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-betreiber', betreiber)
    .respond('admin/get-kassenidentitaet', kassenidentitaet)
    .respond('admin/get-tse-status', tseStatus)
    .respond('admin/get-tse-signatur-queue', queue)
    .respond('admin/get-tse-stoerungen', { stoerungen })
}

// While the TSE status loads, the traffic light shows green and the checklist
// shows the TSE step as open; assertions wait until every query has settled.
async function renderPage(fake: FakeBackend = backend()) {
  const { queryClient } = renderWithBackend(<FinanzamtPage />, fake)
  await waitFor(() => {
    expect(queryClient.isFetching()).toBe(0)
  })
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('FinanzamtPage — Einrichtungs-Checkliste', () => {
  it('zeigt „0 von 3" ohne Vereinsdaten, TSE und Meldung', async () => {
    await renderPage(
      backend({
        betreiber: makeBetreiber({
          vereinsname: '',
          strasse: '',
          plz: '',
          ort: '',
        }),
        tseStatus: { umgebung: '', istKonfiguriert: false },
      }),
    )

    expect(
      screen.getByText('Einrichtung — 0 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'TSE einrichten' }),
    ).toHaveAttribute('href', '/admin/tse-einrichtung')
  })

  it('bietet den Wizard-Link auch bei aktiver TSE an (Wechsel TEST → LIVE)', async () => {
    await renderPage(
      backend({ tseStatus: { umgebung: 'TEST', istKonfiguriert: true } }),
    )

    expect(screen.getByText(/Cloud-TSE verbunden/)).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'TSE einrichten' }),
    ).toHaveAttribute('href', '/admin/tse-einrichtung')
  })

  it('zeigt einen Ladefehler statt der leeren Checkliste, wenn die Betreiber-Query fehlschlägt', async () => {
    const getBetreiber = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(400, 'netzfehler')
      })
      .mockReturnValue(makeBetreiber())
    await renderPage(backend().respond('admin/get-betreiber', getBetreiber))

    expect(
      screen.getByText('Vereinsdaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(/von 3 Schritten erledigt/),
    ).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: 'Erneut versuchen' }),
    )
    expect(
      await screen.findByText(/von 3 Schritten erledigt/),
    ).toBeInTheDocument()
    expect(getBetreiber).toHaveBeenCalledTimes(2)
  })

  it('zeigt „2 von 3" mit Vereinsdaten und TSE, aber offener Meldung', async () => {
    await renderPage()

    expect(
      screen.getByText('Einrichtung — 2 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    expect(screen.getByText(/§ 146a Abs\. 4 AO/)).toBeInTheDocument()
    expect(
      screen.getByText('Seriennummer des elektronischen Aufzeichnungssystems'),
    ).toBeInTheDocument()
    expect(screen.getByText(kassenidentitaet.seriennummer)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Als erledigt markieren' }),
    ).toBeInTheDocument()
  })

  it('zeigt „3 von 3" und „Gemeldet am {Datum}" nach erfolgter Meldung', async () => {
    await renderPage(
      backend({ betreiber: makeBetreiber({ elsterGemeldetAm: '2026-07-12' }) }),
    )

    expect(
      screen.getByText('Einrichtung — 3 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    expect(screen.getByText(/Gemeldet am/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Zurücknehmen' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Noch offen — Frist/)).not.toBeInTheDocument()
  })

  it('ruft setElsterMeldung beim Abhaken der Kassenmeldung', async () => {
    let elsterGemeldetAm: string | null = null
    const fake = backend()
      .respond('admin/get-betreiber', () => makeBetreiber({ elsterGemeldetAm }))
      .respond('admin/elster-meldung-setzen', () => {
        elsterGemeldetAm = '2026-07-12'
        return {}
      })
    await renderPage(fake)

    await userEvent.click(
      screen.getByRole('button', { name: 'Als erledigt markieren' }),
    )

    expect(
      await screen.findByText('Einrichtung — 3 von 3 Schritten erledigt'),
    ).toBeInTheDocument()
    expect(fake.bodies('admin/elster-meldung-setzen')).toHaveLength(1)
  })
})

describe('FinanzamtPage — Läuft-alles-Ampel', () => {
  it('zeigt den grünen Normalzustand als Klartext', async () => {
    await renderPage()

    expect(screen.getByText('Ja — TSE signiert normal')).toBeInTheDocument()
  })

  it('zeigt den roten Fehlerzustand bei fehlgeschlagenen Signaturen', async () => {
    await renderPage(
      backend({ queue: { ...normaleQueue(), fehlgeschlageneAuftraege: 2 } }),
    )

    expect(screen.getByText('TSE braucht Aufmerksamkeit')).toBeInTheDocument()
  })

  it('zeigt den roten Fehlerzustand bei Rückstand über der 60-s-Schwelle', async () => {
    await renderPage(
      backend({ queue: { ...normaleQueue(), rueckstandSekunden: 90 } }),
    )

    expect(screen.getByText('TSE braucht Aufmerksamkeit')).toBeInTheDocument()
  })
})

describe('FinanzamtPage — Signatur-Warteschlange', () => {
  it('meldet fehlgeschlagene Signaturen zuerst, auch ohne offene Aufträge', async () => {
    await renderPage(
      backend({
        queue: {
          ...normaleQueue(),
          offeneAuftraege: 0,
          rueckstandSekunden: 0,
          fehlgeschlageneAuftraege: 2,
          letzterFehler: 'TSE nicht erreichbar',
        },
      }),
    )

    expect(
      screen.getByText(
        '2 Vorgänge sind fehlgeschlagen. Keine Vorgänge in der Warteschlange.',
      ),
    ).toBeInTheDocument()
  })

  it('beruhigt nicht, wenn neben einem kleinen Rückstand ein Vorgang fehlgeschlagen ist', async () => {
    await renderPage(
      backend({ queue: { ...normaleQueue(), fehlgeschlageneAuftraege: 2 } }),
    )

    expect(
      screen.getByText(
        '2 Vorgänge sind fehlgeschlagen. 3 Vorgänge warten (ältester 12 s).',
      ),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(/normal bei vollem Betrieb/),
    ).not.toBeInTheDocument()
  })

  it('beruhigt nicht mehr, wenn der Rückstand die Warnschwelle erreicht', async () => {
    await renderPage(
      backend({ queue: { ...normaleQueue(), rueckstandSekunden: 90 } }),
    )

    expect(screen.getByText(/der Rückstand ist zu groß/)).toBeInTheDocument()
    expect(
      screen.queryByText(/normal bei vollem Betrieb/),
    ).not.toBeInTheDocument()
  })

  it('führt Fehler-Zähler und letzten Fehlertext in den Roh-Metriken', async () => {
    await renderPage(
      backend({
        queue: {
          ...normaleQueue(),
          fehlgeschlageneAuftraege: 2,
          letzterFehler: 'TSE nicht erreichbar',
        },
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: /Technische Details/ }),
    )

    expect(screen.getByText('Fehlgeschlagen')).toBeInTheDocument()
    expect(screen.getByText('Letzter Fehler')).toBeInTheDocument()
    expect(screen.getByText('TSE nicht erreichbar')).toBeInTheDocument()
  })
})

describe('FinanzamtPage — Collapsibles', () => {
  it('blendet die Roh-Metriken erst nach Klick auf „Technische Details" ein', async () => {
    await renderPage()

    expect(screen.queryByText('Signaturen/Minute')).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: /Technische Details/ }),
    )

    expect(screen.getByText('Signaturen/Minute')).toBeInTheDocument()
  })

  it('bietet das Störungsprotokoll aufklappbar an, wenn Störungen vorliegen', async () => {
    await renderPage(
      backend({
        stoerungen: [
          {
            id: 1,
            beginn: '2026-07-05T14:00:00Z',
            ende: '2026-07-05T14:04:00Z',
            grundArt: 'rueckstand',
            fehlertext: 'Nachsigniert nach kurzem Rückstand',
          },
        ],
      }),
    )

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
