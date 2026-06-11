import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import {
  NameField,
  RoleField,
  UsernameField,
} from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { type User, UserSchema } from './User'
import type { UserBackend } from './UserBackend'

const FormDataSchema = UserSchema.pick({
  name: true,
  username: true,
  role: true,
})
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  backend: Pick<UserBackend, 'updateUser' | 'resetPassword'>
  open: boolean
  user: User
  updated: (user: User) => void
  onPasswordReset: (username: string, onetimePassword: string) => void
  close: () => void
}

export function EditUserDialog(props: NewUserDialogProps) {
  const form = useForm<FormData>({
    defaultValues: props.user,
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading: saveLoading, run: runSave } = useFormActionSubmit({
    form,
    actionLabel: 'Benutzer speichern',
    fieldErrorsByCode: {
      username_already_exists: {
        username: 'Dieser Benutzername ist bereits vergeben.',
      },
    },
  })
  const { loading: resetLoading, run: runResetPassword } = useActionSubmit({
    actionLabel: 'Passwort zurücksetzen',
  })
  const loading = saveLoading || resetLoading

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset()
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    await runSave(async () => {
      await props.backend.updateUser({
        id: props.user.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.user, ...data })
      props.close()
    })
  }

  const resetPassword = async () => {
    await runResetPassword(async () => {
      const onetimePassword = await props.backend.resetPassword(props.user.id)
      form.reset()
      props.onPasswordReset(props.user.username, onetimePassword)
      props.close()
    })
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.user.name}</DialogTitle>
          <DialogDescription>
            Du kannst Name, Benutzername, Rolle und Status des Benutzers ändern.
          </DialogDescription>
        </DialogHeader>
        <form
          id="user-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
            return false
          }}
        >
          <FieldGroup>
            <NameField form={form} withLabel />
            <UsernameField form={form} withLabel />
            <RoleField form={form} withLabel />
          </FieldGroup>
        </form>
        <DialogFooter className="mt-4">
          <Button
            variant="ghost"
            disabled={loading || !form.formState.isValid}
            onClick={() => {
              void resetPassword()
            }}
          >
            {loading ? <Spinner /> : null} Passwort zurücksetzen
          </Button>
          <DialogClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DialogClose>
          <Button
            type="submit"
            form="user-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
