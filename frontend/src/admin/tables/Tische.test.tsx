import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

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

function backend() {
  return {
    activateTisch: vi.fn().mockResolvedValue(undefined),
    deactivateTisch: vi.fn().mockResolvedValue(undefined),
  }
}

function renderTische(tische: Tisch[], be = backend()) {
  render(
    <Tische
      loading={false}
      backend={be}
      tische={tische}
      onEdit={vi.fn()}
      onStatusChange={vi.fn()}
    />,
  )
  return be
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
    // Der Switch ist gesperrt (Backend erzwingt es zusätzlich als SSOT).
    expect(screen.getByRole('switch')).toBeDisabled()
    // Die Begründung steht als stets sichtbare Zeile (kein Hover-Tooltip).
    expect(
      screen.getByText('Erst abrechnen, dann deaktivieren'),
    ).toBeInTheDocument()
  })

  it('deactivates a tisch via its switch', async () => {
    const user = userEvent.setup()
    const be = renderTische([
      tisch({ id: 7, name: 'Tisch 7', status: 'active' }),
    ])

    await user.click(screen.getByRole('switch', { name: /deaktivieren/i }))

    expect(be.deactivateTisch).toHaveBeenCalledWith(7)
  })

  it('opens the edit dialog when a kachel is clicked, not when the switch is toggled', async () => {
    const user = userEvent.setup()
    const onEdit = vi.fn()
    const be = backend()
    render(
      <Tische
        loading={false}
        backend={be}
        tische={[tisch({ id: 3, name: 'Tisch 3' })]}
        onEdit={onEdit}
        onStatusChange={vi.fn()}
      />,
    )

    // Klick auf den Kachel-Namen öffnet den Edit-Dialog.
    await user.click(screen.getByText('Tisch 3'))
    expect(onEdit).toHaveBeenCalledWith(3)

    // Der Switch-Klick löst kein zusätzliches onEdit aus (stopPropagation).
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
