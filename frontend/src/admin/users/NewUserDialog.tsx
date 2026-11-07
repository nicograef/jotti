import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField, UsernameField } from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import { UserRole } from '@/lib/identity'

import { userBackend } from './hooks'
import { RoleField } from './RoleField'
import { CreateUserSchema } from './UserBackend'

const FormDataSchema = CreateUserSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  created: (user: FormData, onetimePassword: string) => void
}

export function NewUserDialog(props: NewUserDialogProps) {
  const [open, setOpen] = useState(false)
  const form = useForm<FormData>({
    defaultValues: { name: '', username: '', role: UserRole.SERVICE },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Helfer anlegen',
    fieldErrorsByCode: {
      username_already_exists: {
        username: 'Dieser Benutzername ist bereits vergeben.',
      },
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      const { onetimePassword } = await userBackend.createUser(
        data.name,
        data.username,
        data.role,
      )
      form.reset()
      setOpen(false)
      props.created(data, onetimePassword)
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        setOpen(isOpen)
        if (!isOpen) form.reset()
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <Plus /> Neuer Helfer
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Helfer anlegen</DialogTitle>
          <DialogDescription>
            Das Passwort kann der Helfer später selbst festlegen.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="user-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} />
              <UsernameField form={form} withLabel />
              <RoleField form={form} />
            </FieldGroup>
          </form>
        </DialogBody>
        <DialogFooter className="mt-4">
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
            {loading ? <Spinner /> : null} Helfer anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
