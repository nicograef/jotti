import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField, PriceField } from '@/components/common/FormFields'
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
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import type { Variante } from './Produkt'
import { type ProduktBackend, UpdateVarianteSchema } from './ProduktBackend'

const FormDataSchema = UpdateVarianteSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditVariantDialogProps {
  open: boolean
  variant: Variante
  backend: Pick<ProduktBackend, 'updateVariante'>
  updated: (variante: Variante) => void
  close: () => void
}

export function EditVariantDialog(props: EditVariantDialogProps) {
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
            return false
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
            form="edit-variant-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
