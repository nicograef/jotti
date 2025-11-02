import React from "react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { AuthSingleton } from "@/lib/auth"
import { Controller, useForm } from "react-hook-form"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { useNavigate } from "react-router"

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

export function Login() {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const form = useForm<FormData>({
    defaultValues: { username: "", password: "" },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      await AuthSingleton.login(data.username, data.password)
      if (AuthSingleton.isAdmin) {
        console.log("Redirecting to /admin")
        await navigate("/admin")
      } else {
        await navigate("/")
      }
    } catch (error: unknown) {
      console.error("Login failed:", error)
    }

    setLoading(false)
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <form
        id="login-form"
        onSubmit={(e) => {
          e.preventDefault()
          void form.handleSubmit(onSubmit)()
          return false
        }}
        className="w-full max-w-sm space-y-4"
      >
        <FieldGroup>
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
                <FieldLabel htmlFor="login-form-username">
                  Benutzername
                </FieldLabel>
                <Input
                  {...field}
                  onChange={(e) => {
                    const username = toUsername(e.target.value)
                    field.onChange(username)
                  }}
                  id="login-form-username"
                  aria-invalid={fieldState.invalid}
                  placeholder="nico"
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
                <FieldLabel htmlFor="login-form-password">Passwort</FieldLabel>
                <Input
                  {...field}
                  id="login-form-password"
                  type="password"
                  aria-invalid={fieldState.invalid}
                  placeholder="••••••••"
                  autoComplete="current-password"
                />
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />
        </FieldGroup>
        <Button
          type="submit"
          form="login-form"
          className="w-full"
          disabled={loading}
        >
          {loading ? <Spinner /> : <></>} Anmelden
        </Button>
      </form>
    </div>
  )
}
