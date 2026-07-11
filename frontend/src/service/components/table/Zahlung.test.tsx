import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { Zahlung } from './Zahlung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1 },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

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

// Position einer anderen Servicekraft (bestellerUserId ≠ 1).
const fremdePosition: Position = {
  positionId: '00000000-0000-0000-0000-000000000002',
  varianteId: 2,
  produktName: 'Pommes',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 250,
  menge: 1,
  bestellerUserId: 2,
  bestellerName: 'Kollegin',
}

function renderZahlung(positionen: Position[] = [position]) {
  render(
    <Zahlung
      backend={{
        zahlungKassieren: vi.fn().mockResolvedValue(undefined),
      }}
      tisch={tisch}
      positionen={positionen}
      loading={false}
      onZahlungKassiert={vi.fn()}
    />,
  )
}

describe('Zahlung Aktionsleiste', () => {
  it('ist ohne Auswahl deaktiviert und zeigt nach Auswahl Anzahl und Summe', async () => {
    const user = userEvent.setup()
    renderZahlung()

    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()

    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))

    const bar = screen.getByRole('button', { name: /Kassieren/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('3,50')
  })
})

describe('Zahlung Positionsgruppen', () => {
  it('zeigt eigene Positionen direkt und fremde erst nach „Alle anzeigen"', async () => {
    const user = userEvent.setup()
    renderZahlung([position, fremdePosition])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(screen.queryByText('Pommes Normal')).not.toBeInTheDocument()

    await user.click(
      screen.getByRole('button', { name: /Alle anzeigen \(1 von/ }),
    )

    expect(screen.getByText('Pommes Normal')).toBeInTheDocument()
    expect(screen.getByText('von Kollegin')).toBeInTheDocument()
  })

  it('ohne fremde Positionen gibt es keinen „Alle anzeigen"-Schalter', () => {
    renderZahlung([position])

    expect(screen.getByText('Bratwurst Normal')).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Alle anzeigen/ }),
    ).not.toBeInTheDocument()
  })
})
