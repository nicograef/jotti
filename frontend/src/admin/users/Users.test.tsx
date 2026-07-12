import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

// Der eigene Account (für den „das bist du“-Fall) wird über das Auth-Singleton
// aufgelöst; im Test wird die userId fest auf 1 gesetzt.
vi.mock('@/lib/Auth', () => ({
  AuthSingleton: {
    get userId() {
      return 1
    },
  },
}))

// Radix DropdownMenu misst seinen Anker über ResizeObserver, den jsdom nicht
// kennt. Ein No-op-Stub reicht für den Test.
class ResizeObserverStub {
  observe(): void {
    // no-op
  }
  unobserve(): void {
    // no-op
  }
  disconnect(): void {
    // no-op
  }
}

beforeAll(() => {
  vi.stubGlobal('ResizeObserver', ResizeObserverStub)
})

import { type User } from './User'
import { Users } from './Users'

function user(overrides: Partial<User> = {}): User {
  return {
    id: 2,
    name: 'Sophie Renz',
    username: 'sophie',
    role: 'service',
    status: 'active',
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
    ...overrides,
  }
}

function backend() {
  return {
    activateUser: vi.fn().mockResolvedValue(undefined),
    deactivateUser: vi.fn().mockResolvedValue(undefined),
    deleteUser: vi.fn().mockResolvedValue(undefined),
  }
}

function renderUsers(
  users: User[],
  overrides: Partial<Parameters<typeof Users>[0]> = {},
) {
  const be = overrides.backend ?? backend()
  const onResetPassword = vi.fn().mockResolvedValue(undefined)
  render(
    <Users
      loading={false}
      backend={be}
      users={users}
      onEdit={vi.fn()}
      onStatusChange={vi.fn()}
      onResetPassword={onResetPassword}
      onDeleted={vi.fn()}
      {...overrides}
    />,
  )
  return { be, onResetPassword }
}

afterEach(cleanup)

describe('Users', () => {
  it('renders labelled role badges instead of star symbols', () => {
    renderUsers([
      user({ id: 2, name: 'Nadine Admin', role: 'admin' }),
      user({ id: 3, name: 'Sophie Renz', role: 'serviceleitung' }),
      user({ id: 4, name: 'Felix Maier', role: 'service' }),
    ])

    expect(screen.getByText('Admin')).toBeInTheDocument()
    expect(screen.getByText('Serviceleitung')).toBeInTheDocument()
    expect(screen.getByText('Service')).toBeInTheDocument()
  })

  it('marks the own account with "das bist du" and offers no delete in its menu', async () => {
    const u = userEvent.setup()
    // id 1 ist laut gemocktem Auth-Singleton der eigene Account.
    renderUsers([user({ id: 1, name: 'Ich Selbst', role: 'admin' })])

    expect(screen.getByText('das bist du')).toBeInTheDocument()

    await u.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))
    const menu = screen.getByRole('menu')
    expect(
      within(menu).getByRole('menuitem', { name: /Passwort zurücksetzen/ }),
    ).toBeInTheDocument()
    expect(
      within(menu).queryByRole('menuitem', { name: /Löschen/ }),
    ).not.toBeInTheDocument()
  })

  it('offers delete for other accounts', async () => {
    const u = userEvent.setup()
    renderUsers([user({ id: 5, name: 'Felix Maier' })])

    await u.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))
    const menu = screen.getByRole('menu')
    expect(
      within(menu).getByRole('menuitem', { name: /Löschen/ }),
    ).toBeInTheDocument()
  })

  it('reaches password reset via the row menu', async () => {
    const u = userEvent.setup()
    const { onResetPassword } = renderUsers([
      user({ id: 5, name: 'Felix Maier' }),
    ])

    await u.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))
    await u.click(
      screen.getByRole('menuitem', { name: /Passwort zurücksetzen/ }),
    )

    expect(onResetPassword).toHaveBeenCalledWith(5)
  })

  it('deactivates a user via the status switch', async () => {
    const u = userEvent.setup()
    const { be } = renderUsers([
      user({ id: 7, name: 'Felix Maier', status: 'active' }),
    ])

    await u.click(screen.getByRole('switch', { name: /deaktivieren/i }))

    expect(be.deactivateUser).toHaveBeenCalledWith(7)
  })
})
