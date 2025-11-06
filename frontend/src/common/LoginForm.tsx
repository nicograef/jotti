import React from "react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { AuthSingleton } from "@/lib/auth"
import { Controller, useForm } from "react-hook-form"
import { Field, FieldError, FieldGroup } from "@/components/ui/field"
import { NavLink, useNavigate } from "react-router"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"
import { BackendError, BackendSingleton } from "@/lib/backend"

interface FormData {
  username: string
  password: string
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

export function LoginForm() {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const form = useForm<FormData>({
    defaultValues: { username: "", password: "" },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const token = await BackendSingleton.login(data.username, data.password)
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
            message: "Benutzername oder Passwort ist ungültig.",
          })
          form.setError("password", {
            type: "manual",
            message: "Benutzername oder Passwort ist ungültig.",
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
      <CardContent>
        <form
          id="login-form"
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
                    placeholder="Passwort"
                    autoComplete="current-password"
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          </FieldGroup>
        </form>
      </CardContent>
      <CardFooter className="flex-col gap-4">
        <Button
          type="submit"
          form="login-form"
          className="w-full"
          disabled={loading || !form.formState.isValid}
        >
          {loading ? <Spinner /> : <></>} Anmelden
        </Button>
        <Button asChild className="w-full" variant="link" disabled={loading}>
          <NavLink to="/set-password">Neues Passwort festlegen</NavLink>
        </Button>
      </CardFooter>
    </Card>
  )
}
