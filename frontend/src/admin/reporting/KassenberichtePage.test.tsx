import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { OffeneKassensitzung } from '@/admin/kasse/KasseBackend'

import { KassenberichtePage } from './KassenberichtePage'
import type { AbgeschlosseneSitzung, ReportingData } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

const hookState = vi.hoisted(() => ({
  kassensitzungen: [] as AbgeschlosseneSitzung[],
  listLoading: false,
  offeneSitzung: null as OffeneKassensitzung | null,
  report: null as ReportingData | null,
  reportLoading: false,
}))

vi.mock('./hooks', () => ({
  useAbgeschlosseneKassensitzungen: () => ({
    kassensitzungen: hookState.kassensitzungen,
    isPending: hookState.listLoading,
  }),
  useReport: () => ({
    result: hookState.report,
    isPending: hookState.reportLoading,
  }),
  useDsfinvkExport: () => ({ exportieren: vi.fn(), isPending: false }),
}))

vi.mock('@/admin/kasse/hooks', () => ({
  useOffeneKassensitzung: () => ({
    kassensitzung: hookState.offeneSitzung,
    isPending: false,
    isError: false,
    refetch: vi.fn(),
  }),
}))

function makeReport(zNr: number): ReportingData {
  return {
    kassensitzungNr: zNr,
    metadaten: {
      eroeffnetAm: '2026-07-05T07:58:00Z',
      abgeschlossenAm: '2026-07-05T21:12:00Z',
      abgeschlossenVon: 'nico',
      kassensturzDifferenzCents: -150,
    },
    summary: {
      gesamtUmsatzCents: 341200,
      gesamtBestellungenCents: 0,
      gesamtStornierungenCents: 0,
      geldtransitCents: 0,
      anzahlBestellungen: 214,
      anzahlStornierungen: 0,
      anzahlDirektverkaeufe: 0,
      direktverkaufUmsatzCents: 0,
    },
    breakdowns: {
      umsatzProServicekraft: [],
      stornierungenProServicekraft: [],
    },
    umsatzProSteuersatz: [],
    stornierungen: [],
  }
}

afterEach(() => {
  cleanup()
  hookState.kassensitzungen = []
  hookState.listLoading = false
  hookState.offeneSitzung = null
  hookState.report = null
  hookState.reportLoading = false
})

describe('KassenberichtePage', () => {
  it('zeigt ohne abgeschlossene Kassensitzung einen erklärenden leeren Zustand mit Link zur Kasse', () => {
    hookState.kassensitzungen = []
    render(<KassenberichtePage />)

    expect(
      screen.getByText('Noch keine abgeschlossene Kassensitzung'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Zur Kassensitzungs-Seite' }),
    ).toHaveAttribute('href', '/admin/kasse')
  })

  it('zeigt die Sitzungsliste mit Datum, Nr. und Umsatz und den Berichtskopf', () => {
    hookState.kassensitzungen = [
      {
        zNr: 11,
        datum: '2026-07-05',
        bezeichnung: 'Sommerfest Tag 1',
        umsatzGesamtCents: 341200,
        abgeschlossenAm: '2026-07-05T21:12:00Z',
      },
    ]
    hookState.report = makeReport(11)
    render(<KassenberichtePage />)

    // Sitzungslisten-Karte: Datum + Nr. und Umsatz, keine Status-Emojis.
    expect(
      screen.getByText((_content, el) => {
        const text = el?.textContent ?? ''
        return (
          el?.tagName === 'SPAN' &&
          text.includes('Nr. 11') &&
          text.includes('05.07.')
        )
      }),
    ).toBeInTheDocument()
    // Umsatz erscheint sowohl in der Karte als auch in der Kennzahl-Kachel.
    expect(screen.getAllByText('3412,00 €').length).toBeGreaterThanOrEqual(1)
    expect(screen.queryByText('🟢')).not.toBeInTheDocument()
    expect(screen.queryByText('🔴')).not.toBeInTheDocument()

    // Berichtskopf rechts.
    expect(
      screen.getByRole('heading', {
        name: 'Tagesbericht Nr. 11 — Sommerfest Tag 1',
      }),
    ).toBeInTheDocument()

    // Export-Block.
    expect(
      screen.getByRole('button', { name: 'Archiv herunterladen (ZIP)' }),
    ).toBeInTheDocument()
  })

  it('zeigt die offene Sitzung als nicht wählbaren Eintrag mit Verweis zur Übersicht', () => {
    hookState.kassensitzungen = [
      {
        zNr: 11,
        datum: '2026-07-05',
        bezeichnung: 'Sommerfest Tag 1',
        umsatzGesamtCents: 341200,
        abgeschlossenAm: '2026-07-05T21:12:00Z',
      },
    ]
    hookState.report = makeReport(11)
    hookState.offeneSitzung = {
      zNr: 12,
      datum: '2026-07-06',
      bezeichnung: 'Sommerfest Tag 2',
      status: 'offen',
      eroeffnetAm: '2026-07-06T08:00:00Z',
    }
    render(<KassenberichtePage />)

    expect(screen.getByText(/läuft — siehe Übersicht/)).toBeInTheDocument()
    // Die offene Sitzung ist kein Button (nicht wählbar), sondern ein Link zur Übersicht.
    const links = screen.getAllByRole('link')
    expect(
      links.some((l) => l.getAttribute('href') === '/admin/auswertung'),
    ).toBe(true)
  })
})
