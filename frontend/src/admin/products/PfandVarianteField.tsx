import { useId } from 'react'
import { Controller, type FieldValues, type Path } from 'react-hook-form'

import type { FieldProps } from '@/components/common/FormFields'
import { Field, FieldError, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import type { PfandOption } from './pfandVerknuepfung'

// Radix Select verbietet den leeren Wert; „Kein Pfand" braucht einen eigenen.
const KEIN_PFAND = 'kein'

interface PfandForm {
  pfandVarianteId: number | null
}

export function PfandVarianteField<AllFormFields extends FieldValues>({
  form,
  optionen,
}: Pick<FieldProps<PfandForm & AllFormFields>, 'form'> & {
  optionen: PfandOption[]
}) {
  const id = useId()
  return (
    <Controller
      name={'pfandVarianteId' as Path<PfandForm & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <FieldLabel htmlFor={id}>Pfand</FieldLabel>
          <Select
            name={field.name}
            value={field.value === null ? KEIN_PFAND : String(field.value)}
            onValueChange={(wert) => {
              field.onChange(wert === KEIN_PFAND ? null : Number(wert))
            }}
          >
            <SelectTrigger id={id} aria-invalid={fieldState.invalid}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={KEIN_PFAND}>Kein Pfand</SelectItem>
              {optionen.map((option) => (
                <SelectItem
                  key={option.varianteId}
                  value={String(option.varianteId)}
                >
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
