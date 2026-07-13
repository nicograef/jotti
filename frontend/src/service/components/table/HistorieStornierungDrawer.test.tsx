import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung, Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { HistorieStornierungDrawer } from './HistorieStornierungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
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

const quelle: Bestellung = {
  art: 'bestellung',
  id: '00000000-0000-0000-0000-000000000042',
  userId: 1,
  userName: 'Nico',
  tischId: 1,
  positionen: [position],
  gesamtPreisCents: 700,
  kommentar: '',
  aufgenommenAm: '2026-06-18T12:00:00Z',
  stornierbarePositionen: [position],
  umbuchbarePositionen: [position],
}

describe('HistorieStornierungDrawer', () => {
  it('rendert Positionsliste im DrawerBody; Pflicht-Kommentar und Buttons im sichtbaren Footer', () => {
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    const footer = dialog.querySelector('[data-slot="drawer-footer"]')
    expect(body).not.toBeNull()
    expect(footer).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    // Das Pflichtfeld steht im nicht-scrollenden Footer, nicht im Body.
    const kommentar = screen.getByPlaceholderText('Kommentar (erforderlich)')
    expect(footer).toContainElement(kommentar)
    expect(body).not.toContainElement(kommentar)
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )
  })

  it('titelt menschenlesbar mit Vorgangstyp und Name statt UUID-Fragment', () => {
    render(
      <HistorieStornierungDrawer
        backend={{ stornierungErteilen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        quelle={quelle}
        onClose={vi.fn()}
        onStornierungErteilt={vi.fn()}
      />,
    )

    const title = screen.getByText(/^Bestellung ·/)
    expect(title).toBeInTheDocument()
    expect(title).toHaveTextContent('Nico')
    expect(screen.queryByText(/00000000/)).not.toBeInTheDocument()
  })
})
