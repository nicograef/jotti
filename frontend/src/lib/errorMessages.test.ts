import { describe, expect, it } from 'vitest'

import { BackendError } from './Backend'
import { getActionErrorMessage } from './errorMessages'

const serverErrorMessage =
  'Es ist ein unerwarteter Serverfehler aufgetreten. Bitte Seite neu laden oder den Administrator kontaktieren.'

const mappedCodes: [string, string][] = [
  [
    'already_has_password',
    'Für diesen Benutzer wurde bereits ein Passwort gesetzt.',
  ],
  [
    'betreiber_nicht_konfiguriert',
    'Die Betreiberdaten sind unvollständig. Bitte im Bereich Finanzamt vervollständigen und erneut versuchen.',
  ],
  [
    'cannot_delete_self',
    'Der aktuell angemeldete Benutzer kann nicht gelöscht werden. Bitte einen anderen Benutzer wählen.',
  ],
  [
    'conflict',
    'Die Daten wurden gerade von jemand anderem geändert. Bitte aktualisieren und erneut versuchen.',
  ],
  [
    'invalid_json',
    'Die Anfrage konnte nicht verarbeitet werden. Bitte Eingaben prüfen und erneut versuchen.',
  ],
  [
    'invalid_kassensitzung_nr',
    'Die Kassensitzung konnte nicht gefunden werden. Bitte neu auswählen und erneut versuchen.',
  ],
  [
    'invalid_produkt_data',
    'Die Produktdaten sind ungültig. Bitte alle Felder prüfen.',
  ],
  [
    'invalid_tisch_data',
    'Die Tischdaten sind ungültig. Bitte Eingaben prüfen.',
  ],
  [
    'invalid_variante_data',
    'Die Variantendaten sind ungültig. Bitte Eingaben prüfen.',
  ],
  [
    'insufficient_permissions',
    'Dafür fehlen die nötigen Berechtigungen. Bitte einen Administrator kontaktieren.',
  ],
  [
    'kasse_bereits_geoeffnet',
    'Es gibt bereits eine offene Kassensitzung. Bitte zuerst die aktuelle Kassensitzung abschließen.',
  ],
  [
    'kasse_nicht_geoeffnet',
    'Die Kasse ist noch nicht geöffnet. Bitte zuerst eine Kassensitzung eröffnen.',
  ],
  [
    'kasse_wird_abgeschlossen',
    'Die Kasse wird gerade abgeschlossen. Bitte warten, bis der Abschluss fertig ist, und dann erneut versuchen.',
  ],
  [
    'kassenbeleg_drucker_nicht_konfiguriert',
    'Für Kassenbelege ist kein Drucker konfiguriert. Bitte die Druckstation-Einstellungen prüfen.',
  ],
  [
    'kassensturz_erforderlich',
    'Vor dem Tagesabschluss muss ein Kassensturz durchgeführt werden.',
  ],
  [
    'no_password_set',
    'Für diesen Benutzer wurde noch kein Passwort gesetzt. Bitte zuerst ein Passwort vergeben.',
  ],
  [
    'password_too_weak',
    'Das Passwort ist zu schwach. Bitte ein stärkeres Passwort verwenden.',
  ],
  [
    'position_nicht_ausgebbar',
    'Mindestens eine Position kann nicht ausgegeben werden. Bitte Tischstatus aktualisieren und erneut versuchen.',
  ],
  [
    'position_nicht_bezahlbar',
    'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Tischstatus aktualisieren und erneut versuchen.',
  ],
  [
    'position_nicht_stornierbar',
    'Mindestens eine Position kann nicht storniert werden. Bitte Tischstatus aktualisieren und erneut versuchen.',
  ],
  [
    'position_nicht_umbuchbar',
    'Mindestens eine Position kann nicht umgebucht werden. Bitte Tischstatus aktualisieren und erneut versuchen.',
  ],
  [
    'produkt_already_exists',
    'Ein Produkt mit diesem Namen existiert bereits. Bitte einen anderen Namen verwenden.',
  ],
  [
    'produkt_not_found',
    'Das Produkt wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  ],
  [
    'request_too_large',
    'Die Anfrage ist zu groß. Bitte weniger Daten auf einmal senden und erneut versuchen.',
  ],
  [
    'tisch_already_exists',
    'Ein Tisch mit diesem Namen existiert bereits. Bitte einen anderen Namen verwenden.',
  ],
  [
    'tisch_not_found',
    'Der Tisch wurde nicht gefunden. Bitte zur Tischübersicht zurückkehren und neu öffnen.',
  ],
  ['tisch_not_active', 'Dieser Tisch ist aktuell nicht aktiv.'],
  [
    'tische_saldo_offen',
    'Es gibt noch offene Tische mit ausstehenden Beträgen. Bitte alle Tische abrechnen.',
  ],
  [
    'tse_nicht_konfiguriert',
    'Die TSE ist nicht konfiguriert. Bitte sie im Bereich Finanzamt einrichten.',
  ],
  [
    'tse_verbindung_fehlgeschlagen',
    'Die Verbindung zur TSE ist fehlgeschlagen. Bitte Verbindung und TSE-Konfiguration prüfen.',
  ],
  [
    'umbuchung_gleicher_tisch',
    'Umbuchung auf denselben Tisch ist nicht möglich. Bitte einen anderen Ziel-Tisch wählen.',
  ],
  [
    'user_inactive',
    'Dieser Benutzer ist inaktiv. Bitte einen Administrator kontaktieren.',
  ],
  [
    'user_not_found',
    'Der Benutzer wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  ],
  [
    'username_already_exists',
    'Dieser Benutzername ist bereits vergeben. Bitte einen anderen Benutzernamen wählen.',
  ],
  ['validation_error', 'Bitte die Eingaben prüfen und erneut versuchen.'],
  [
    'variante_not_found',
    'Die Variante wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  ],
  [
    'verkauf_not_found',
    'Der Verkauf wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  ],
  [
    'zahlung_not_found',
    'Die Zahlung wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  ],
  ['unknown', serverErrorMessage],
]

describe('getActionErrorMessage', () => {
  it.each(mappedCodes)(
    'returns mapped message for code %s',
    (code, expected) => {
      expect(
        getActionErrorMessage({
          actionLabel: 'Aktion',
          error: new BackendError(400, code),
        }),
      ).toBe(expected)
    },
  )

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
})
