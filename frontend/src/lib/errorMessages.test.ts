import { describe, expect, it } from 'vitest'

import { BackendError } from './Backend'
import { commonErrorMessages, getActionErrorMessage } from './errorMessages'

const serverErrorMessage =
  'Es ist ein unerwarteter Serverfehler aufgetreten. Bitte Seite neu laden oder den Administrator kontaktieren.'

describe('getActionErrorMessage', () => {
  it.each(Object.entries(commonErrorMessages))(
    'returns the central message for code %s',
    (code, expected) => {
      expect(
        getActionErrorMessage({
          actionLabel: 'Aktion',
          error: new BackendError(400, code),
        }),
      ).toBe(expected)
    },
  )

  it('returns the server error message for code unknown', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(400, 'unknown'),
      }),
    ).toBe(serverErrorMessage)
  })

  it('returns server error message for 5xx backend errors', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(502, 'gateway_timeout'),
      }),
    ).toBe(serverErrorMessage)
  })

  it('returns server error message for internal_server_error regardless of status', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(400, 'internal_server_error'),
      }),
    ).toBe(serverErrorMessage)
  })

  it('prioritizes byCode overrides over common code messages', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Speichern',
        error: new BackendError(400, 'produkt_already_exists'),
        byCode: {
          produkt_already_exists: 'Bitte anderen Produktnamen wählen.',
        },
      }),
    ).toBe('Bitte anderen Produktnamen wählen.')
  })

  it('returns action fallback for unknown backend 4xx code', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Speichern',
        error: new BackendError(400, 'something_new'),
      }),
    ).toBe('Speichern fehlgeschlagen. Bitte erneut versuchen.')
  })

  it('returns action fallback for non-backend errors', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Laden',
        error: new Error('boom'),
      }),
    ).toBe('Laden fehlgeschlagen. Bitte erneut versuchen.')
  })

  it('appends the correlation ID reference for 5xx errors with a referenz', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(
          500,
          'internal_server_error',
          undefined,
          'a1b2c3d4',
        ),
      }),
    ).toBe(`${serverErrorMessage} Referenz: a1b2c3d4`)
  })

  it('returns the plain server error message for 5xx errors without a referenz', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(500, 'internal_server_error'),
      }),
    ).toBe(serverErrorMessage)
  })

  it('appends the correlation ID reference for code unknown with a referenz', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Kassieren',
        error: new BackendError(400, 'unknown', undefined, 'a1b2c3d4'),
      }),
    ).toBe(`${serverErrorMessage} Referenz: a1b2c3d4`)
  })

  it('does not append a referenz for non-server errors even if one is set', () => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Speichern',
        error: new BackendError(
          400,
          'produkt_not_found',
          undefined,
          'a1b2c3d4',
        ),
      }),
    ).toBe(
      'Das Produkt wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
    )
  })
})
