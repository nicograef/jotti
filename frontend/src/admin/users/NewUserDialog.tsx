import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { type User, UserRole } from '@/lib/user/User'
import { CreateUserRequestSchema, UserBackend } from '@/lib/user/UserBackend'

const FormDataSchema = CreateUserRequestSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  backend: Pick<UserBackend, 'createUser'>
  created: (user: User, onetimePassword: string) => void
}

export function NewUserDialog(props: NewUserDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: { name: '', username: '', role: UserRole.SERVICE },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const { user, onetimePassword } = await props.backend.createUser(
        data.name,
        data.username,
        data.role as UserRole,
      )
      form.reset()
      setOpen(false)
      props.created(user, onetimePassword)
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-16 right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <Plus /> Neuer Benutzer
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Benutzer anlegen</DialogTitle>
          <DialogDescription>
            Das Passwort kann der Benutzer später selbst festlegen.
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
            {loading ? <Spinner /> : <></>} Benutzer anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
