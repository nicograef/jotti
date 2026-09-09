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

import type { UserRole } from './User'

// Rollen-Auswahl der Benutzer-Formulare. Liegt bei der Benutzerverwaltung, weil
// nur sie die Rollen und ihre Beschreibungen kennt.
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
          {field.value === 'admin' && (
            <FieldDescription>
              Administratoren können alle Funktionen nutzen.
            </FieldDescription>
          )}
          {field.value === 'serviceleitung' && (
            <FieldDescription>
              Serviceleitung kann bestellen, kassieren und stornieren.
            </FieldDescription>
          )}
          {field.value === 'service' && (
            <FieldDescription>
              Servicekräfte können bestellen, liefern und kassieren.
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
