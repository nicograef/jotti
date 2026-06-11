import { useState } from 'react'
import type { FieldValues, Path, UseFormReturn } from 'react-hook-form'
import { toast } from 'sonner'

import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'

interface UseFormActionSubmitOptions<TFieldValues extends FieldValues> {
  form: UseFormReturn<TFieldValues>
  actionLabel: string
  byCode?: Record<string, string>
  fieldErrorsByCode?: Partial<
    Record<string, Partial<Record<Path<TFieldValues>, string>>>
  >
  onSuccess?: () => void
}

export function useFormActionSubmit<TFieldValues extends FieldValues>({
  form,
  actionLabel,
  byCode,
  fieldErrorsByCode,
  onSuccess,
}: UseFormActionSubmitOptions<TFieldValues>) {
  const [loading, setLoading] = useState(false)

  const run = async (fn: () => Promise<void>) => {
    setLoading(true)

    try {
      await fn()
      onSuccess?.()
    } catch (error: unknown) {
      console.error(error)

      if (error instanceof BackendError && fieldErrorsByCode) {
        const fieldMessages = fieldErrorsByCode[error.code]
        if (fieldMessages) {
          for (const field of Object.keys(
            fieldMessages,
          ) as Path<TFieldValues>[]) {
            const message = fieldMessages[field]
            if (typeof message === 'string' && message.length > 0) {
              form.setError(field, { message })
            }
          }
          return
        }
      }

      toast.error(getActionErrorMessage({ actionLabel, error, byCode }))
    } finally {
      setLoading(false)
    }
  }

  return { loading, run }
}
