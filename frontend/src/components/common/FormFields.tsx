import { REGEXP_ONLY_DIGITS } from 'input-otp'
import { EyeClosedIcon, EyeIcon } from 'lucide-react'
import { useState } from 'react'
import {
  Controller,
  type FieldValues,
  type Path,
  type UseFormReturn,
} from 'react-hook-form'

import type { Kategorie, Steuersatz } from '@/admin/products/Produkt'
import { toUsername, UserRole } from '@/admin/users/User'
import { Button } from '@/components/ui/button'
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
import { Textarea } from '@/components/ui/textarea'
import { formatCents, parseCents } from '@/lib/utils'

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
  const [visible, setVisible] = useState(false)

  return (
    <Controller
      name={'password' as Path<{ password: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <div className="flex">
            <Input
              {...field}
              type={visible ? 'text' : 'password'}
              aria-invalid={fieldState.invalid}
              placeholder={placeholder ?? 'Passwort'}
              autoComplete="current-password"
              className="rounded-r-none"
            />
            <Button
              variant="outline"
              size="icon"
              aria-label="Passwort anzeigen"
              className="rounded-l-none"
              type="button"
              onClick={() => {
                setVisible(!visible)
              }}
            >
              {visible ? <EyeIcon /> : <EyeClosedIcon />}
            </Button>
          </div>
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
  const [visible, setVisible] = useState(false)

  return (
    <Controller
      name={'password' as Path<{ password: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <div className="flex">
            <Input
              {...field}
              type={visible ? 'text' : 'password'}
              aria-invalid={fieldState.invalid}
              placeholder={placeholder ?? 'Neues Passwort'}
              autoComplete="off"
              className="rounded-r-none"
            />
            <Button
              variant="outline"
              size="icon"
              aria-label="Passwort anzeigen"
              className="rounded-l-none"
              type="button"
              onClick={() => {
                setVisible(!visible)
              }}
            >
              {visible ? <EyeIcon /> : <EyeClosedIcon />}
            </Button>
          </div>
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
          >
            <SelectTrigger id="form-role" aria-invalid={fieldState.invalid}>
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

const cleanInput = (input: string): string => {
  return input.replace(/[^0-9,]/g, '').replace(/,+/g, ',')
}

/** Input field for the net price. Converts the data representation as cents to the user-friendly Euro format. */
export function PriceField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ preisCents: number } & AllFormFields>) {
  const [value, setValue] = useState<string>(() => {
    const cents = form.getValues().preisCents
    return cents > 0 ? formatCents(cents) : ''
  })
  const [debounceTimeout, setDebounceTimeout] = useState<number | null>(null)

  return (
    <Controller
      name={'preisCents' as Path<{ preisCents: number } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-preisCents">Preis</FieldLabel>
          )}
          <div className="flex">
            <span className="inline-flex items-center px-3 rounded-l-md border border-r-0 border-input bg-background text-muted-foreground text-sm">
              €
            </span>
            <Input
              {...field}
              id="form-preisCents"
              className="border-l-0"
              type="text"
              aria-invalid={fieldState.invalid}
              placeholder={placeholder ?? 'Preis in Euro (z.B. 4,50)'}
              autoComplete="off"
              value={value}
              onChange={(e) => {
                const cleanedValue = cleanInput(e.target.value)
                setValue(cleanedValue)
                const preisCents = parseCents(cleanedValue)
                field.onChange(preisCents)

                //  Debounce the conversion to avoid cursor jumping
                if (debounceTimeout) clearTimeout(debounceTimeout)
                const newTimeout = setTimeout(() => {
                  setValue(preisCents > 0 ? formatCents(preisCents) : '')
                }, 1000)
                setDebounceTimeout(newTimeout)
              }}
              onBlur={(e) => {
                if (debounceTimeout) clearTimeout(debounceTimeout)
                const preisCents = parseCents(e.target.value)
                setValue(preisCents > 0 ? formatCents(preisCents) : '')
                field.onChange(preisCents)
              }}
            />
          </div>
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
}: FieldProps<{ kategorie: Kategorie } & AllFormFields>) {
  return (
    <Controller
      name={'kategorie' as Path<{ kategorie: Kategorie } & AllFormFields>}
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
              <SelectItem value="essen">Essen</SelectItem>
              <SelectItem value="getraenk">Getränk</SelectItem>
              <SelectItem value="sonstiges">Sonstiges</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}

export function SteuersatzField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ steuersatz: Steuersatz } & AllFormFields>) {
  return (
    <Controller
      name={'steuersatz' as Path<{ steuersatz: Steuersatz } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && (
            <FieldLabel htmlFor="form-steuersatz">Steuersatz</FieldLabel>
          )}
          <Select
            name={field.name}
            value={field.value}
            onValueChange={field.onChange}
          >
            <SelectTrigger
              id="form-steuersatz"
              aria-invalid={fieldState.invalid}
            >
              <SelectValue placeholder={placeholder ?? 'Auswählen'} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="regel">Regelsteuersatz (19 %)</SelectItem>
              <SelectItem value="ermaessigt">Ermäßigter Satz (7 %)</SelectItem>
              <SelectItem value="befreit">Steuerbefreit (0 %)</SelectItem>
              <SelectItem value="kombi">Kombi (70/30)</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
