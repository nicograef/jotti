import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import { BestellungDrawer } from './BestellungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
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

function renderDrawer(mengen: Record<number, number>) {
  render(
    <BestellungDrawer
      backend={{ bestellungAufnehmen: vi.fn().mockResolvedValue(undefined) }}
      tisch={tisch}
      products={[testProdukt]}
      mengen={mengen}
      bestellungAufgenommen={vi.fn()}
    />,
  )
}

describe('BestellungDrawer', () => {
  it('öffnet bei leerer Auswahl nicht (onOpenChange-Guard)', async () => {
    const user = userEvent.setup()
    renderDrawer({})

    // Klick auf die Trigger-Fläche (das Wurzel-Div der Aktionsleiste), nicht
    // auf den deaktivierten Button — der Guard muss das Öffnen verhindern.
    const bar = screen
      .getByRole('button', { name: /Bestellung überprüfen/ })
      .closest('div[aria-haspopup="dialog"]')
    if (bar === null) throw new Error('Trigger-Fläche nicht gefunden')
    await user.click(bar)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('öffnet mit Auswahl über den Trigger; Footer-Buttons liegen außerhalb des Scrollbereichs', async () => {
    const user = userEvent.setup()
    renderDrawer({ 1: 2 })

    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )

    const dialog = await screen.findByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    expect(body).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })
})
