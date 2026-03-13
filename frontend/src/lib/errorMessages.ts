import { BackendError } from './Backend'

const commonErrorMessages: Record<string, string> = {
  conflict:
    'Die Daten wurden gerade von jemand anderem geändert. Bitte aktualisieren und erneut versuchen.',
  tisch_not_found:
    'Der Tisch wurde nicht gefunden. Bitte zur Tischübersicht zurückkehren und neu öffnen.',
  tisch_not_active: 'Dieser Tisch ist aktuell nicht aktiv.',
  unknown: 'Es ist ein unerwarteter Fehler aufgetreten.',
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

    if (Object.prototype.hasOwnProperty.call(commonErrorMessages, error.code)) {
      return commonErrorMessages[error.code]
    }

    return fallback
  }

  return fallback
}
