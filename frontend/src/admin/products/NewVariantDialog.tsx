import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useState } from 'react'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { type Variante, VarianteStatus } from './Produkt'
import { CreateVarianteSchema, type ProduktBackend } from './ProduktBackend'

const FormDataSchema = CreateVarianteSchema.omit({ produktId: true })
type FormData = z.infer<typeof FormDataSchema>

interface NewVariantDialogProps {
  productId: number
  backend: Pick<ProduktBackend, 'createVariante'>
  created: (variante: Variante) => void
  children: ReactNode
}

export function NewVariantDialog(props: NewVariantDialogProps) {
  const [open, setOpen] = useState(false)
  const form = useForm<FormData>({
    defaultValues: {
      name: '',
      preisCents: 0,
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Variante anlegen',
    byCode: {
      variante_already_exists: 'Dieser Name ist bereits vergeben.',
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      const id = await props.backend.createVariante({
        produktId: props.productId,
        ...data,
      })
      form.reset()
      setOpen(false)
      props.created({
        id,
        ...data,
        status: VarianteStatus.INACTIVE,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{props.children}</DialogTrigger>
      <DialogContent className="sm:max-w-100">
        <DialogHeader className="mb-4">
          <DialogTitle>Neue Variante anlegen</DialogTitle>
          <DialogDescription>
            Varianten haben einen Namen und Preis. Sie können später aktiviert
            werden.
          </DialogDescription>
        </DialogHeader>
        <form
          id="variant-form"
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
            form="variant-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Variante anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
