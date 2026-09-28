import { cleanup, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AktiveKassensitzung } from '@/admin/kasse/KasseBackend'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { KassenberichtePage } from './KassenberichtePage'
import type { AbgeschlosseneSitzung, ReportingData } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

const sommerfestTag1: AbgeschlosseneSitzung = {
  zNr: 11,
  datum: '2026-07-05',
  bezeichnung: 'Sommerfest Tag 1',
  umsatzGesamtCents: 341200,
  abgeschlossenAm: '2026-07-05T21:12:00Z',
}

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
      abrechnungProServicekraft: [],
    },
    umsatzProSteuersatz: [],
    stornierungen: [],
    produktStatistik: [],
  }
}

function backend({
  kassensitzungen = [],
  aktiveSitzung = null,
}: {
  kassensitzungen?: AbgeschlosseneSitzung[]
  aktiveSitzung?: AktiveKassensitzung | null
} = {}): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-abgeschlossene-kassensitzungen', { kassensitzungen })
    .respond('admin/get-aktive-kassensitzung', aktiveSitzung)
    .respond('admin/get-abrechnung', (body: unknown) =>
      makeReport((body as { kassensitzungNr: number }).kassensitzungNr),
    )
}

afterEach(() => {
  cleanup()
})

describe('KassenberichtePage', () => {
  it('zeigt ohne abgeschlossene Kassensitzung einen erklärenden leeren Zustand mit Link zur Kasse', async () => {
    renderWithBackend(<KassenberichtePage />, backend())

    expect(
      await screen.findByText('Noch keine abgeschlossene Kassensitzung'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Zur Kassensitzungs-Seite' }),
    ).toHaveAttribute('href', '/admin/kasse')
  })

  it('zeigt die Sitzungsliste mit Datum, Nr. und Umsatz und den Berichtskopf', async () => {
    renderWithBackend(
      <KassenberichtePage />,
      backend({ kassensitzungen: [sommerfestTag1] }),
    )

    expect(
      await screen.findByRole('heading', {
        name: 'Tagesbericht Nr. 11 — Sommerfest Tag 1',
      }),
    ).toBeInTheDocument()
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

    expect(
      screen.getByRole('button', { name: 'Archiv herunterladen (ZIP)' }),
    ).toBeInTheDocument()
  })

  it('zeigt die offene Sitzung als nicht wählbaren Eintrag mit Verweis zur Übersicht', async () => {
    renderWithBackend(
      <KassenberichtePage />,
      backend({
        kassensitzungen: [sommerfestTag1],
        aktiveSitzung: {
          zNr: 12,
          datum: '2026-07-06',
          bezeichnung: 'Sommerfest Tag 2',
          status: 'offen',
          eroeffnetAm: '2026-07-06T08:00:00Z',
        },
      }),
    )

    expect(
      await screen.findByText(/läuft — siehe Übersicht/),
    ).toBeInTheDocument()
    // Die aktive Sitzung ist kein Button (nicht wählbar), sondern ein Link zur Übersicht.
    const links = screen.getAllByRole('link')
    expect(
      links.some((l) => l.getAttribute('href') === '/admin/auswertung'),
    ).toBe(true)
  })

  it('weist die aktive Sitzung im Barrierestatus als unterbrochenen Abschluss aus', async () => {
    renderWithBackend(
      <KassenberichtePage />,
      backend({
        kassensitzungen: [sommerfestTag1],
        aktiveSitzung: {
          zNr: 12,
          datum: '2026-07-06',
          bezeichnung: 'Sommerfest Tag 2',
          status: 'wird_abgeschlossen',
          eroeffnetAm: '2026-07-06T08:00:00Z',
        },
      }),
    )

    expect(
      await screen.findByText('Abschluss unterbrochen'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/läuft/)).not.toBeInTheDocument()
  })
})
