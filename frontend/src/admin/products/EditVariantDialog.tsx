import { zodResolver } from '@hookform/resolvers/zod'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField, PriceField } from '@/components/common/FormFields'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
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

import type { Variante } from './Produkt'
import { type ProduktBackend, UpdateVarianteSchema } from './ProduktBackend'

const FormDataSchema = UpdateVarianteSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditVariantDialogProps {
  open: boolean
  produktId: number
  variant: Variante
  backend: Pick<ProduktBackend, 'updateVariante' | 'deleteVariante'>
  updated: (variante: Variante) => void
  deleted: () => void
  close: () => void
}

export function EditVariantDialog(props: EditVariantDialogProps) {
  const [deleteOpen, setDeleteOpen] = useState(false)

  const form = useForm<FormData>({
    defaultValues: {
      name: props.variant.name,
      preisCents: props.variant.preisCents,
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Variante speichern',
  })

  const { loading: deleteLoading, run: runDelete } = useActionSubmit({
    actionLabel: 'Variante löschen',
  })

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset({
        name: props.variant.name,
        preisCents: props.variant.preisCents,
      })
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await props.backend.updateVariante({
        id: props.variant.id,
        ...data,
      })
      props.updated({ ...props.variant, ...data })
      props.close()
    })
  }

  const onDelete = async () => {
    await runDelete(async () => {
      await props.backend.deleteVariante(props.produktId, props.variant.id)
      setDeleteOpen(false)
      props.deleted()
      props.close()
    })
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-100">
        <DialogHeader className="mb-4">
          <DialogTitle>Variante bearbeiten</DialogTitle>
          <DialogDescription>
            Name und Preis der Variante ändern.
          </DialogDescription>
        </DialogHeader>
        <form
          id="edit-variant-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <NameField
              form={form}
              withLabel
              placeholder="z.B. Klein, Groß, 0.5L"
            />
            <PriceField form={form} withLabel />
          </FieldGroup>
        </form>
        <DialogFooter className="mt-4 sm:justify-between">
          <Button
            variant="ghost"
            className="text-destructive hover:text-destructive"
            onClick={() => {
              setDeleteOpen(true)
            }}
            disabled={loading || deleteLoading}
          >
            <Trash2 /> Variante löschen
          </Button>
          <div className="flex flex-col-reverse gap-2 sm:flex-row">
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
              form="edit-variant-form"
              disabled={loading || !form.formState.isValid}
            >
              {loading ? <Spinner /> : null} Speichern
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Variante löschen?</AlertDialogTitle>
            <AlertDialogDescription>
              Die Variante &quot;{props.variant.name}&quot; wird unwiderruflich
              gelöscht.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-white hover:bg-destructive/90"
              onClick={(e) => {
                e.preventDefault()
                void onDelete()
              }}
              disabled={deleteLoading}
            >
              Löschen
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  )
}
