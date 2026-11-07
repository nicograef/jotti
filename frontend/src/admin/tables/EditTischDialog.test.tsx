import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { EditTischDialog } from './EditTischDialog'
import { type Tisch } from './Tisch'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function tisch(overrides: Partial<Tisch> = {}): Tisch {
  return {
    id: 1,
    name: 'Tisch 1',
    status: 'active',
    saldoCents: 0,
    ...overrides,
  }
}

function renderDialog(t: Tisch) {
  const fake = new FakeBackend()
    .respond('admin/update-tisch', {})
    .respond('admin/delete-tisch', {})
  renderWithBackend(
    <EditTischDialog
      tisch={t}
      updated={vi.fn()}
      deleted={vi.fn()}
      close={vi.fn()}
    />,
    fake,
  )
  return fake
}

afterEach(cleanup)

describe('EditTischDialog', () => {
  it('disables delete and shows the saldo reason when the tisch has an open saldo', () => {
    renderDialog(tisch({ saldoCents: 9850 }))

    const deleteButton = screen.getByRole('button', { name: /Tisch löschen/i })
    expect(deleteButton).toBeDisabled()
    expect(
      screen.getByText(/Offener Saldo: 98,50 € — erst abrechnen/),
    ).toBeInTheDocument()
  })

  it('deletes the tisch via the confirmation dialog when there is no saldo', async () => {
    const user = userEvent.setup()
    const fake = renderDialog(tisch({ id: 7, saldoCents: 0 }))

    await user.click(screen.getByRole('button', { name: /Tisch löschen/i }))
    await user.click(screen.getByRole('button', { name: 'Löschen' }))

    await waitFor(() => {
      expect(fake.bodies('admin/delete-tisch')).toEqual([{ id: 7 }])
    })
  })
})
