import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { ZahlungAbschluss } from './ZahlungAbschluss'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 900 }

const position: Position = {
  positionId: '00000000-0000-0000-0000-000000000001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 350,
  menge: 2,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

// Die feste Spalte (variant="spalte") ist container-neutral testbar: kein
// Drawer, kein Dock.
function renderSpalte(
  zahlungKassieren: () => Promise<void> = vi.fn().mockResolvedValue(undefined),
) {
  const zahlungKassiert = vi.fn()
  render(
    <ZahlungAbschluss
      variant="spalte"
      backend={{ zahlungKassieren }}
      tisch={tisch}
      positionenToPay={[{ ...position, menge: 2 }]}
      totalCents={700}
      restNachZahlungCents={200}
      zahlungKassiert={zahlungKassiert}
    />,
  )
  return { zahlungKassieren, zahlungKassiert }
}

describe('ZahlungAbschluss (Spalte)', () => {
  it('zeigt im Leerzustand einen Hinweis und deaktiviert den Aktionsbutton', () => {
    render(
      <ZahlungAbschluss
        variant="spalte"
        backend={{ zahlungKassieren: vi.fn() }}
        tisch={tisch}
        positionenToPay={[]}
        totalCents={0}
        restNachZahlungCents={900}
        zahlungKassiert={vi.fn()}
      />,
    )

    expect(screen.getByText(/Positionen auswählen/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeDisabled()
  })

  it('zeigt die Restbetrag-Zeile in der Spalte', () => {
    renderSpalte()
    expect(screen.getByText('Nach dieser Zahlung noch offen')).toBeVisible()
    expect(screen.getByText('2,00 €')).toBeInTheDocument()
  })

  it('rechnet Rückgeld ohne Zielbetrag und Trinkgeld mit Zielbetrag', async () => {
    const user = userEvent.setup()
    renderSpalte()

    // Ohne Zielbetrag: Rückgeld = Erhalten − Gesamt.
    await user.type(screen.getByLabelText('Erhalten'), '10,00')
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '3,00',
    )

    // Mit Zielbetrag (Aufrundung): Trinkgeld = Zielbetrag − Gesamt.
    await user.type(screen.getByLabelText('Zahlbetrag inkl. Trinkgeld'), '8,00')
    expect(screen.getByText('Trinkgeld').nextElementSibling).toHaveTextContent(
      '1,00',
    )
    expect(screen.getByText('Rückgeld').nextElementSibling).toHaveTextContent(
      '2,00',
    )
  })

  it('kassiert mit genau einem Backend-Call und der erwarteten Nutzlast (ohne Schlüssel)', async () => {
    const user = userEvent.setup()
    const { zahlungKassieren, zahlungKassiert } = renderSpalte()

    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    expect(zahlungKassieren).toHaveBeenCalledWith({
      tischId: 1,
      positionen: [{ positionId: position.positionId, menge: 2 }],
      kommentar: '',
    })
    expect(zahlungKassiert).toHaveBeenCalledTimes(1)
  })

  it('löst bei Doppelklick keinen zweiten Aufruf aus', async () => {
    const user = userEvent.setup()
    let resolve: () => void = () => undefined
    const zahlungKassieren = vi.fn(
      () =>
        new Promise<void>((r) => {
          resolve = r
        }),
    )
    render(
      <ZahlungAbschluss
        variant="spalte"
        backend={{ zahlungKassieren }}
        tisch={tisch}
        positionenToPay={[{ ...position, menge: 2 }]}
        totalCents={700}
        restNachZahlungCents={200}
        zahlungKassiert={vi.fn()}
      />,
    )

    const button = screen.getByRole('button', { name: 'Kassieren' })
    await user.click(button)
    // Zweiter Klick während des laufenden Submits: der Loading-Guard
    // deaktiviert den Button, es darf kein zweiter Aufruf entstehen.
    await user.click(button)
    resolve()

    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
  })
})
