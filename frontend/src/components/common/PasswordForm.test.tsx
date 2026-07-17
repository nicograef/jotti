import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'

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

function renderPassword(setPassword = vi.fn()) {
  const result = render(
    <MemoryRouter>
      <PasswordForm backend={{ setPassword }} />
    </MemoryRouter>,
  )
  return { setPassword, ...result }
}

describe('PasswordForm', () => {
  it('zeigt bei leerem Formular die Feldfehler und löst keinen Aufruf aus', async () => {
    const user = userEvent.setup()
    const { setPassword } = renderPassword()

    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    expect(
      await screen.findByText(
        'Benutzername muss mindestens 3 Zeichen lang sein.',
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Passwort muss mindestens 6 Zeichen lang sein.'),
    ).toBeInTheDocument()
    expect(setPassword).not.toHaveBeenCalled()
  })

  it('deaktiviert den Button während des Ladens (kein Doppel-Submit)', async () => {
    const user = userEvent.setup()
    // setPassword hängt, damit der Ladezustand während der Assertion aktiv bleibt.
    const setPassword = vi.fn(
      () =>
        new Promise<void>(() => {
          /* bleibt pending */
        }),
    )
    const { container } = renderPassword(setPassword)

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Neues Passwort'), 'geheim1')
    await user.type(otpInput(container), '123456')
    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Passwort festlegen/ }),
      ).toBeDisabled()
    })
    expect(setPassword).toHaveBeenCalledTimes(1)
  })

  it('submittet gültige Eingaben und ruft den Backend-Aufruf auf', async () => {
    const user = userEvent.setup()
    const setPassword = vi.fn().mockResolvedValue(undefined)
    const { container } = renderPassword(setPassword)

    await user.type(screen.getByPlaceholderText('Benutzername'), 'anna')
    await user.type(screen.getByPlaceholderText('Neues Passwort'), 'geheim1')
    await user.type(otpInput(container), '123456')
    await user.click(screen.getByRole('button', { name: /Passwort festlegen/ }))

    await waitFor(() => {
      expect(setPassword).toHaveBeenCalledWith('anna', 'geheim1', '123456')
    })
  })
})
