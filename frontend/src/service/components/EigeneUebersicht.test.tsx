import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import type { EigeneUebersicht } from '../table/Tisch'
import { EigeneUebersichtKarten } from './EigeneUebersicht'

afterEach(cleanup)

function uebersicht(
  overrides: Partial<EigeneUebersicht> = {},
): EigeneUebersicht {
  return {
    anzahlBestellungen: 4,
    bestellungenCents: 4200,
    anzahlZahlungen: 3,
    zahlungenCents: 10000,
    anzahlRuecknahmen: 0,
    ruecknahmenCents: 0,
    abzugebenCents: 10000,
    ...overrides,
  }
}

describe('EigeneUebersichtKarten', () => {
  it('zeigt ohne zugeordnete Rücknahme nur die zwei Kacheln', () => {
    render(<EigeneUebersichtKarten uebersicht={uebersicht()} loading={false} />)

    expect(screen.getByText('Bestellungen')).toBeInTheDocument()
    expect(screen.getByText('Kassiert')).toBeInTheDocument()
    expect(screen.getByText('· 100,00 €')).toBeInTheDocument()
    // Die Hinweiszeile bleibt aus — der Normalfall ist unverändert.
    expect(screen.queryByText(/Rücknahme/)).not.toBeInTheDocument()
    expect(screen.queryByText(/gibst/)).not.toBeInTheDocument()
  })

  it('erklärt eine zugeordnete Rücknahme und nennt den abzugebenden Betrag', () => {
    render(
      <EigeneUebersichtKarten
        uebersicht={uebersicht({
          anzahlRuecknahmen: 1,
          ruecknahmenCents: 1250,
          abzugebenCents: 8750,
        })}
        loading={false}
      />,
    )

    const hinweis = screen.getByText(/Rücknahme/)
    expect(hinweis).toHaveTextContent('Eine Rücknahme')
    expect(hinweis).toHaveTextContent('12,50 €')
    expect(hinweis).toHaveTextContent('87,50 €')
    // Die Kacheln zeigen weiterhin den kassierten Bruttobetrag.
    expect(screen.getByText('Kassiert')).toBeInTheDocument()
    expect(screen.getByText('· 100,00 €')).toBeInTheDocument()
  })

  it('formuliert mehrere Rücknahmen im Plural', () => {
    render(
      <EigeneUebersichtKarten
        uebersicht={uebersicht({
          anzahlRuecknahmen: 3,
          ruecknahmenCents: 2000,
          abzugebenCents: 8000,
        })}
        loading={false}
      />,
    )

    expect(screen.getByText(/Rücknahmen/)).toHaveTextContent('3 Rücknahmen')
  })

  it('zeigt im Ladezustand zwei Skeleton-Spalten und keine Hinweiszeile', () => {
    const { container } = render(
      <EigeneUebersichtKarten
        uebersicht={uebersicht({
          anzahlRuecknahmen: 1,
          ruecknahmenCents: 1250,
          abzugebenCents: 8750,
        })}
        loading
      />,
    )

    expect(screen.queryByText(/Rücknahme/)).not.toBeInTheDocument()
    expect(container.querySelectorAll('.grid-cols-2')).toHaveLength(1)
  })
})
