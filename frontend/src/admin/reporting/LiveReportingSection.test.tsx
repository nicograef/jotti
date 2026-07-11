import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { LiveReportingSection } from './LiveReportingSection'
import type { LiveReportingData } from './types'

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
      gesamtBestellungenCents: 0,
      gesamtStornierungenCents: 0,
      geldtransitCents: 0,
      anzahlBestellungen: 0,
      anzahlStornierungen: 0,
      anzahlDirektverkaeufe: 0,
      direktverkaufUmsatzCents: 0,
    },
    breakdowns: { servicekraefte, umsatzProTisch: [] },
    stornierungen: [],
  }
}

describe('LiveReportingSection — Servicekräfte', () => {
  it('zeigt pro Servicekraft offene Arbeit bzw. einen Fertig-Hinweis', async () => {
    const user = userEvent.setup()
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
      />,
    )

    await user.click(screen.getByRole('tab', { name: /Servicekräfte/ }))

    // Servicekraft mit offener Arbeit: Tisch + Anzahl zu kassierender Positionen.
    expect(screen.getByText('Anna (Anna A.)')).toBeInTheDocument()
    expect(screen.getByText('Tisch 3')).toBeInTheDocument()
    expect(screen.getByText(/2 zu kassieren/)).toBeInTheDocument()

    // Fertige Servicekraft: nur Hinweis, keine offenen Tische.
    expect(screen.getByText('Cleo')).toBeInTheDocument()
    expect(screen.getByText('Fertig')).toBeInTheDocument()

    // Mobile-Politur: keine Progressbar und kein "Zahlungen"-Badge mehr.
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
    expect(screen.queryByText(/Zahlungen/)).not.toBeInTheDocument()
    // Der kassierte Betrag bleibt sichtbar.
    expect(screen.getByText('15,00 €')).toBeInTheDocument()
  })
})
