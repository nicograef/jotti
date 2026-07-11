import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { ReportingResults } from './ReportingResults'
import type { ReportingData } from './types'

const reportingResult: ReportingData = {
  kassensitzungNr: 1,
  summary: {
    gesamtUmsatzCents: 12345,
    gesamtBestellungenCents: 6500,
    gesamtStornierungenCents: 300,
    geldtransitCents: 0,
    anzahlBestellungen: 8,
    anzahlStornierungen: 1,
    anzahlDirektverkaeufe: 7,
    direktverkaufUmsatzCents: 4500,
  },
  breakdowns: {
    umsatzProServicekraft: [
      {
        userId: 5,
        userName: 'Bea',
        name: 'Bea B.',
        zahlungenCents: 6789,
        anzahlZahlungen: 3,
      },
    ],
    umsatzProTisch: [],
  },
  umsatzProSteuersatz: [
    {
      satz: 'regel',
      bruttoCents: 1190,
      nettoCents: 1000,
      steuerCents: 190,
    },
  ],
  stornierungen: [],
}

afterEach(() => {
  cleanup()
})

describe('ReportingResults', () => {
  it('shows the Direktverkauf summary card in overview', () => {
    render(<ReportingResults result={reportingResult} loading={false} />)

    expect(screen.getByText('Direktverkauf')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
    expect(screen.getByText('45,00 €')).toBeInTheDocument()
    expect(screen.getByText('Umsatz nach Steuersatz')).toBeInTheDocument()
    expect(screen.getByText('Regelsteuersatz (19 %)')).toBeInTheDocument()
  })

  it('zeigt die Steuersätze als gestapelte Blöcke mit beschrifteten Zeilen', () => {
    render(<ReportingResults result={reportingResult} loading={false} />)

    // Mobile-Politur: keine Tabelle mit horizontalem Scroll, sondern
    // gestapelte Blöcke mit beschrifteten Brutto/Netto/Steuer-Zeilen.
    expect(screen.getByText('Brutto')).toBeInTheDocument()
    expect(screen.getByText('Netto')).toBeInTheDocument()
    expect(screen.getByText('Steuer')).toBeInTheDocument()
    expect(screen.getByText('11,90 €')).toBeInTheDocument()
    expect(screen.getByText('10,00 €')).toBeInTheDocument()
    expect(screen.getByText('1,90 €')).toBeInTheDocument()
  })

  it('zeigt Servicekräfte ohne Progressbar und ohne Zahlungen-Badge', async () => {
    const user = userEvent.setup()
    render(<ReportingResults result={reportingResult} loading={false} />)

    await user.click(screen.getByRole('tab', { name: /Servicekräfte/ }))

    expect(screen.getByText('Bea (Bea B.)')).toBeInTheDocument()
    expect(screen.getByText('67,89 €')).toBeInTheDocument()
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
    expect(screen.queryByText(/Zahlungen/)).not.toBeInTheDocument()
  })
})
