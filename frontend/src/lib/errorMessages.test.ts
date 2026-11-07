import { describe, expect, it } from 'vitest'

import { BackendError } from './Backend'
import { getActionErrorMessage } from './errorMessages'

const serverErrorMessage =
  'Es ist ein unerwarteter Serverfehler aufgetreten. Bitte Seite neu laden oder den Administrator kontaktieren.'

describe('getActionErrorMessage', () => {
  // Literal rows pin the wording of messages a volunteer must act on.
  it.each([
    [
      'position_nicht_bezahlbar',
      'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Tischstatus aktualisieren und erneut versuchen.',
    ],
    ['validation_error', 'Bitte die Eingaben prüfen und erneut versuchen.'],
    [
      'pfandrueckgabe_abweichend',
      'Diese Pfandrückgabe wurde mit anderer Auswahl schon gebucht. Bitte Historie prüfen und neu erfassen.',
    ],
  ])('returns the central message for code %s', (code, expected) => {
    expect(
      getActionErrorMessage({
        actionLabel: 'Aktion',
        error: new BackendError(400, code),
      }),
    ).toBe(expected)
  })

  it.each<[string, Parameters<typeof getActionErrorMessage>[0], string]>([
    [
      'the server message for code unknown',
      { actionLabel: 'Kassieren', error: new BackendError(400, 'unknown') },
      serverErrorMessage,
    ],
    [
      'the server message for a 5xx',
      {
        actionLabel: 'Kassieren',
        error: new BackendError(502, 'gateway_timeout'),
      },
      serverErrorMessage,
    ],
    [
      'the server message for internal_server_error regardless of status',
      {
        actionLabel: 'Kassieren',
        error: new BackendError(400, 'internal_server_error'),
      },
      serverErrorMessage,
    ],
    [
      'a byCode override before the common message',
      {
        actionLabel: 'Speichern',
        error: new BackendError(400, 'produkt_already_exists'),
        byCode: {
          produkt_already_exists: 'Bitte anderen Produktnamen wählen.',
        },
      },
      'Bitte anderen Produktnamen wählen.',
    ],
    [
      'the action fallback for an unknown 4xx code',
      {
        actionLabel: 'Speichern',
        error: new BackendError(400, 'something_new'),
      },
      'Speichern fehlgeschlagen. Bitte erneut versuchen.',
    ],
    [
      'the action fallback for a non-backend error',
      { actionLabel: 'Laden', error: new Error('boom') },
      'Laden fehlgeschlagen. Bitte erneut versuchen.',
    ],
    [
      'the referenz after the server message for a 5xx',
      {
        actionLabel: 'Kassieren',
        error: new BackendError(
          500,
          'internal_server_error',
          undefined,
          'a1b2c3d4',
        ),
      },
      `${serverErrorMessage} Referenz: a1b2c3d4`,
    ],
    [
      'the referenz after the server message for code unknown',
      {
        actionLabel: 'Kassieren',
        error: new BackendError(400, 'unknown', undefined, 'a1b2c3d4'),
      },
      `${serverErrorMessage} Referenz: a1b2c3d4`,
    ],
    [
      'no referenz for a non-server error',
      {
        actionLabel: 'Speichern',
        error: new BackendError(
          400,
          'produkt_not_found',
          undefined,
          'a1b2c3d4',
        ),
      },
      'Das Produkt wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
    ],
  ])('returns %s', (_, options, expected) => {
    expect(getActionErrorMessage(options)).toBe(expected)
  })
})
