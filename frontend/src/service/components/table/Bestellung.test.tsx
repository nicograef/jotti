import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import { Bestellung } from './Bestellung'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// vaul's Drawer braucht Browser-APIs, die jsdom nicht bereitstellt. Der Trigger
// (Sticky-Leiste) wird inline gerendert, der Drawer-Inhalt ausgeblendet — so
// bleibt nur die Aktionsleiste übrig, die hier geprüft wird.
vi.mock('@/components/ui/drawer', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return {
    Drawer: Passthrough,
    DrawerTrigger: Passthrough,
    DrawerContent: () => null,
    DrawerHeader: Passthrough,
    DrawerTitle: Passthrough,
    DrawerDescription: Passthrough,
    DrawerFooter: Passthrough,
    DrawerClose: Passthrough,
  }
})

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
      <Bestellung
        backend={{ bestellungAufnehmen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        products={[testProdukt]}
        productsLoading={false}
        onBestellungAufgenommen={vi.fn()}
      />,
    )

    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeDisabled()

    await user.click(screen.getByText('Bratwurst'))
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )

    const bar = screen.getByRole('button', { name: /Bestellung überprüfen/ })
    expect(bar).toBeEnabled()
    expect(bar).toHaveTextContent('3,50')
  })
})
