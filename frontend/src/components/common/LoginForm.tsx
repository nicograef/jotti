import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { NavLink, useNavigate } from 'react-router'
import z from 'zod'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { AuthSingleton } from '@/lib/auth'
import { type AuthBackend, LoginSchema } from '@/lib/AuthBackend'
import { BackendError } from '@/lib/Backend'

import { PasswordField, UsernameField } from './FormFields'

const FormDataSchema = LoginSchema
type FormData = z.infer<typeof FormDataSchema>

interface LoginFormProps {
  backend: Pick<AuthBackend, 'login'>
}

export function LoginForm(props: LoginFormProps) {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    resolver: zodResolver(FormDataSchema),
    mode: 'onSubmit',
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const token = await props.backend.login(data.username, data.password)
      AuthSingleton.validateAndSetToken(token)
      if (AuthSingleton.isAdmin) {
        await navigate('/admin')
      } else {
        await navigate('/')
      }
    } catch (error: unknown) {
      console.error(error)

      if (error instanceof BackendError) {
        if (error.code === 'invalid_credentials') {
          form.setError('username', {
            type: 'manual',
            message: 'Benutzername oder Passwort ungültig.',
          })
          form.setError('password', {
            type: 'manual',
            message: 'Benutzername oder Passwort ungültig.',
          })
        } else if (error.code === 'no_password_set') {
          form.setError('username', {
            type: 'manual',
            message: 'Für dieses Konto wurde noch kein Passwort festgelegt.',
          })
        } else if (error.code === 'user_inactive') {
          form.setError('username', {
            type: 'manual',
            message: 'Dieses Konto ist deaktiviert.',
          })
        } else {
          form.setError('username', {
            type: 'manual',
            message: 'Ein Fehler ist aufgetreten. Bitte versuche es erneut.',
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
            <UsernameField form={form} />
            <PasswordField form={form} />
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
