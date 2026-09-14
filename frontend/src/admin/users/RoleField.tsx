import { useId } from 'react'
import { Controller, type FieldValues, type Path } from 'react-hook-form'

import type { FieldProps } from '@/components/common/FormFields'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import { UserRole } from './User'

const ROLLE_BESCHREIBUNG: Record<UserRole, string> = {
  [UserRole.ADMIN]: 'Administratoren können alle Funktionen nutzen.',
  [UserRole.SERVICELEITUNG]:
    'Serviceleitung kann bestellen, kassieren und stornieren.',
  [UserRole.SERVICE]: 'Servicekräfte können bestellen, liefern und kassieren.',
}

export function RoleField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
  disabled,
}: FieldProps<{ role: UserRole } & AllFormFields> & { disabled?: boolean }) {
  const id = useId()
  return (
    <Controller
      name={'role' as Path<{ role: UserRole } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor={id}>Rolle</FieldLabel>}
          {field.value && (
            <FieldDescription>
              {ROLLE_BESCHREIBUNG[field.value]}
            </FieldDescription>
          )}
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
            disabled={disabled}
          >
            <SelectTrigger id={id} aria-invalid={fieldState.invalid}>
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="admin">Administrator</SelectItem>
              <SelectItem value="serviceleitung">Serviceleitung</SelectItem>
              <SelectItem value="service">Service</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
