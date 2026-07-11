import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { LiveReportingSection } from './LiveReportingSection'
import type { LiveReportingData } from './types'

afterEach(() => {
  cleanup()
})

function liveData(
  servicekraefte: LiveReportingData['breakdowns']['servicekraefte'],
): LiveReportingData {
  return {
    kassensitzungNr: 1,
    bezeichnung: 'Sommerfest',
    datum: '2026-06-18',
    offeneTische: [],
    offeneSaldiCents: 0,
    summary: {
      gesamtUmsatzCents: 2400,
      gesamtBestellungenCents: 3600,
      gesamtStornierungenCents: 0,
      geldtransitCents: 0,
      anzahlBestellungen: 5,
      anzahlStornierungen: 0,
      anzahlDirektverkaeufe: 2,
      direktverkaufUmsatzCents: 800,
    },
    breakdowns: { servicekraefte },
    stornierungen: [],
  }
}

const noopRefresh = () => {
  // no-op
}

describe('LiveReportingSection — Single-Scroll', () => {
  it('rendert alle Blöcke ohne Tabs und zeigt offene Arbeit als Euro-Betrag mit Tischnamen', () => {
    render(
      <LiveReportingSection
        liveData={liveData([
          {
            userId: 7,
            userName: 'Anna',
            name: 'Anna A.',
            zahlungenCents: 1500,
            anzahlZahlungen: 2,
            offeneTische: [
              {
                tischId: 3,
                tischName: 'Tisch 3',
                anzahlUnbezahlt: 2,
                anzahlOffen: 2,
                offenCents: 500,
              },
              {
                tischId: 4,
                tischName: 'Zelt A2',
                anzahlUnbezahlt: 1,
                anzahlOffen: 1,
                offenCents: 250,
              },
            ],
            erledigt: false,
          },
          {
            userId: 9,
            userName: 'Cleo',
            name: '',
            zahlungenCents: 900,
            anzahlZahlungen: 1,
            offeneTische: [],
            erledigt: true,
          },
        ])}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Single-Scroll: keine Tabs mehr.
    expect(screen.queryByRole('tab')).not.toBeInTheDocument()

    // Servicekraft mit offener Arbeit: Euro-Betrag (5,00 + 2,50 = 7,50 €) und
    // die Tischnamen inline, kein "N zu kassieren" mehr.
    expect(screen.getByText('Anna (Anna A.)')).toBeInTheDocument()
    expect(screen.getByText('7,50 €')).toBeInTheDocument()
    expect(screen.getByText(/Tisch 3, Zelt A2/)).toBeInTheDocument()
    expect(screen.queryByText(/zu kassieren/)).not.toBeInTheDocument()

    // Fertige Servicekraft: nur Hinweis.
    expect(screen.getByText('Cleo')).toBeInTheDocument()
    expect(screen.getByText('Fertig')).toBeInTheDocument()

    // Keine Progressbar und kein "Zahlungen"-Badge.
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
    expect(screen.queryByText(/Zahlungen/)).not.toBeInTheDocument()
    // Der kassierte Betrag bleibt sichtbar.
    expect(screen.getByText('15,00 €')).toBeInTheDocument()
  })

  it('zeigt die Kennzahlen in kanonischer Reihenfolge mit erklärenden Sub-Labels', () => {
    render(
      <LiveReportingSection
        liveData={liveData([])}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Erste Kachel ist Gesamtumsatz, danach folgt Bestellungen im DOM.
    const gesamtumsatz = screen.getByText('Gesamtumsatz')
    const bestellungen = screen.getByText('Bestellungen')
    expect(
      gesamtumsatz.compareDocumentPosition(bestellungen) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()

    // Sub-Labels benennen den Zusammenhang bestellt/kassiert.
    expect(
      screen.getByText('kassiert, abzüglich Warenrücknahmen'),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Bestellwert, inkl. noch nicht kassiert'),
    ).toBeInTheDocument()
  })

  it('zeigt "Stand HH:MM" und löst den Refresh-Button aus', async () => {
    const user = userEvent.setup()
    const onRefresh = vi.fn()
    const stand = new Date('2026-06-18T14:05:00Z').getTime()

    render(
      <LiveReportingSection
        liveData={liveData([])}
        loading={false}
        dataUpdatedAt={stand}
        onRefresh={onRefresh}
      />,
    )

    expect(screen.getByText(/^Stand \d{2}:\d{2}$/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Aktualisieren' }))
    expect(onRefresh).toHaveBeenCalledTimes(1)
  })
})
