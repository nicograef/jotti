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

function applyMappedCodeError<TFieldValues extends FieldValues>(
  form: UseFormReturn<TFieldValues>,
  error: unknown,
  byCode?: Record<string, string>,
): boolean {
  if (!(error instanceof BackendError) || !byCode) {
    return false
  }

  const message = byCode[error.code]
  if (!message) {
    return false
  }

  const field =
    error.code === 'username_already_exists'
      ? 'username'
      : error.code.endsWith('_already_exists')
        ? 'name'
        : undefined

  if (!field) {
    return false
  }

  form.setError(field as Path<TFieldValues>, { message })
  return true
}

function applyExplicitCodeErrors<TFieldValues extends FieldValues>(
  form: UseFormReturn<TFieldValues>,
  error: unknown,
  fieldErrorsByCode?: Partial<
    Record<string, Partial<Record<Path<TFieldValues>, string>>>
  >,
): boolean {
  if (!(error instanceof BackendError) || !fieldErrorsByCode) {
    return false
  }

  const fieldMessages = fieldErrorsByCode[error.code]
  if (!fieldMessages) {
    return false
  }

  let applied = false
  for (const field of Object.keys(fieldMessages) as Path<TFieldValues>[]) {
    const message = fieldMessages[field]
    if (typeof message !== 'string' || message.length === 0) {
      continue
    }

    form.setError(field, { message })
    applied = true
  }

  return applied
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

      if (error instanceof BackendError && error.code === 'validation_error') {
        const details = toValidationDetails(error.details)
        if (details && applyValidationErrors(form, details)) {
          return
        }
      }

      if (applyExplicitCodeErrors(form, error, fieldErrorsByCode)) {
        return
      }

      if (applyMappedCodeError(form, error, byCode)) {
        return
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
