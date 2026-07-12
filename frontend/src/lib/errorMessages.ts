import { BackendError } from './Backend'

const serverErrorMessage =
  'Es ist ein unerwarteter Serverfehler aufgetreten. Bitte Seite neu laden oder den Administrator kontaktieren.'

function appendReferenz(message: string, referenz?: string): string {
  return referenz ? `${message} Referenz: ${referenz}` : message
}

const commonErrorMessages: Record<string, string> = {
  onetime_password_locked:
    'Das Einmalpasswort wurde nach zu vielen Fehlversuchen gesperrt. Bitte einen Admin um ein neues Einmalpasswort.',
  already_has_password:
    'Für diesen Benutzer wurde bereits ein Passwort gesetzt.',
  betreiber_nicht_konfiguriert:
    'Die Betreiberdaten sind unvollständig. Bitte im Bereich Finanzamt vervollständigen und erneut versuchen.',
  buchungen_nach_kassensturz:
    'Nach dem Kassensturz wurden noch Buchungen erfasst. Der Abschluss kann so nicht wiederholt werden. Bitte den Administrator kontaktieren.',
  cannot_delete_self:
    'Der aktuell angemeldete Benutzer kann nicht gelöscht werden. Bitte einen anderen Benutzer wählen.',
  druckstation_nicht_konfiguriert:
    'Für diese Station ist kein Drucker konfiguriert. Bitte zuerst eine Drucker-IP eintragen, dann den Testbon senden.',
  conflict:
    'Die Daten wurden gerade von jemand anderem geändert. Bitte aktualisieren und erneut versuchen.',
  invalid_json:
    'Die Anfrage konnte nicht verarbeitet werden. Bitte Eingaben prüfen und erneut versuchen.',
  invalid_kassensitzung_nr:
    'Die Kassensitzung konnte nicht gefunden werden. Bitte neu auswählen und erneut versuchen.',
  invalid_produkt_data:
    'Die Produktdaten sind ungültig. Bitte alle Felder prüfen.',
  invalid_tisch_data: 'Die Tischdaten sind ungültig. Bitte Eingaben prüfen.',
  invalid_variante_data:
    'Die Variantendaten sind ungültig. Bitte Eingaben prüfen.',
  insufficient_permissions:
    'Dafür fehlen die nötigen Berechtigungen. Bitte einen Administrator kontaktieren.',
  internal_server_error: serverErrorMessage,
  tisch_not_found:
    'Der Tisch wurde nicht gefunden. Bitte zur Tischübersicht zurückkehren und neu öffnen.',
  tisch_not_active: 'Dieser Tisch ist aktuell nicht aktiv.',
  tisch_already_exists:
    'Ein Tisch mit diesem Namen existiert bereits. Bitte einen anderen Namen verwenden.',
  kasse_nicht_geoeffnet:
    'Die Kasse ist noch nicht geöffnet. Bitte zuerst eine Kassensitzung eröffnen.',
  kasse_wird_abgeschlossen:
    'Die Kasse wird gerade abgeschlossen. Bitte warten, bis der Abschluss fertig ist, und dann erneut versuchen.',
  kasse_bereits_geoeffnet:
    'Es gibt bereits eine offene Kassensitzung. Bitte zuerst die aktuelle Kassensitzung abschließen.',
  kassensturz_erforderlich:
    'Vor dem Tagesabschluss muss ein Kassensturz durchgeführt werden.',
  kassenbeleg_drucker_nicht_konfiguriert:
    'Für Kassenbelege ist kein Drucker konfiguriert. Bitte die Druckstation-Einstellungen prüfen.',
  no_password_set:
    'Für diesen Benutzer wurde noch kein Passwort gesetzt. Bitte zuerst ein Passwort vergeben.',
  password_too_weak:
    'Das Passwort ist zu schwach. Bitte ein stärkeres Passwort verwenden.',
  position_nicht_bezahlbar:
    'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Tischstatus aktualisieren und erneut versuchen.',
  position_nicht_stornierbar:
    'Mindestens eine Position kann nicht storniert werden. Bitte Tischstatus aktualisieren und erneut versuchen.',
  position_nicht_umbuchbar:
    'Mindestens eine Position kann nicht umgebucht werden. Bitte Tischstatus aktualisieren und erneut versuchen.',
  produkt_already_exists:
    'Ein Produkt mit diesem Namen existiert bereits. Bitte einen anderen Namen verwenden.',
  produkt_hat_verkaeufe:
    'Produkte mit Verkäufen können nur deaktiviert werden, damit die Berichte vollständig bleiben.',
  produkt_not_found:
    'Das Produkt wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  request_too_large:
    'Die Anfrage ist zu groß. Bitte weniger Daten auf einmal senden und erneut versuchen.',
  tisch_saldo_offen:
    'Dieser Tisch hat noch einen offenen Saldo. Bitte zuerst abrechnen, dann lässt er sich deaktivieren oder löschen.',
  tische_saldo_offen:
    'Es gibt noch offene Tische mit ausstehenden Beträgen. Bitte alle Tische abrechnen.',
  tse_nicht_konfiguriert:
    'Die TSE ist nicht konfiguriert. Bitte sie im Bereich Finanzamt einrichten.',
  tse_verbindung_fehlgeschlagen:
    'Die Verbindung zur TSE ist fehlgeschlagen. Bitte Verbindung und TSE-Konfiguration prüfen.',
  umbuchung_gleicher_tisch:
    'Umbuchung auf denselben Tisch ist nicht möglich. Bitte einen anderen Ziel-Tisch wählen.',
  user_inactive:
    'Dieser Benutzer ist inaktiv. Bitte einen Administrator kontaktieren.',
  user_not_found:
    'Der Benutzer wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  username_already_exists:
    'Dieser Benutzername ist bereits vergeben. Bitte einen anderen Benutzernamen wählen.',
  validation_error: 'Bitte die Eingaben prüfen und erneut versuchen.',
  variante_not_found:
    'Die Variante wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  verkauf_not_found:
    'Der Verkauf wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  zahlung_not_found:
    'Die Zahlung wurde nicht gefunden. Bitte neu laden und erneut versuchen.',
  unknown: serverErrorMessage,
}

interface ErrorMessageOptions {
  actionLabel: string
  error: unknown
  byCode?: Record<string, string>
}

export function getActionErrorMessage({
  actionLabel,
  error,
  byCode = {},
}: ErrorMessageOptions): string {
  const fallback = `${actionLabel} fehlgeschlagen. Bitte erneut versuchen.`

  if (error instanceof BackendError) {
    if (Object.prototype.hasOwnProperty.call(byCode, error.code)) {
      return byCode[error.code]
    }

    if (
      error.status >= 500 ||
      error.code === 'internal_server_error' ||
      error.code === 'unknown'
    ) {
      return appendReferenz(serverErrorMessage, error.referenz)
    }

    if (Object.prototype.hasOwnProperty.call(commonErrorMessages, error.code)) {
      return commonErrorMessages[error.code]
    }

    return fallback
  }

  return fallback
}
