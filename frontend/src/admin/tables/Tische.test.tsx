import { cleanup, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { type Tisch } from './Tisch'
import { Tische } from './Tische'

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

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('admin/activate-tisch', {})
    .respond('admin/deactivate-tisch', {})
}

function renderTische(tische: Tisch[], onEdit = vi.fn()) {
  const fake = backend()
  renderWithBackend(
    <Tische
      loading={false}
      tische={tische}
      onEdit={onEdit}
      onStatusChange={vi.fn()}
    />,
    fake,
  )
  return fake
}

afterEach(cleanup)

describe('Tische', () => {
  it('renders tische grouped by name prefix with group headings', () => {
    renderTische([
      tisch({ id: 1, name: 'Zelt 1' }),
      tisch({ id: 2, name: 'Zelt 2' }),
      tisch({ id: 3, name: 'Biergarten 1' }),
    ])

    const headings = screen.getAllByRole('heading', { level: 2 })
    expect(headings.map((h) => h.textContent)).toEqual(['Zelt', 'Biergarten'])
    expect(screen.getByText('Zelt 1')).toBeInTheDocument()
    expect(screen.getByText('Biergarten 1')).toBeInTheDocument()
  })

  it('shows the open amount and disables the switch when a tisch has a saldo', () => {
    renderTische([tisch({ id: 1, name: 'Tisch 4', saldoCents: 9850 })])

    expect(screen.getByText('98,50 € offen')).toBeInTheDocument()
    expect(screen.getByRole('switch')).toBeDisabled()
    expect(
      screen.getByText('Erst abrechnen, dann deaktivieren'),
    ).toBeInTheDocument()
  })

  it('deactivates a tisch via its switch', async () => {
    const user = userEvent.setup()
    const fake = renderTische([
      tisch({ id: 7, name: 'Tisch 7', status: 'active' }),
    ])

    await user.click(screen.getByRole('switch', { name: /deaktivieren/i }))

    await waitFor(() => {
      expect(fake.bodies('admin/deactivate-tisch')).toEqual([{ id: 7 }])
    })
  })

  it('opens the edit dialog when a kachel is clicked, not when the switch is toggled', async () => {
    const user = userEvent.setup()
    const onEdit = vi.fn()
    const tisch3 = tisch({ id: 3, name: 'Tisch 3' })
    renderTische([tisch3], onEdit)

    await user.click(screen.getByText('Tisch 3'))
    expect(onEdit).toHaveBeenCalledWith(tisch3)

    onEdit.mockClear()
    await user.click(screen.getByRole('switch'))
    expect(onEdit).not.toHaveBeenCalled()
  })

  it('collects tische without a trailing number under "Weitere"', () => {
    renderTische([
      tisch({ id: 1, name: 'Eingang' }),
      tisch({ id: 2, name: 'Zelt 1' }),
    ])

    const weitere = screen.getByRole('heading', { name: 'Weitere' })
    const section = weitere.closest('div')
    expect(section).not.toBeNull()
    expect(
      within(section as HTMLElement).getByText('Eingang'),
    ).toBeInTheDocument()
  })
})
