import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Bestellung, Position } from '../../table/Bestellung'
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

function renderDrawer(
  bestellungUmbuchen = vi.fn().mockResolvedValue(undefined),
) {
  render(
    <HistorieUmbuchungDrawer
      backend={{ bestellungUmbuchen }}
      tisch={tisch}
      quelle={quelle}
      onClose={vi.fn()}
      onBestellungUmgebucht={vi.fn()}
    />,
  )
  return bestellungUmbuchen
}

describe('HistorieUmbuchungDrawer', () => {
  it('rendert Positionsliste und Ziel-Tisch-Auswahl im DrawerBody, Buttons im Footer außerhalb', () => {
    renderDrawer()

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

  it('sperrt die Umbuchung, bis ein Ziel-Tisch aktiv gewählt wurde', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = renderDrawer()

    // Ohne Wahl steht der Placeholder, nicht ein vorbelegter Tisch.
    const select = screen.getByRole('combobox')
    expect(select).toHaveValue('')
    expect(screen.getByText('Ziel-Tisch wählen…')).toBeInTheDocument()

    const button = screen.getByRole('button', { name: 'Umbuchung ausführen' })
    expect(button).toBeDisabled()

    await user.selectOptions(select, 'Nebentisch')
    expect(button).toBeEnabled()

    await user.click(button)
    expect(bestellungUmbuchen).toHaveBeenCalledWith(
      expect.objectContaining({ quellTischId: 1, zielTischId: 2 }),
    )
  })

  it('titelt menschenlesbar mit Vorgangstyp und Name statt UUID-Fragment', () => {
    renderDrawer()

    const title = screen.getByText(/^Bestellung ·/)
    expect(title).toHaveTextContent('Nico')
    expect(screen.queryByText(/00000000/)).not.toBeInTheDocument()
  })

  it('reicht ein optionales Kommentar an die Umbuchung durch', async () => {
    const user = userEvent.setup()
    const bestellungUmbuchen = renderDrawer()

    await user.type(
      screen.getByPlaceholderText('Kommentar (optional)'),
      'Gast gewechselt',
    )
    await user.selectOptions(screen.getByRole('combobox'), 'Nebentisch')
    await user.click(
      screen.getByRole('button', { name: 'Umbuchung ausführen' }),
    )

    expect(bestellungUmbuchen).toHaveBeenCalledWith(
      expect.objectContaining({
        quellTischId: 1,
        zielTischId: 2,
        benutzerKommentar: 'Gast gewechselt',
      }),
    )
  })
})
