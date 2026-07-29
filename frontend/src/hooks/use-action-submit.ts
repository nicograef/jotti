import { useState } from 'react'
import { toast } from 'sonner'

import { getActionErrorMessage } from '@/lib/errorMessages'

import { useOffenerVorgang } from './use-offener-vorgang'

interface UseActionSubmitOptions {
  actionLabel: string
  byCode?: Record<string, string>
  onSuccess?: () => void
}

export function useActionSubmit({
  actionLabel,
  byCode,
  onSuccess,
}: UseActionSubmitOptions) {
  const [loading, setLoading] = useState(false)

  // Eine laufende Buchung ist ein offener Vorgang: Ein Reload mitten im Flug
  // ließe die Servicekraft ohne Antwort zurück. Das `finally` unten gibt ihn
  // nach Erfolg wie nach Fehlschlag wieder frei.
  useOffenerVorgang(loading)

  const run = async (fn: () => Promise<void>) => {
    setLoading(true)

    try {
      await fn()
      onSuccess?.()
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({
          actionLabel,
          error,
          byCode,
        }),
      )
    } finally {
      setLoading(false)
    }
  }

  return { loading, run }
}
