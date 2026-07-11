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
  stornierungenProServicekraft: LiveReportingData['breakdowns']['stornierungenProServicekraft'] = [],
  stornierungen: LiveReportingData['stornierungen'] = [],
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
    breakdowns: { servicekraefte, stornierungenProServicekraft },
    stornierungen,
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

  it('markiert Servicekraft-Zeilen mit Stornos und zeigt das Aggregat über der Detail-Liste', () => {
    render(
      <LiveReportingSection
        liveData={liveData(
          [
            {
              userId: 3,
              userName: 'felix',
              name: 'Felix W.',
              zahlungenCents: 1500,
              anzahlZahlungen: 2,
              offeneTische: [],
              erledigt: true,
            },
            {
              userId: 9,
              userName: 'cleo',
              name: '',
              zahlungenCents: 900,
              anzahlZahlungen: 1,
              offeneTische: [],
              erledigt: true,
            },
          ],
          [
            {
              userId: 3,
              userName: 'felix',
              name: 'Felix W.',
              anzahlStornierungen: 1,
              stornierungenCents: 500,
            },
            {
              userId: 7,
              userName: 'sophie',
              name: 'Sophie B.',
              anzahlStornierungen: 1,
              stornierungenCents: 250,
            },
          ],
          [
            {
              zeitpunkt: '2026-06-18T12:00:00Z',
              quelle: 'tisch',
              barRueckgabe: true,
              tischId: 9,
              tischName: 'Tisch 9',
              userId: 3,
              userName: 'felix',
              name: 'Felix W.',
              betragCents: 500,
              kommentar: '',
              positionen: [],
            },
          ],
        )}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Roter Marker an felix' Servicekraft-Zeile (hat Stornos), cleo ohne Marker.
    expect(screen.getByText('1 Storno')).toBeInTheDocument()

    // Aggregat-Zeile pro Servicekraft über der Detail-Liste.
    expect(
      screen.getByText('felix (Felix W.) 1 · sophie (Sophie B.) 1'),
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
