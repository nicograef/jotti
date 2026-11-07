import type { QueryClient } from '@tanstack/react-query'
import { cleanup, screen, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AktiveKassensitzung } from '@/admin/kasse/KasseBackend'
import { TSE_TEST_WARNUNG } from '@/admin/tse/tseAmpel'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { AdminDashboardPage } from './AdminDashboardPage'
import type { LiveReportingData } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

function makeLiveData(): LiveReportingData {
  return {
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
      anzahlPfandrueckgaben: 0,
      pfandrueckgabenCents: 0,
    },
    breakdowns: { servicekraefte: [] },
    stornierungen: [],
    produktStatistik: [],
  }
}

function kassensitzung(
  status: AktiveKassensitzung['status'] = 'offen',
): AktiveKassensitzung {
  return {
    zNr: 1,
    datum: '2026-06-18',
    bezeichnung: 'Sommerfest',
    status,
    eroeffnetAm: '2026-06-18T08:02:00Z',
  }
}

// Healthy defaults: open session, TSE configured with 3 queued jobs, printer idle.
function backend({
  liveData = makeLiveData(),
  sitzung = liveData === null ? null : kassensitzung(),
  tseKonfiguriert = true,
  tseUmgebung = 'LIVE',
  fehlgeschlageneDrucke = 0,
}: {
  liveData?: LiveReportingData | null
  sitzung?: AktiveKassensitzung | null
  tseKonfiguriert?: boolean
  tseUmgebung?: 'TEST' | 'LIVE'
  fehlgeschlageneDrucke?: number
} = {}): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-live-reporting', liveData)
    .respond('admin/get-aktive-kassensitzung', sitzung)
    .respond('admin/get-kassenbestand', {
      sollBestandCents: 123450,
      anfangsbestandCents: 0,
      bareinnahmenCents: 123450,
      einlagenCents: 0,
      entnahmenCents: 0,
    })
    .respond('admin/get-tse-status', {
      umgebung: tseUmgebung,
      istKonfiguriert: tseKonfiguriert,
    })
    .respond('admin/get-tse-signatur-queue', {
      offeneAuftraege: 3,
      fehlgeschlageneAuftraege: 0,
      letzterFehler: '',
      rueckstandSekunden: 0,
      rueckstandKritisch: false,
      signaturenProMinute: 0,
      signierdauerP95Sekunden: 0,
    })
    .respond('admin/get-fehlgeschlagene-druckauftraege', {
      druckauftraege: Array.from({ length: fehlgeschlageneDrucke }, (_, i) => ({
        id: i + 1,
        bonArt: 'arbeitsbon',
        zielIp: '192.168.1.50',
        referenz: '',
        referenzAnzeige: '',
        versuche: 3,
        fehlerMeldung: 'Papier leer',
        erstelltAm: '2026-06-18T12:00:00Z',
      })),
    })
}

// Absence checks only hold once every status query has answered.
async function alleGeladen(queryClient: QueryClient) {
  await waitFor(() => {
    expect(queryClient.isFetching()).toBe(0)
  })
}

afterEach(() => {
  cleanup()
})

describe('AdminDashboardPage Status-Zeile', () => {
  it('zeigt ohne offene Kassensitzung den Leerzustand statt der Status-Zeile', async () => {
    const { queryClient } = renderWithBackend(
      <AdminDashboardPage />,
      backend({ liveData: null }),
    )

    expect(
      await screen.findByText('Keine Kassensitzung geöffnet'),
    ).toBeInTheDocument()
    await alleGeladen(queryClient)
    expect(screen.queryByText(/Soll-Bestand/)).not.toBeInTheDocument()
  })

  it('zeigt im Normalzustand Kasse/TSE/Drucker ohne Beheben-Button', async () => {
    const { queryClient } = renderWithBackend(<AdminDashboardPage />, backend())

    expect(
      await screen.findByText(/seit \d{2}:\d{2} · Soll-Bestand 1234,50 €/),
    ).toBeInTheDocument()
    await alleGeladen(queryClient)
    expect(
      screen.getByText('3 Vorgänge in Warteschlange (normal)'),
    ).toBeInTheDocument()
    expect(screen.getByText('Drucker bereit')).toBeInTheDocument()
    expect(
      screen.queryByRole('link', { name: 'Beheben' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt im Barrierestatus die Kassenzelle als unterbrochenen Abschluss mit Beheben-Link', async () => {
    const { queryClient } = renderWithBackend(
      <AdminDashboardPage />,
      backend({ sitzung: kassensitzung('wird_abgeschlossen') }),
    )

    expect(
      await screen.findByText('Abschluss unterbrochen'),
    ).toBeInTheDocument()
    await alleGeladen(queryClient)
    expect(screen.queryByText('Kasse offen')).not.toBeInTheDocument()
    const beheben = screen.getByRole('link', { name: 'Beheben' })
    expect(beheben).toHaveAttribute('href', '/admin/kasse')
  })

  it.each([
    [
      'nicht konfigurierter TSE',
      { tseKonfiguriert: false },
      'Nicht konfiguriert',
    ],
    [
      'TSE in der Test-Umgebung',
      { tseUmgebung: 'TEST' as const },
      TSE_TEST_WARNUNG,
    ],
  ])(
    'zeigt bei %s die Fehlerzelle mit Beheben-Link zum Finanzamt',
    async (_fall, optionen, text) => {
      const { queryClient } = renderWithBackend(
        <AdminDashboardPage />,
        backend(optionen),
      )

      expect(
        await screen.findByText('TSE benötigt Aufmerksamkeit'),
      ).toBeInTheDocument()
      await alleGeladen(queryClient)
      expect(screen.getByText(text)).toBeInTheDocument()
      const beheben = screen.getByRole('link', { name: 'Beheben' })
      expect(beheben).toHaveAttribute('href', '/admin/finanzamt')
    },
  )

  it('zeigt bei fehlgeschlagenen Druckaufträgen die Drucker-Fehlerzelle mit Beheben-Link', async () => {
    const { queryClient } = renderWithBackend(
      <AdminDashboardPage />,
      backend({ fehlgeschlageneDrucke: 1 }),
    )

    expect(await screen.findByText('1 Bon nicht gedruckt')).toBeInTheDocument()
    await alleGeladen(queryClient)
    const beheben = screen.getByRole('link', { name: 'Beheben' })
    expect(beheben).toHaveAttribute('href', '/admin/druckstationen')
  })
})
