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
import {
  type Kategorie,
  KATEGORIE_LABEL,
  KATEGORIE_ORDER,
} from '@/lib/produktSchemas'

export function CategoryField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ kategorie: Kategorie } & AllFormFields>) {
  const id = useId()
  return (
    <Controller
      name={'kategorie' as Path<{ kategorie: Kategorie } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <FieldLabel htmlFor={id}>Kategorie</FieldLabel>
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
          >
            <SelectTrigger id={id} aria-invalid={fieldState.invalid}>
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              {KATEGORIE_ORDER.map((kategorie) => (
                <SelectItem key={kategorie} value={kategorie}>
                  {KATEGORIE_LABEL[kategorie]}
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
