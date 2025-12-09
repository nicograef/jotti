import { zodResolver } from '@hookform/resolvers/zod'
import { DialogDescription } from '@radix-ui/react-dialog'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { BackendError } from '@/lib/Backend'

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
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: props.user,
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset()
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      await props.backend.updateUser({
        id: props.user.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.user, ...data })
      props.close()
    } catch (error: unknown) {
      console.error(error)

      if (error instanceof BackendError) {
        if (error.code === 'username_already_exists') {
          form.setError('username', {
            type: 'custom',
            message: 'Dieser Benutzername ist bereits vergeben.',
          })
        }
      }
    }

    setLoading(false)
  }

  const resetPassword = async () => {
    setLoading(true)

    try {
      const onetimePassword = await props.backend.resetPassword(props.user.id)
      form.reset()
      props.onPasswordReset(props.user.username, onetimePassword)
      props.close()
    } catch (error: unknown) {
      console.error(error)
      toast.error('Fehler beim Zurücksetzen des Passworts.')
    }

    setLoading(false)
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
            {loading ? <Spinner /> : <></>} Passwort zurücksetzen
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
            {loading ? <Spinner /> : <></>} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
