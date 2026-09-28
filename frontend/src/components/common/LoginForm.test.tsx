import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { AuthBackend } from '@/lib/AuthBackend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { signOut, testToken } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'

import { LoginForm } from './LoginForm'

// A login that never answers keeps the loading state active during the assertion.
function haengenderLogin(): FakeBackend {
  return new FakeBackend().respond(
    'auth/login',
    () =>
      new Promise<never>(() => {
        /* bleibt pending */
      }),
  )
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  signOut()
})

function renderLogin(
  fake = new FakeBackend().respond('auth/login', { token: testToken() }),
) {
  render(
    <MemoryRouter>
      <LoginForm backend={new AuthBackend(fake)} />
    </MemoryRouter>,
  )
  return fake
}

describe('LoginForm', () => {
  it('zeigt bei leerem Formular die Feldfehler und löst keinen Login aus', async () => {
    const user = userEvent.setup()
    const fake = renderLogin()

    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    expect(
      await screen.findByText(
        'Benutzername muss mindestens 3 Zeichen lang sein.',
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Passwort muss mindestens 6 Zeichen lang sein.'),
    ).toBeInTheDocument()
    expect(fake.bodies('auth/login')).toHaveLength(0)
  })

  it('deaktiviert den Button während des Ladens (kein Doppel-Submit)', async () => {
    const user = userEvent.setup()
    const fake = renderLogin(haengenderLogin())

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Passwort'), 'geheim1')
    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Anmelden/ })).toBeDisabled()
    })
    expect(fake.bodies('auth/login')).toHaveLength(1)
  })

  it('submittet gültige Eingaben und ruft den Login-Backend-Aufruf auf', async () => {
    const user = userEvent.setup()
    const fake = renderLogin()

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Passwort'), 'geheim1')
    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    await waitFor(() => {
      expect(fake.bodies('auth/login')).toEqual([
        { username: 'anna', password: 'geheim1' },
      ])
    })
  })
})

describe('LoginForm im Vorgangs-Register', () => {
  // Bewusst gepinnt: Meldete sich das Anmeldeformular, wartete der erzwungene
  // Reload bis nach der Anmeldung und feuerte genau dort, wo er am meisten
  // stört; ohne Meldung greift er beim Aufschlagen der Anmeldeseite.
  it('meldet weder getippte Zugangsdaten noch den laufenden Login', async () => {
    const user = userEvent.setup()
    renderLogin(haengenderLogin())

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Passwort'), 'geheim1')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.click(screen.getByRole('button', { name: /Anmelden/ }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Anmelden/ })).toBeDisabled()
    })
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
