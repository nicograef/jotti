import { useState } from 'react'
import type { FieldValues, Path, UseFormReturn } from 'react-hook-form'
import { toast } from 'sonner'

import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'

interface UseFormActionSubmitOptions<TFieldValues extends FieldValues> {
  form: UseFormReturn<TFieldValues>
  actionLabel: string
  byCode?: Record<string, string>
  onSuccess?: () => void
}

function toValidationDetails(
  details: unknown,
): Record<string, string[]> | undefined {
  if (
    typeof details !== 'object' ||
    details === null ||
    Array.isArray(details)
  ) {
    return undefined
  }

  const entries = Object.entries(details)

  if (
    !entries.every(
      ([, value]) =>
        Array.isArray(value) &&
        value.every((message) => typeof message === 'string'),
    )
  ) {
    return undefined
  }

  return details as Record<string, string[]>
}

function applyValidationErrors<TFieldValues extends FieldValues>(
  form: UseFormReturn<TFieldValues>,
  details: Record<string, string[]>,
): boolean {
  let applied = false

  for (const [field, messages] of Object.entries(details)) {
    const message = messages[0]
    if (!message) {
      continue
    }

    form.setError(field as Path<TFieldValues>, { message })
    applied = true
  }

  return applied
}

export function useFormActionSubmit<TFieldValues extends FieldValues>({
  form,
  actionLabel,
  byCode,
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

      if (error instanceof BackendError && error.code === 'validation_error') {
        const details = toValidationDetails(error.details)
        if (details && applyValidationErrors(form, details)) {
          return
        }
      }

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
