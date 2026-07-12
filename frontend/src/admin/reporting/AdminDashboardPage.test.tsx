import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { AdminDashboardPage } from './AdminDashboardPage'
import type { LiveReportingData } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

const liveState = vi.hoisted(() => ({
  data: null as LiveReportingData | null,
}))
const tseState = vi.hoisted(() => ({
  istKonfiguriert: true,
  offeneAuftraege: 3,
  fehlgeschlageneAuftraege: 0,
  rueckstandSekunden: 0,
}))
const druckState = vi.hoisted(() => ({ anzahl: 0 }))

vi.mock('./hooks', () => ({
  useLiveReporting: () => ({
    liveData: liveState.data,
    isPending: false,
    dataUpdatedAt: 0,
    refetch: vi.fn(),
  }),
}))

vi.mock('@/admin/kasse/hooks', () => ({
  useOffeneKassensitzung: () => ({
    kassensitzung:
      liveState.data === null
        ? null
        : { zNr: 1, eroeffnetAm: '2026-06-18T08:02:00Z' },
  }),
  useKassenbestand: () => ({ kassenbestand: { sollBestandCents: 123450 } }),
}))

vi.mock('@/admin/tse/hooks', () => ({
  RUECKSTAND_WARN_SEKUNDEN: 60,
  useTSEStatus: () => ({
    tseStatus: { istKonfiguriert: tseState.istKonfiguriert },
    isPending: false,
  }),
  useTSESignaturQueue: () => ({
    queue: {
      offeneAuftraege: tseState.offeneAuftraege,
      fehlgeschlageneAuftraege: tseState.fehlgeschlageneAuftraege,
      rueckstandSekunden: tseState.rueckstandSekunden,
      letzterFehler: '',
    },
  }),
}))

vi.mock('@/admin/settings/hooks', () => ({
  useFehlgeschlageneDruckauftraege: () => ({
    druckauftraege: Array.from({ length: druckState.anzahl }, (_, i) => ({
      id: i + 1,
    })),
  }),
}))

function makeLiveData(): LiveReportingData {
  return {
    kassensitzungNr: 1,
    bezeichnung: 'Sommerfest',
    datum: '2026-06-18',
    offeneTische: [],
    offeneSaldiCents: 0,
    summary: {
      gesamtUmsatzCents: 284750,
      gesamtBestellungenCents: 325950,
      gesamtStornierungenCents: 3650,
      geldtransitCents: 0,
      anzahlBestellungen: 42,
      anzahlStornierungen: 4,
      anzahlDirektverkaeufe: 63,
      direktverkaufUmsatzCents: 48650,
    },
    breakdowns: { servicekraefte: [], stornierungenProServicekraft: [] },
    stornierungen: [],
  }
}

afterEach(() => {
  cleanup()
  liveState.data = null
  tseState.istKonfiguriert = true
  tseState.offeneAuftraege = 3
  tseState.fehlgeschlageneAuftraege = 0
  tseState.rueckstandSekunden = 0
  druckState.anzahl = 0
})

describe('AdminDashboardPage Status-Zeile', () => {
  it('zeigt ohne offene Kassensitzung den Leerzustand statt der Status-Zeile', () => {
    liveState.data = null
    render(<AdminDashboardPage />)

    expect(screen.getByText('Keine Kassensitzung geöffnet')).toBeInTheDocument()
    expect(screen.queryByText(/Soll-Bestand/)).not.toBeInTheDocument()
  })

  it('zeigt im Normalzustand Kasse/TSE/Drucker ohne Beheben-Button', () => {
    liveState.data = makeLiveData()
    render(<AdminDashboardPage />)

    // Kasse: seit HH:MM plus Soll-Bestand (1234,50 €).
    expect(
      screen.getByText(/seit \d{2}:\d{2} · Soll-Bestand 1234,50 €/),
    ).toBeInTheDocument()
    // TSE normal: Warteschlangen-Text.
    expect(
      screen.getByText('3 Vorgänge in Warteschlange (normal)'),
    ).toBeInTheDocument()
    expect(screen.getByText('Drucker bereit')).toBeInTheDocument()
    // Keine Beheben-Buttons im Normalzustand.
    expect(
      screen.queryByRole('link', { name: 'Beheben' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei nicht konfigurierter TSE die Fehlerzelle mit Beheben-Link zum Finanzamt', () => {
    liveState.data = makeLiveData()
    tseState.istKonfiguriert = false
    render(<AdminDashboardPage />)

    expect(screen.getByText('TSE benötigt Aufmerksamkeit')).toBeInTheDocument()
    const beheben = screen.getByRole('link', { name: 'Beheben' })
    expect(beheben).toHaveAttribute('href', '/admin/finanzamt')
  })

  it('zeigt bei fehlgeschlagenen Druckaufträgen die Drucker-Fehlerzelle mit Beheben-Link', () => {
    liveState.data = makeLiveData()
    druckState.anzahl = 1
    render(<AdminDashboardPage />)

    expect(screen.getByText('1 Bon nicht gedruckt')).toBeInTheDocument()
    const beheben = screen.getByRole('link', { name: 'Beheben' })
    expect(beheben).toHaveAttribute('href', '/admin/druckstationen')
  })
})
