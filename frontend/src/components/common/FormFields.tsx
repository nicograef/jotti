import { REGEXP_ONLY_DIGITS } from 'input-otp'
import { EyeClosedIcon, EyeIcon } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import {
  Controller,
  type ControllerFieldState,
  type ControllerRenderProps,
  type FieldPath,
  type FieldValues,
  type Path,
  type UseFormReturn,
} from 'react-hook-form'

import {
  type Kategorie,
  type Steuersatz,
  STEUERSATZ_LABEL,
} from '@/admin/products/Produkt'
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

import { EuroInput } from './EuroInput'

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

function PasswordInput<AllFormFields extends FieldValues>({
  form,
  placeholder,
  autoComplete,
}: FieldProps<{ password: string } & AllFormFields> & {
  placeholder: string
  autoComplete: string
}) {
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
              placeholder={placeholder}
              autoComplete={autoComplete}
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

export function PasswordField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ password: string } & AllFormFields>) {
  return (
    <PasswordInput
      form={form}
      placeholder={placeholder ?? 'Passwort'}
      autoComplete="current-password"
    />
  )
}

export function NewPasswordField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ password: string } & AllFormFields>) {
  return (
    <PasswordInput
      form={form}
      placeholder={placeholder ?? 'Neues Passwort'}
      autoComplete="off"
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
            inputMode="numeric"
            autoComplete="one-time-code"
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

const centsToDisplay = (value: unknown): string =>
  typeof value === 'number' && value > 0 ? formatCents(value) : ''

interface EuroFieldProps<TField extends FieldValues> {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<TField, any, TField>
  name: FieldPath<TField>
  label?: string
  withLabel?: boolean
  placeholder?: string
  description?: string
  className?: string
}

/**
 * Form field for a Euro amount. Stores integer cents in the form while showing
 * the user-friendly Euro string via {@link EuroInput}. Use for any monetary
 * input bound to react-hook-form.
 */
export function EuroField<TField extends FieldValues>({
  form,
  name,
  label,
  withLabel,
  placeholder,
  description,
  className,
}: EuroFieldProps<TField>) {
  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <EuroFieldControl
          field={field as ControllerRenderProps}
          fieldState={fieldState}
          id={`form-${name}`}
          label={label}
          withLabel={withLabel}
          placeholder={placeholder}
          description={description}
          className={className}
        />
      )}
    />
  )
}

function EuroFieldControl({
  field,
  fieldState,
  id,
  label,
  withLabel,
  placeholder,
  description,
  className,
}: {
  field: ControllerRenderProps
  fieldState: ControllerFieldState
  id: string
  label?: string
  withLabel?: boolean
  placeholder?: string
  description?: string
  className?: string
}) {
  const cents = field.value as number
  const [display, setDisplay] = useState<string>(() => centsToDisplay(cents))
  const lastEmitted = useRef<number>(cents)

  // Reflect external changes (e.g. form.reset after a successful booking) in the
  // display without overwriting an amount the user is currently typing.
  useEffect(() => {
    if (cents !== lastEmitted.current) {
      setDisplay(centsToDisplay(cents))
      lastEmitted.current = cents
    }
  }, [cents])

  return (
    <Field data-invalid={fieldState.invalid} className="gap-1">
      {withLabel && <FieldLabel htmlFor={id}>{label ?? 'Betrag'}</FieldLabel>}
      {description && <FieldDescription>{description}</FieldDescription>}
      <EuroInput
        id={id}
        aria-invalid={fieldState.invalid}
        placeholder={placeholder}
        className={className}
        value={display}
        onValueChange={(raw) => {
          setDisplay(raw)
          const next = parseCents(raw)
          lastEmitted.current = next
          field.onChange(next)
        }}
        onBlur={field.onBlur}
      />
      {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
    </Field>
  )
}

/** Euro input for a product's net price, bound to the `preisCents` form field. */
export function PriceField<AllFormFields extends FieldValues>({
  form,
  withLabel,
  placeholder,
}: FieldProps<{ preisCents: number } & AllFormFields>) {
  return (
    <EuroField
      form={form}
      name={'preisCents' as FieldPath<{ preisCents: number } & AllFormFields>}
      label="Preis"
      withLabel={withLabel}
      placeholder={placeholder ?? 'Preis in Euro (z.B. 4,50)'}
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
              <SelectItem value="regel">{STEUERSATZ_LABEL.regel}</SelectItem>
              <SelectItem value="ermaessigt">
                {STEUERSATZ_LABEL.ermaessigt}
              </SelectItem>
              <SelectItem value="befreit">
                {STEUERSATZ_LABEL.befreit}
              </SelectItem>
              <SelectItem value="kombi">{STEUERSATZ_LABEL.kombi}</SelectItem>
            </SelectContent>
          </Select>
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
