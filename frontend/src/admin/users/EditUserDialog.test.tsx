import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { EditUserDialog } from './EditUserDialog'
import { type User } from './User'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

// Der eigene Account wird über das Auth-Singleton aufgelöst; im Test ist das
// die userId 1.
vi.mock('@/lib/Auth', () => ({
  AuthSingleton: {
    get userId() {
      return 1
    },
  },
}))

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

function renderDialog(overrides: Partial<User> = {}) {
  render(
    <EditUserDialog
      backend={{ updateUser: vi.fn().mockResolvedValue(undefined) }}
      open
      user={user(overrides)}
      updated={vi.fn()}
      close={vi.fn()}
    />,
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
    // Ohne die Sperre stuft sich der letzte Admin selbst herab und sperrt die
    // Instanz aus; das Backend lehnt das zusätzlich mit cannot_demote_self ab.
    renderDialog({ id: 1, role: 'admin' })

    expect(screen.getByRole('combobox', { name: /Rolle/ })).toBeDisabled()
  })

  it('lässt das Rollenfeld für andere Konten offen', () => {
    renderDialog()

    expect(screen.getByRole('combobox', { name: /Rolle/ })).toBeEnabled()
  })

  it('zeigt nur Abbrechen und Speichern im Footer', () => {
    renderDialog()

    expect(
      screen.getByRole('button', { name: 'Abbrechen' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Speichern' }),
    ).toBeInTheDocument()
  })
})
