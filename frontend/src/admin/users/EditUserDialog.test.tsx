import { cleanup, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { EditUserDialog } from './EditUserDialog'
import { type User } from './User'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function user(overrides: Partial<User> = {}): User {
  return {
    id: 2,
    name: 'Sophie Renz',
    username: 'sophie',
    role: 'service',
    status: 'active',
    ...overrides,
  }
}

function renderDialog(isSelf = false) {
  renderWithBackend(
    <EditUserDialog
      user={user()}
      isSelf={isSelf}
      updated={vi.fn()}
      close={vi.fn()}
    />,
    new FakeBackend(),
  )
}

afterEach(cleanup)

describe('EditUserDialog', () => {
  it('bietet keinen Passwort-Reset-Einstieg im Footer an (nur das Zeilen-Menü)', () => {
    renderDialog()

    // Der Passwort-Reset läuft ausschließlich über das „···“-Zeilenmenü der
    // Benutzertabelle; der Bearbeiten-Dialog darf keinen zweiten Einstieg haben.
    expect(
      screen.queryByRole('button', { name: /Passwort zurücksetzen/ }),
    ).not.toBeInTheDocument()
  })

  it('sperrt das Rollenfeld im eigenen Konto', () => {
    renderDialog(true)

    expect(screen.getByRole('combobox', { name: /Rolle/ })).toBeDisabled()
  })

  it('lässt das Rollenfeld für andere Konten offen', () => {
    renderDialog()

    expect(screen.getByRole('combobox', { name: /Rolle/ })).toBeEnabled()
  })
})
