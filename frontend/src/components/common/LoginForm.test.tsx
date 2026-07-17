import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { LoginForm } from './LoginForm'

// AuthSingleton dekodiert echte JWTs; im Happy-Path-Test wird der Token-Schritt
// neutralisiert, damit der Test nur den Formular-Submit prüft.
const authState = vi.hoisted<{ isAdmin: boolean }>(() => ({ isAdmin: false }))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: {
    validateAndSetToken: vi.fn(),
    get isAdmin() {
      return authState.isAdmin
    },
  },
}))

afterEach(() => {
  cleanup()
  authState.isAdmin = false
})

function renderLogin(login = vi.fn()) {
  render(
    <MemoryRouter>
      <LoginForm backend={{ login }} />
    </MemoryRouter>,
  )
  return login
}

describe('LoginForm', () => {
  it('zeigt bei leerem Formular die Feldfehler und löst keinen Login aus', async () => {
    const user = userEvent.setup()
    const login = renderLogin()

    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    expect(
      await screen.findByText(
        'Benutzername muss mindestens 3 Zeichen lang sein.',
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Passwort muss mindestens 6 Zeichen lang sein.'),
    ).toBeInTheDocument()
    expect(login).not.toHaveBeenCalled()
  })

  it('deaktiviert den Button während des Ladens (kein Doppel-Submit)', async () => {
    const user = userEvent.setup()
    // Login hängt, damit der Ladezustand während der Assertion aktiv bleibt.
    const login = vi.fn(
      () =>
        new Promise<string>(() => {
          /* bleibt pending */
        }),
    )
    renderLogin(login)

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Passwort'), 'geheim1')
    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Anmelden/ })).toBeDisabled()
    })
    expect(login).toHaveBeenCalledTimes(1)
  })

  it('submittet gültige Eingaben und ruft den Login-Backend-Aufruf auf', async () => {
    const user = userEvent.setup()
    const login = vi.fn().mockResolvedValue('token')
    renderLogin(login)

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Passwort'), 'geheim1')
    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith('anna', 'geheim1')
    })
  })
})
