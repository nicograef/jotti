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
import { Steuersatz } from '@/lib/produktSchemas'

import { STEUERSATZ_LABEL } from './Produkt'

export function SteuersatzField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ steuersatz: Steuersatz } & AllFormFields>) {
  const id = useId()
  return (
    <Controller
      name={'steuersatz' as Path<{ steuersatz: Steuersatz } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor={id}>Steuersatz</FieldLabel>}
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
          >
            <SelectTrigger id={id} aria-invalid={fieldState.invalid}>
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              {Object.values(Steuersatz).map((steuersatz) => (
                <SelectItem key={steuersatz} value={steuersatz}>
                  {STEUERSATZ_LABEL[steuersatz]}
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
