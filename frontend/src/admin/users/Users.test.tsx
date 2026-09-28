import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  afterEach,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from 'vitest'

import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'

import { type User } from './User'
import { UserBackend } from './UserBackend'
import { Users } from './Users'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

// The own account comes from the auth singleton; in the test the userId is 1.
beforeEach(() => {
  signIn({ userId: 1, role: 'admin' })
})

// Radix DropdownMenu misst seinen Anker über ResizeObserver, den jsdom nicht kennt.
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

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('admin/activate-user', {})
    .respond('admin/deactivate-user', {})
    .respond('admin/delete-user', {})
}

function renderUsers(
  users: User[],
  overrides: Partial<Parameters<typeof Users>[0]> = {},
) {
  const fake = backend()
  const onResetPassword = vi.fn().mockResolvedValue(undefined)
  render(
    <Users
      loading={false}
      backend={new UserBackend(fake)}
      users={users}
      onEdit={vi.fn()}
      onStatusChange={vi.fn()}
      onResetPassword={onResetPassword}
      onDeleted={vi.fn()}
      {...overrides}
    />,
  )
  return { fake, onResetPassword }
}

afterEach(() => {
  cleanup()
  signOut()
})

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

  it('locks the status switch on the own account', () => {
    renderUsers([user({ id: 1, name: 'Ich Selbst', role: 'admin' })])

    expect(screen.getByRole('switch', { name: /deaktivieren/i })).toBeDisabled()
  })

  it('deactivates a user via the status switch', async () => {
    const u = userEvent.setup()
    const { fake } = renderUsers([
      user({ id: 7, name: 'Felix Maier', status: 'active' }),
    ])

    await u.click(screen.getByRole('switch', { name: /deaktivieren/i }))

    expect(fake.bodies('admin/deactivate-user')).toEqual([{ id: 7 }])
  })
})
