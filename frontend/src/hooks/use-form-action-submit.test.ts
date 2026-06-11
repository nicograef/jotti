import { act, renderHook } from '@testing-library/react'
import type { UseFormReturn } from 'react-hook-form'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'

import { useFormActionSubmit } from './use-form-action-submit'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

function createFormMock() {
  return {
    setError: vi.fn(),
  } as unknown as UseFormReturn
}

afterEach(() => {
  vi.clearAllMocks()
})

describe('useFormActionSubmit', () => {
  it('sets field errors for validation_error details and shows no toast', async () => {
    const form = createFormMock()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Speichern',
      }),
    )

    await act(async () => {
      await result.current.run(() =>
        Promise.reject(
          new BackendError(400, 'validation_error', {
            name: ['Name ist erforderlich'],
            username: ['Benutzername ist erforderlich'],
          }),
        ),
      )
    })

    expect(form.setError).toHaveBeenCalledTimes(2)
    expect(form.setError).toHaveBeenNthCalledWith(1, 'name', {
      message: 'Name ist erforderlich',
    })
    expect(form.setError).toHaveBeenNthCalledWith(2, 'username', {
      message: 'Benutzername ist erforderlich',
    })
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('shows validation fallback toast when validation_error has no usable details', async () => {
    const form = createFormMock()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Speichern',
      }),
    )

    await act(async () => {
      await result.current.run(() =>
        Promise.reject(new BackendError(400, 'validation_error')),
      )
    })

    expect(form.setError).not.toHaveBeenCalled()
    expect(toast.error).toHaveBeenCalledWith(
      'Bitte die markierten Eingaben prüfen.',
    )
  })

  it('shows toast for non-validation backend errors and does not set field errors', async () => {
    const form = createFormMock()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Speichern',
      }),
    )

    await act(async () => {
      await result.current.run(() =>
        Promise.reject(new BackendError(409, 'conflict')),
      )
    })

    expect(form.setError).not.toHaveBeenCalled()
    expect(toast.error).toHaveBeenCalledWith(
      'Die Daten wurden gerade von jemand anderem geändert. Bitte aktualisieren und erneut versuchen.',
    )
  })

  it('sets field errors based on fieldErrorsByCode without showing toast', async () => {
    const form = createFormMock()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Passwort setzen',
        fieldErrorsByCode: {
          invalid_credentials: {
            username: 'Benutzername oder Code ungültig.',
            onetimePassword: 'Benutzername oder Code ungültig.',
          },
        },
      }),
    )

    await act(async () => {
      await result.current.run(() =>
        Promise.reject(new BackendError(401, 'invalid_credentials')),
      )
    })

    expect(form.setError).toHaveBeenCalledTimes(2)
    expect(form.setError).toHaveBeenNthCalledWith(1, 'username', {
      message: 'Benutzername oder Code ungültig.',
    })
    expect(form.setError).toHaveBeenNthCalledWith(2, 'onetimePassword', {
      message: 'Benutzername oder Code ungültig.',
    })
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('sets inline name error for *_already_exists using byCode without showing toast', async () => {
    const form = createFormMock()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Produkt anlegen',
        byCode: {
          produkt_already_exists: 'Dieser Name ist bereits vergeben.',
        },
      }),
    )

    await act(async () => {
      await result.current.run(() =>
        Promise.reject(new BackendError(409, 'produkt_already_exists')),
      )
    })

    expect(form.setError).toHaveBeenCalledTimes(1)
    expect(form.setError).toHaveBeenCalledWith('name', {
      message: 'Dieser Name ist bereits vergeben.',
    })
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('calls onSuccess on successful execution without setting errors or showing toast', async () => {
    const form = createFormMock()
    const onSuccess = vi.fn()
    const { result } = renderHook(() =>
      useFormActionSubmit({
        form,
        actionLabel: 'Speichern',
        onSuccess,
      }),
    )

    await act(async () => {
      await result.current.run(() => Promise.resolve())
    })

    expect(onSuccess).toHaveBeenCalledTimes(1)
    expect(form.setError).not.toHaveBeenCalled()
    expect(toast.error).not.toHaveBeenCalled()
  })
})
