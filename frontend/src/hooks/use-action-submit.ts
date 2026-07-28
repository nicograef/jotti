import { useState } from 'react'
import { toast } from 'sonner'

import { BackendError, type NetzwerkFehlerArt } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'

interface UseActionSubmitOptions {
  actionLabel: string
  byCode?: Record<string, string>
  byNetzwerkArt?: Partial<Record<NetzwerkFehlerArt, string>>
  // Seiteneffekt je Fehlercode, parallel zu `byCode` (Meldung je Fehlercode);
  // läuft vor dem Fehler-Toast. Nötig für Fehler, die den Vorgang beenden statt
  // ihn wiederholbar zu lassen: `vorgang_daten_abweichend` etwa belegt, dass der
  // Vorgang unter diesem Schlüssel gebucht ist — die Aufrufstelle räumt daraufhin
  // ihre Zusammenstellung ab, damit der nächste Vorgang einen neuen Schlüssel
  // bekommt und nicht wieder in denselben Fehler läuft.
  onCode?: Record<string, () => void>
  onSuccess?: () => void
}

export function useActionSubmit({
  actionLabel,
  byCode,
  byNetzwerkArt,
  onCode,
  onSuccess,
}: UseActionSubmitOptions) {
  const [loading, setLoading] = useState(false)

  const run = async (fn: () => Promise<void>) => {
    setLoading(true)

    try {
      await fn()
      onSuccess?.()
    } catch (error: unknown) {
      console.error(error)

      // hasOwnProperty wie in getActionErrorMessage: Ein Fehlercode wie
      // „constructor" darf nichts vom Prototyp treffen.
      if (
        error instanceof BackendError &&
        onCode &&
        Object.prototype.hasOwnProperty.call(onCode, error.code)
      ) {
        onCode[error.code]()
      }

      toast.error(
        getActionErrorMessage({
          actionLabel,
          error,
          byCode,
          byNetzwerkArt,
        }),
      )
    } finally {
      setLoading(false)
    }
  }

  return { loading, run }
}
