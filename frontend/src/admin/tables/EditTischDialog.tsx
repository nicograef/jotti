import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
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
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { getActionErrorMessage } from '@/lib/errorMessages'

import type { Tisch } from './Tisch'
import { TischBackend, UpdateTischSchema } from './TischBackend'

const FormDataSchema = UpdateTischSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditTischDialogProps {
  backend: Pick<TischBackend, 'updateTisch'>
  open: boolean
  tisch: Tisch
  updated: (tisch: Tisch) => void
  close: () => void
}

export function EditTischDialog(props: EditTischDialogProps) {
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: props.tisch,
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
      await props.backend.updateTisch({
        id: props.tisch.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.tisch, ...data })
      props.close()
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Tisch speichern', error }),
      )
    }

    setLoading(false)
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.tisch.name}</DialogTitle>
          <DialogDescription>
            Du kannst Namen und Status des Tisches ändern.
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
            <NameField form={form} withLabel />
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
            {loading ? <Spinner /> : null} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
