import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import { ServiceDock } from '../ServiceDock'
import { Bestellung } from './Bestellung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Diese Suite prüft das Handy-Layout (Dock-Aktionsbutton); ab lg rendert
// Bestellung die feste Abschluss-Spalte, deren Verhalten
// BestellungAbschluss.test.tsx container-neutral abdeckt.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

describe('Bestellung Aktionsleiste', () => {
  it('ist ohne Auswahl deaktiviert und zeigt nach Auswahl Anzahl und Summe', async () => {
    const user = userEvent.setup()
    render(
      <ServiceDock leiste={null}>
        <Bestellung
          backend={{
            bestellungAufnehmen: vi.fn().mockResolvedValue(undefined),
          }}
          tisch={tisch}
          products={[testProdukt]}
          productsLoading={false}
          onErfolg={vi.fn()}
        />
      </ServiceDock>,
    )

    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeDisabled()

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )

    const bar = screen.getByRole('button', { name: /Bestellung überprüfen/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('3,50')
  })
})
