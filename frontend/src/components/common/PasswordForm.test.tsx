import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeAll, describe, expect, it } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { PasswordForm } from './PasswordForm'

// input-otp registriert intern einen ResizeObserver; jsdom bringt keinen mit.
class ResizeObserverStub {
  observe() {
    /* no-op */
  }
  unobserve() {
    /* no-op */
  }
  disconnect() {
    /* no-op */
  }
}

beforeAll(() => {
  globalThis.ResizeObserver = ResizeObserverStub
  // input-otp ruft nach Fokus/Eingabe in einem Timer document.elementFromPoint
  // auf; jsdom implementiert es nicht.
  document.elementFromPoint = () => null
})

afterEach(() => {
  cleanup()
})

function otpInput(container: HTMLElement): HTMLInputElement {
  const input = container.querySelector('[data-input-otp]')
  if (!(input instanceof HTMLInputElement)) {
    throw new Error('OTP-Eingabefeld nicht gefunden')
  }
  return input
}

function renderPassword(
  fake = new FakeBackend().respond('auth/set-password', {}),
) {
  const result = renderWithBackend(
    <MemoryRouter>
      <PasswordForm />
    </MemoryRouter>,
    fake,
  )
  return { fake, ...result }
}

describe('PasswordForm', () => {
  it('zeigt bei leerem Formular die Feldfehler und löst keinen Aufruf aus', async () => {
    const user = userEvent.setup()
    const { fake } = renderPassword()

    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    expect(
      await screen.findByText(
        'Benutzername muss mindestens 3 Zeichen lang sein.',
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Passwort muss mindestens 6 Zeichen lang sein.'),
    ).toBeInTheDocument()
    expect(fake.bodies('auth/set-password')).toHaveLength(0)
  })

  it('deaktiviert den Button während des Ladens (kein Doppel-Submit)', async () => {
    const user = userEvent.setup()
    // Die Anfrage hängt, damit der Ladezustand während der Assertion aktiv bleibt.
    const { fake, container } = renderPassword(
      new FakeBackend().respond(
        'auth/set-password',
        () =>
          new Promise<never>(() => {
            /* bleibt pending */
          }),
      ),
    )

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Neues Passwort'), 'geheim1')
    await user.type(otpInput(container), '123456')
    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Passwort festlegen/ }),
      ).toBeDisabled()
    })
    expect(fake.bodies('auth/set-password')).toHaveLength(1)
  })

  it('submittet gültige Eingaben und ruft den Backend-Aufruf auf', async () => {
    const user = userEvent.setup()
    const { fake, container } = renderPassword()

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Neues Passwort'), 'geheim1')
    await user.type(otpInput(container), '123456')
    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    await waitFor(() => {
      expect(fake.bodies('auth/set-password')).toEqual([
        { username: 'anna', password: 'geheim1', onetimePassword: '123456' },
      ])
    })
  })
})
