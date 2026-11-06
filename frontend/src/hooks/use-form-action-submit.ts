import type { FieldValues, Path, UseFormReturn } from 'react-hook-form'

import {
  type UseActionSubmitOptions,
  useActionSubmit,
} from './use-action-submit'

interface UseFormActionSubmitOptions<
  TFieldValues extends FieldValues,
> extends Omit<UseActionSubmitOptions, 'onCode' | 'onError'> {
  form: UseFormReturn<TFieldValues>
  fieldErrorsByCode?: Partial<
    Record<string, Partial<Record<Path<TFieldValues>, string>>>
  >
}

// useActionSubmit, das bekannte Fehlercodes als Feldfehler ins Formular
// schreibt statt als Toast.
export function useFormActionSubmit<TFieldValues extends FieldValues>({
  form,
  fieldErrorsByCode = {},
  ...optionen
}: UseFormActionSubmitOptions<TFieldValues>): {
  loading: boolean
  run: (fn: () => Promise<void>) => Promise<void>
} {
  const onCode: UseActionSubmitOptions['onCode'] = {}
  for (const [code, feldMeldungen] of Object.entries(fieldErrorsByCode)) {
    if (!feldMeldungen) continue
    onCode[code] = () => {
      for (const field of Object.keys(feldMeldungen) as Path<TFieldValues>[]) {
        const message = feldMeldungen[field]
        if (typeof message === 'string' && message.length > 0) {
          form.setError(field, { message })
        }
      }
    }
  }

  return useActionSubmit({ ...optionen, onCode })
}
