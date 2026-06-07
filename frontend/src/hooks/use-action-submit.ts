import { useState } from 'react'
import { toast } from 'sonner'

import { getActionErrorMessage } from '@/lib/errorMessages'

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
