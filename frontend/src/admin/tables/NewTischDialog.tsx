import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
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
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import type { Tisch } from './Tisch'
import { CreateTischSchema, TischBackend } from './TischBackend'

const FormDataSchema = CreateTischSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewTischDialogProps {
  backend: Pick<TischBackend, 'createTisch'>
  created: (tisch: Tisch) => void
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
    byCode: {
      tisch_already_exists: 'Dieser Name ist bereits vergeben.',
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      const id = await props.backend.createTisch(data)
      form.reset()
      setOpen(false)
      props.created({
        id,
        ...data,
        status: 'inactive',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] right-4 md:bottom-16 md:right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <Plus /> Neuer Tisch
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Tisch anlegen</DialogTitle>
          <DialogDescription>
            Den Namen kannst du später jederzeit ändern.
          </DialogDescription>
        </DialogHeader>
        <form
          id="table-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
            return false
          }}
        >
          <FieldGroup>
            <NameField form={form} withLabel placeholder="z.B. Tisch 34" />
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
