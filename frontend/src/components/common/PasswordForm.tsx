import { zodResolver } from '@hookform/resolvers/zod'
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
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import { AuthBackend, SetPasswordSchema } from '@/lib/AuthBackend'

import { NewPasswordField, OTPField, UsernameField } from './FormFields'

type FormData = z.infer<typeof SetPasswordSchema>

interface PasswordFormProps {
  backend: Pick<AuthBackend, 'setPassword'>
}

export function PasswordForm(props: PasswordFormProps) {
  const navigate = useNavigate()
  const form = useForm<FormData>({
    resolver: zodResolver(SetPasswordSchema),
    mode: 'onTouched',
    defaultValues: { username: '', password: '', onetimePassword: '' },
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Passwort setzen',
    fieldErrorsByCode: {
      invalid_credentials: {
        username: 'Benutzername oder Code ungültig.',
        onetimePassword: 'Benutzername oder Code ungültig.',
      },
      already_has_password: {
        password: 'Dieses Konto hat bereits ein Passwort festgelegt.',
      },
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await props.backend.setPassword(
        data.username,
        data.password,
        data.onetimePassword,
      )
      await navigate('/login')
    })
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
          <FieldGroup className="my-8">
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
          {loading ? <Spinner /> : null} Passwort festlegen
        </Button>
        <Button asChild className="w-full" variant="link" disabled={loading}>
          <NavLink to="/login">Zum Login</NavLink>
        </Button>
      </CardFooter>
    </Card>
  )
}
