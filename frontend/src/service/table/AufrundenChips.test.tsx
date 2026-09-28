import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { AufrundenChips } from './AufrundenChips'

afterEach(() => {
  cleanup()
})

// Kontrollierter Harness: die Chips halten ihren Zielbetrag-/Anderer-State beim
// Aufrufer, deshalb spiegelt der Test genau dieses Zusammenspiel.
function Harness({ gesamtCents = 1230 }: { gesamtCents?: number }) {
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)
  return (
    <>
      <output data-testid="ziel">{zielbetragEuro}</output>
      <AufrundenChips
        gesamtCents={gesamtCents}
        zielbetragEuro={zielbetragEuro}
        onZielbetragEuroChange={setZielbetragEuro}
        andererAktiv={andererAktiv}
        onAndererAktivChange={setAndererAktiv}
      />
    </>
  )
}

describe('AufrundenChips', () => {
  it('zeigt „genau", zwei Vorschläge und „Anderer …"; „genau" ist anfangs aktiv', () => {
    render(<Harness />)

    const genau = screen.getByRole('button', { name: /genau/ })
    expect(genau).toHaveTextContent('12,30 € genau')
    expect(genau).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: /13,00/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /15,00/ })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Anderer …' }),
    ).toBeInTheDocument()
  })

  it('setzt den Zielbetrag bei Chip-Tap und wählt bei erneutem Tap ab', async () => {
    const user = userEvent.setup()
    render(<Harness />)

    const vorschlag = screen.getByRole('button', { name: /13,00/ })
    await user.click(vorschlag)
    expect(screen.getByTestId('ziel')).toHaveTextContent('13,00')
    expect(vorschlag).toHaveAttribute('aria-pressed', 'true')

    // Erneuter Tap auf den aktiven Chip: zurück zu „genau".
    await user.click(vorschlag)
    expect(screen.getByTestId('ziel')).toHaveTextContent('')
    expect(screen.getByRole('button', { name: /genau/ })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })

  it('blendet über „Anderer …" das freie Euro-Feld ein', async () => {
    const user = userEvent.setup()
    render(<Harness />)

    expect(
      screen.queryByLabelText('Zahlbetrag inkl. Trinkgeld'),
    ).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Anderer …' }))

    const feld = screen.getByLabelText('Zahlbetrag inkl. Trinkgeld')
    expect(feld).toBeInTheDocument()
    await user.type(feld, '14,50')
    expect(screen.getByTestId('ziel')).toHaveTextContent('14,50')
  })
})
