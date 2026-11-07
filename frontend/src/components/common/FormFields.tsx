import { REGEXP_ONLY_DIGITS } from 'input-otp'
import { EyeClosedIcon, EyeIcon } from 'lucide-react'
import { useEffect, useId, useRef, useState } from 'react'
import {
  Controller,
  type ControllerFieldState,
  type ControllerRenderProps,
  type FieldPath,
  type FieldValues,
  type Path,
  type UseFormReturn,
} from 'react-hook-form'

import { Button } from '@/components/ui/button'
import {
  Field,
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
import { toUsername } from '@/lib/identity'
import { formatCents, parseCents } from '@/lib/utils'

import { EuroInput } from './EuroInput'

export interface FieldProps<TField extends FieldValues> {
  form: UseFormReturn<TField>
  placeholder?: string
}

export function NameField<AllFormFields extends FieldValues>({
  form,
  placeholder,
}: FieldProps<{ name: string } & AllFormFields>) {
  const id = useId()
  return (
    <Controller
      name={'name' as Path<{ name: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          <FieldLabel htmlFor={id}>Name</FieldLabel>
          <Input
            {...field}
            id={id}
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
}: FieldProps<{ username: string } & AllFormFields> & {
  // Anmelde- und Passwortformular zeigen nur den Platzhalter.
  withLabel?: boolean
}) {
  const id = useId()
  return (
    <Controller
      name={'username' as Path<{ username: string } & AllFormFields>}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1">
          {withLabel && <FieldLabel htmlFor={id}>Benutzername</FieldLabel>}
          <Input
            {...field}
            id={id}
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

const centsToDisplay = (value: unknown): string =>
  typeof value === 'number' && value > 0 ? formatCents(value) : ''

interface EuroFieldProps<TField extends FieldValues> {
  form: UseFormReturn<TField>
  name: FieldPath<TField>
  label: string
  placeholder?: string
  className?: string
}

/** Euro amount field: the form value is integer cents, the display a Euro string. */
export function EuroField<TField extends FieldValues>({
  form,
  name,
  label,
  placeholder,
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
          placeholder={placeholder}
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
  placeholder,
  className,
}: {
  field: ControllerRenderProps
  fieldState: ControllerFieldState
  id: string
  label: string
  placeholder?: string
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
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
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

interface TextFieldProps<TField extends FieldValues> {
  form: UseFormReturn<TField>
  name: FieldPath<TField>
  label: string
  placeholder?: string
  description?: string
  type?: 'text' | 'password'
  autoComplete?: string
  className?: string
}

/** Labelled text field; a null form value shows as an empty input. */
export function TextField<TField extends FieldValues>({
  form,
  name,
  label,
  placeholder,
  description,
  type = 'text',
  autoComplete,
  className,
}: TextFieldProps<TField>) {
  const id = `form-${name}`
  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} className="gap-1.5">
          <FieldLabel htmlFor={id}>{label}</FieldLabel>
          <Input
            {...field}
            value={typeof field.value === 'string' ? field.value : ''}
            id={id}
            type={type}
            aria-invalid={fieldState.invalid}
            placeholder={placeholder}
            autoComplete={autoComplete}
            className={className}
          />
          {description && <FieldDescription>{description}</FieldDescription>}
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )}
    />
  )
}
