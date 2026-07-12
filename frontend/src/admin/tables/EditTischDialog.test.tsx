import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { EditTischDialog } from './EditTischDialog'
import { type Tisch } from './Tisch'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function tisch(overrides: Partial<Tisch> = {}): Tisch {
  return {
    id: 1,
    name: 'Tisch 1',
    status: 'active',
    saldoCents: 0,
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
    ...overrides,
  }
}

function backend() {
  return {
    updateTisch: vi.fn().mockResolvedValue(undefined),
    deleteTisch: vi.fn().mockResolvedValue(undefined),
  }
}

function renderDialog(t: Tisch, be = backend()) {
  render(
    <EditTischDialog
      backend={be}
      open
      tisch={t}
      updated={vi.fn()}
      deleted={vi.fn()}
      close={vi.fn()}
    />,
  )
  return be
}

afterEach(cleanup)

describe('EditTischDialog', () => {
  it('disables delete and shows the saldo reason when the tisch has an open saldo', () => {
    renderDialog(tisch({ saldoCents: 9850 }))

    const deleteButton = screen.getByRole('button', { name: /Tisch löschen/i })
    expect(deleteButton).toBeDisabled()
    // Die Begründung mit dem offenen Betrag steht als stets sichtbare Zeile.
    expect(
      screen.getByText(/Offener Saldo: 98,50 € — erst abrechnen/),
    ).toBeInTheDocument()
  })

  it('deletes the tisch via the confirmation dialog when there is no saldo', async () => {
    const user = userEvent.setup()
    const be = renderDialog(tisch({ id: 7, saldoCents: 0 }))

    // Ohne Saldo öffnet der Löschen-Button den Bestätigungsdialog.
    await user.click(screen.getByRole('button', { name: /Tisch löschen/i }))
    await user.click(screen.getByRole('button', { name: 'Löschen' }))

    expect(be.deleteTisch).toHaveBeenCalledWith(7)
  })
})
