import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ReportingResults } from './ReportingResults'
import type { ReportingData } from './types'

const reportingResult: ReportingData = {
  kassensitzungNr: 1,
  summary: {
    gesamtUmsatzCents: 12345,
    gesamtAuszahlungenCents: 1200,
    gesamtBestellungenCents: 6500,
    gesamtStornierungenCents: 300,
    anzahlBestellungen: 8,
    anzahlStornierungen: 1,
    anzahlDirektverkaeufe: 7,
    direktverkaufUmsatzCents: 4500,
  },
  breakdowns: {
    umsatzProServicekraft: [],
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

describe('ReportingResults', () => {
  it('shows the Direktverkauf summary card in overview', () => {
    render(<ReportingResults result={reportingResult} loading={false} />)

    expect(screen.getByText('Direktverkauf')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
    expect(screen.getByText('45,00 €')).toBeInTheDocument()
    expect(screen.getByText('Umsatz nach Steuersatz')).toBeInTheDocument()
    expect(screen.getByText('Regelsteuersatz (19 %)')).toBeInTheDocument()
  })
})
