import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { NavLink, useNavigate } from 'react-router'
import z from 'zod'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { FieldError, FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { AuthSingleton } from '@/lib/Auth'
import { type AuthBackend, LoginSchema } from '@/lib/AuthBackend'
import { getActionErrorMessage } from '@/lib/errorMessages'

import { PasswordField, UsernameField } from './FormFields'
import { Wortmarke } from './Wortmarke'

type FormData = z.infer<typeof LoginSchema>

interface LoginFormProps {
  backend: Pick<AuthBackend, 'login'>
}

export function LoginForm(props: LoginFormProps) {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    resolver: zodResolver(LoginSchema),
    mode: 'onSubmit',
    defaultValues: { username: '', password: '' },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)
    form.clearErrors('root')

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
      form.setError('root', {
        type: 'manual',
        message: getActionErrorMessage({
          actionLabel: 'Anmeldung',
          error,
          byCode: {
            invalid_credentials: 'Benutzername oder Passwort ungültig.',
            no_password_set:
              'Für dieses Konto wurde noch kein Passwort festgelegt.',
            user_inactive: 'Dieses Konto ist deaktiviert.',
          },
        }),
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <Wortmarke as="h1" className="text-[38px] text-center" />
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
            <FieldError>{form.formState.errors.root?.message}</FieldError>
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
          disabled={loading}
        >
          {loading ? <Spinner /> : null} Anmelden
        </Button>
        <Button asChild className="w-full" variant="link" disabled={loading}>
          <NavLink to="/set-password">Neues Passwort festlegen</NavLink>
        </Button>
      </CardFooter>
    </Card>
  )
}
