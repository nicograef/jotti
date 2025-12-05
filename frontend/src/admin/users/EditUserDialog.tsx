import { zodResolver } from '@hookform/resolvers/zod'
import { DialogDescription } from '@radix-ui/react-dialog'
import { useState } from 'react'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { BackendError } from '@/lib/Backend'
import { type User, UserSchema } from '@/lib/user/User'
import type { UserBackend } from '@/lib/user/UserBackend'

const FormDataSchema = UserSchema.pick({
  name: true,
  username: true,
  role: true,
})
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  backend: Pick<UserBackend, 'updateUser'>
  open: boolean
  user: User
  updated: (user: User) => void
  close: () => void
}

export function EditUserDialog(props: Readonly<NewUserDialogProps>) {
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
      const updatedUser = await props.backend.updateUser({
        id: props.user.id,
        ...data,
      })
      form.reset()
      props.updated(updatedUser)
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
          <DialogClose asChild>
            <Button
              variant="outline"
              onClick={() => {
                form.reset()
              }}
              disabled={loading}
            >
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
