import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { HistorieUmbuchungDrawer } from './HistorieUmbuchungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('../../table/hooks', () => ({
  useAktiveTische: () => ({
    tische: [
      { id: 1, name: 'Stammtisch', saldoCents: 0 },
      { id: 2, name: 'Nebentisch', saldoCents: 0 },
    ],
    isPending: false,
  }),
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

describe('HistorieUmbuchungDrawer', () => {
  it('rendert Positionsliste und Ziel-Tisch-Auswahl im DrawerBody, Buttons im Footer außerhalb', () => {
    render(
      <HistorieUmbuchungDrawer
        backend={{ bestellungUmbuchen: vi.fn().mockResolvedValue(undefined) }}
        tisch={tisch}
        vorgangId="00000000-0000-0000-0000-000000000042"
        positionen={[position]}
        onClose={vi.fn()}
        onBestellungUmgebucht={vi.fn()}
      />,
    )

    const dialog = screen.getByRole('dialog')
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    expect(body).not.toBeNull()
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    expect(body).toContainElement(screen.getByRole('combobox'))
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )
    expect(body).not.toContainElement(
      screen.getByRole('button', { name: 'Abbrechen' }),
    )
  })
})
