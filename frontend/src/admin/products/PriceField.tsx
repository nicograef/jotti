import type { FieldPath, FieldValues } from 'react-hook-form'

import { EuroField, type FieldProps } from '@/components/common/FormFields'

/** Euro input for a product's gross price (Brutto), bound to the `preisCents` form field. */
export function PriceField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ preisCents: number } & AllFormFields>) {
  return (
    <EuroField
      form={form}
      name={'preisCents' as FieldPath<{ preisCents: number } & AllFormFields>}
      label="Preis"
      placeholder={placeholder ?? 'Preis in Euro (z.B. 4,50)'}
    />
  )
}
