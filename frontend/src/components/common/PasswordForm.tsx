import { zodResolver } from '@hookform/resolvers/zod'
import React from 'react'
import { useForm } from 'react-hook-form'
import { NavLink, useNavigate } from 'react-router'
import type z from 'zod'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
} from '@/components/ui/card'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { AuthSingleton } from '@/lib/auth'
import { AuthBackend, SetPasswordRequestSchema } from '@/lib/AuthBackend'
import { BackendError } from '@/lib/Backend'

import { NewPasswordField, OTPField, UsernameField } from './FormFields'

const FormDataSchema = SetPasswordRequestSchema
type FormData = z.infer<typeof FormDataSchema>

interface PasswordFormProps {
  backend: Pick<AuthBackend, 'setPassword'>
}

export function PasswordForm(props: PasswordFormProps) {
  const navigate = useNavigate()
  const [loading, setLoading] = React.useState(false)
  const form = useForm<FormData>({
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
    defaultValues: { username: '', password: '', onetimePassword: '' },
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const token = await props.backend.setPassword(
        data.username,
        data.password,
        data.onetimePassword,
      )
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
            message: 'Benutzername oder Code ungültig.',
          })
          form.setError('onetimePassword', {
            type: 'manual',
            message: 'Benutzername oder Code ungültig.',
          })
        } else if (error.code === 'already_has_password') {
          form.setError('password', {
            type: 'manual',
            message: 'Dieses Konto hat bereits ein Passwort festgelegt.',
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
            <UsernameField form={form} />
            <NewPasswordField form={form} />
          </FieldGroup>
          <FieldGroup
            className="my-8"
            hidden={
              !form.formState.dirtyFields.username ||
              !form.formState.dirtyFields.password
            }
          >
            <OTPField form={form} />
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
