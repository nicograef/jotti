import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
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

import { tischBackend } from './hooks'
import { CreateTischSchema } from './TischBackend'

const FormDataSchema = CreateTischSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewTischDialogProps {
  created: (tisch: FormData) => void
}

export function NewTischDialog(props: NewTischDialogProps) {
  const [open, setOpen] = useState(false)
  const form = useForm<FormData>({
    defaultValues: { name: '' },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Tisch anlegen',
    fieldErrorsByCode: {
      tisch_already_exists: { name: 'Dieser Name ist bereits vergeben.' },
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await tischBackend.createTisch(data)
      form.reset()
      setOpen(false)
      props.created(data)
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
          <Plus /> Neuer Tisch
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Tisch anlegen</DialogTitle>
          <DialogDescription>
            Den Namen kannst du später jederzeit ändern.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="table-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} placeholder="z.B. Tisch 34" />
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
            form="table-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Tisch anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
