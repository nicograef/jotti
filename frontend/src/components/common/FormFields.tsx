import { REGEXP_ONLY_DIGITS } from 'input-otp'
import {
  Controller,
  type FieldValues,
  type Path,
  type UseFormReturn,
} from 'react-hook-form'

import type { ProductCategory } from '@/admin/products/ProductBackend'
import { toUsername, UserRole } from '@/admin/users/UserBackend'
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

import { Textarea } from '../ui/textarea'

interface FieldProps<TField extends FieldValues> {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<TField, any, TField>
  withLabel?: boolean
  placeholder?: string
  description?: string
}

export function NameField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ name: string } & AllFormFields>) {
  return (
    <Controller
      name={'name' as Path<{ name: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor="form-name">Name</FieldLabel>}
          <Input
            {...field}
            id="form-name"
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Vor- und Nachname eingeben'}
            autoComplete="off"
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function UsernameField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ username: string } & AllFormFields>) {
  return (
    <Controller
      name={'username' as Path<{ username: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-username">Benutzername</FieldLabel>
          )}
          <Input
            {...field}
            onChange={(e) => {
              const username = toUsername(e.target.value)
              field.onChange(username)
            }}
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Benutzername'}
            autoComplete="off"
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function PasswordField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ password: string } & AllFormFields>) {
  return (
    <Controller
      name={'password' as Path<{ password: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <Input
            {...field}
            type="password"
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Passwort'}
            autoComplete="current-password"
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function NewPasswordField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ password: string } & AllFormFields>) {
  return (
    <Controller
      name={'password' as Path<{ password: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <Input
            {...field}
            type="password"
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Neues Passwort'}
            autoComplete="off"
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function OTPField<AllFormFields extends FieldValues>({
  form,
}: FieldProps<{ onetimePassword: string } & AllFormFields>) {
  return (
    <Controller
      name={
        'onetimePassword' as Path<{ onetimePassword: string } & AllFormFields>
      }
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <InputOTP
            maxLength={6}
            aria-invalid={fieldState.invalid}
            pattern={REGEXP_ONLY_DIGITS}
            {...field}
          >
            <InputOTPGroup className="mx-auto">
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          <FieldDescription className="text-center">
            Gib deinen Code ein.
          </FieldDescription>
        </Field>
      )}
    />
  )
}

export function RoleField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ role: UserRole } & AllFormFields>) {
  return (
    <Controller
      name={'role' as Path<{ role: UserRole } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor="form-role">Rolle</FieldLabel>}
          {field.value === 'admin' && (
            <FieldDescription>
              Administratoren können alle Funktionen nutzen.
            </FieldDescription>
          )}
          {field.value === 'service' && (
            <FieldDescription>
              Service kann Bestellungen und Bezahlungen verwalten.
            </FieldDescription>
          )}
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
          >
            <SelectTrigger id="form-role" aria-invalid={fieldState.invalid}>
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="admin">Administrator</SelectItem>
              <SelectItem value="service">Service</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function LockedField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  description,
}: FieldProps<{ locked: boolean } & AllFormFields>) {
  return (
    <Controller
      name={'locked' as Path<{ locked: boolean } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor="form-locked">Sperren?</FieldLabel>}
          <FieldContent className="flex flex-row items-center">
            <Switch
              id="form-locked"
              aria-invalid={fieldState.invalid}
              checked={field.value}
              onCheckedChange={field.onChange}
            />
            {field.value && (
              <FieldDescription className="ml-4">
                {description ??
                  'Wenn du diesen Benutzer sperrst, kann er sich nicht mehr anmelden.'}
              </FieldDescription>
            )}
          </FieldContent>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function DescriptionField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ description: string } & AllFormFields>) {
  return (
    <Controller
      name={'description' as Path<{ description: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-description">Beschreibung</FieldLabel>
          )}
          <Textarea
            {...field}
            id="form-description"
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Beschreibung eingeben (optional)'}
            autoComplete="off"
            rows={3}
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function NetPriceField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ netPrice: number } & AllFormFields>) {
  return (
    <Controller
      name={'netPrice' as Path<{ netPrice: number } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-netPrice">Netto-Preis</FieldLabel>
          )}
          <Input
            {...field}
            id="form-netPrice"
            type="number"
            step="0.01"
            min="0"
            aria-invalid={fieldState.invalid}
            placeholder={placeholder ?? 'Preis eingeben (in Euro)'}
            autoComplete="off"
            onChange={(e) => {
              const value = parseFloat(e.target.value)
              field.onChange(isNaN(value) ? 0 : value)
            }}
          />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function CategoryField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ category: ProductCategory } & AllFormFields>) {
  return (
    <Controller
      name={'category' as Path<{ category: ProductCategory } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-category">Kategorie</FieldLabel>
          )}
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
          >
            <SelectTrigger id="form-category" aria-invalid={fieldState.invalid}>
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="food">Essen</SelectItem>
              <SelectItem value="beverage">Getränk</SelectItem>
              <SelectItem value="other">Sonstiges</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
