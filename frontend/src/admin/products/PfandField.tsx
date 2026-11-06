import { useId } from 'react'
import { Controller, type FieldValues, type Path } from 'react-hook-form'

import type { FieldProps } from '@/components/common/FormFields'
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldLabel,
} from '@/components/ui/field'
import { Switch } from '@/components/ui/switch'

export function PfandField<AllFormFields extends FieldValues>({
  form,
}: Pick<FieldProps<{ pfand: boolean } & AllFormFields>, 'form'>) {
  const id = useId()
  return (
    <Controller
      name={'pfand' as Path<{ pfand: boolean } & AllFormFields>}
      control={form.control}
      render={({ field }) => (
        <Field orientation="horizontal">
          <FieldContent>
            <FieldLabel htmlFor={id}>Pfand-Produkt</FieldLabel>
            <FieldDescription>
              Seine Varianten erscheinen bei der Pfandrückgabe.
            </FieldDescription>
          </FieldContent>
          <Switch
            id={id}
            name={field.name}
            checked={field.value}
            onCheckedChange={field.onChange}
          />
        </Field>
      )}
    />
  )
}
