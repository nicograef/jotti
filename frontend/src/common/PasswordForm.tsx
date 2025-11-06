import React from "react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { AuthSingleton } from "@/lib/auth"
import { Controller, useForm } from "react-hook-form"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
} from "@/components/ui/field"
import { NavLink, useNavigate } from "react-router"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
} from "@/components/ui/card"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp"
import { REGEXP_ONLY_DIGITS } from "input-otp"
import { BackendError, BackendSingleton } from "@/lib/backend"

interface FormData {
  username: string
  password: string
  onetimePassword: string
}

function toUsername(name: string) {
  return name
    .toLowerCase()
    .replace(/\s+/g, "")
    .replace(/ä/g, "ae")
    .replace(/ö/g, "oe")
    .replace(/ü/g, "ue")
    .replace(/ß/g, "ss")
    .replace(/[^a-z0-9]/g, "")
}

export function PasswordForm() {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const form = useForm<FormData>({
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
            message: "Benutzername oder Code ist ungültig.",
          })
          form.setError("onetimePassword", {
            type: "manual",
            message: "Benutzername oder Code ist ungültig.",
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
              rules={{
                required: "Benutzername fehlt.",
                minLength: {
                  value: 4,
                  message: "Benutzername muss mindestens 4 Zeichen lang sein.",
                },
                maxLength: {
                  value: 20,
                  message: "Benutzername darf maximal 20 Zeichen lang sein.",
                },
              }}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
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
              rules={{
                required: "Passwort fehlt.",
                minLength: {
                  value: 6,
                  message: "Passwort muss mindestens 6 Zeichen lang sein.",
                },
                maxLength: {
                  value: 20,
                  message: "Passwort darf maximal 20 Zeichen lang sein.",
                },
              }}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
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
              rules={{
                required: "Code fehlt.",
              }}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
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
