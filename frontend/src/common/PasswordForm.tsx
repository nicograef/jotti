import { zodResolver } from "@hookform/resolvers/zod"
import { REGEXP_ONLY_DIGITS } from "input-otp"
import React from "react"
import { Controller, useForm } from "react-hook-form"
import { NavLink, useNavigate } from "react-router"
import type z from "zod"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp"
import { Spinner } from "@/components/ui/spinner"
import { AuthSingleton } from "@/lib/auth"
import { BackendError, BackendSingleton } from "@/lib/backend"
import { SetPasswordRequestSchema, toUsername } from "@/lib/user"

const FormDataSchema = SetPasswordRequestSchema
type FormData = z.infer<typeof FormDataSchema>

export function PasswordForm() {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const form = useForm<FormData>({
    resolver: zodResolver(FormDataSchema),
    mode: "onBlur",
    defaultValues: { username: "", password: "", onetimePassword: "" },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const token = await BackendSingleton.setPassword(
        data.username,
        data.password,
        data.onetimePassword,
      )
      AuthSingleton.validateAndSetToken(token)
      if (AuthSingleton.isAdmin) {
        await navigate("/admin")
      } else {
        await navigate("/")
      }
    } catch (error: unknown) {
      console.error(error)

      if (error instanceof BackendError) {
        if (error.code === "invalid_credentials") {
          form.setError("username", {
            type: "manual",
            message: "Benutzername oder Code ungültig.",
          })
          form.setError("onetimePassword", {
            type: "manual",
            message: "Benutzername oder Code ungültig.",
          })
        } else if (error.code === "already_has_password") {
          form.setError("password", {
            type: "manual",
            message: "Dieses Konto hat bereits ein Passwort festgelegt.",
          })
        }
      }
    }

    setLoading(false)
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <h1 className="text-4xl text-center font-extrabold">jotti</h1>
      </CardHeader>
      <CardDescription className="text-center mb-4 px-8">
        Lege ein neues Passwort für dein Konto fest.
      </CardDescription>
      <CardContent>
        <form
          id="password-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
            return false
          }}
        >
          <FieldGroup className="gap-2">
            <Controller
              name="username"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <Input
                    {...field}
                    onChange={(e) => {
                      const username = toUsername(e.target.value)
                      field.onChange(username)
                    }}
                    aria-invalid={fieldState.invalid}
                    placeholder="Benutzername"
                    autoComplete="off"
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              name="password"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <Input
                    {...field}
                    type="password"
                    aria-invalid={fieldState.invalid}
                    placeholder="Neues Passwort"
                    autoComplete="off"
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          </FieldGroup>
          <FieldGroup
            className="my-8"
            hidden={
              !form.formState.dirtyFields.username ||
              !form.formState.dirtyFields.password
            }
          >
            <Controller
              name="onetimePassword"
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
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                  <FieldDescription className="text-center">
                    Gib deinen Code ein.
                  </FieldDescription>
                </Field>
              )}
            />
          </FieldGroup>
        </form>
      </CardContent>
      <CardFooter className="flex-col gap-4">
        <Button
          type="submit"
          form="password-form"
          className="w-full"
          disabled={loading || !form.formState.isValid}
        >
          {loading ? <Spinner /> : <></>} Passwort festlegen
        </Button>
        <Button asChild className="w-full" variant="link" disabled={loading}>
          <NavLink to="/login">Zum Login</NavLink>
        </Button>
      </CardFooter>
    </Card>
  )
}
