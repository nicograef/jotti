import { useState } from 'react'
import { toast } from 'sonner'

import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'

import { useOffenerVorgang } from './use-offener-vorgang'

export interface UseActionSubmitOptions {
  actionLabel: string
  byCode?: Record<string, string>
  // Fehlercodes mit eigener Anzeige statt Toast; onError läuft dann nicht.
  onCode?: Partial<Record<string, (error: BackendError) => void>>
  onSuccess?: () => void
  // Läuft nach der Fehlermeldung, etwa um eine abgewiesene Auswahl zu verwerfen.
  onError?: (error: unknown) => void
}

export function useActionSubmit({
  actionLabel,
  byCode,
  onCode,
  onSuccess,
  onError,
}: UseActionSubmitOptions): {
  loading: boolean
  run: (fn: () => Promise<void>) => Promise<void>
} {
  const [loading, setLoading] = useState(false)

  // Eine laufende Buchung ist ein offener Vorgang: Ein Reload mitten im Flug
  // ließe die Servicekraft ohne Antwort zurück.
  useOffenerVorgang(loading)

  const run = async (fn: () => Promise<void>) => {
    setLoading(true)

    try {
      await fn()
      onSuccess?.()
    } catch (error: unknown) {
      const eigeneAnzeige =
        error instanceof BackendError ? onCode?.[error.code] : undefined
      if (eigeneAnzeige && error instanceof BackendError) {
        eigeneAnzeige(error)
        return
      }
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel,
          error,
          byCode,
        }),
      )
      onError?.(error)
    } finally {
      setLoading(false)
    }
  }

  return { loading, run }
}
