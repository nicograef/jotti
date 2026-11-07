import { zodResolver } from '@hookform/resolvers/zod'
import { type ReactNode, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField, PriceField } from '@/components/common/FormFields'
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

import { produktBackend } from './hooks'
import { CreateVarianteSchema } from './ProduktBackend'

const FormDataSchema = CreateVarianteSchema.omit({ produktId: true })
type FormData = z.infer<typeof FormDataSchema>

interface NewVariantDialogProps {
  productId: number
  created: (variante: FormData) => void
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
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await produktBackend.createVariante({
        produktId: props.productId,
        ...data,
      })
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
      <DialogTrigger asChild>{props.children}</DialogTrigger>
      <DialogContent className="sm:max-w-100">
        <DialogHeader className="mb-4">
          <DialogTitle>Neue Variante anlegen</DialogTitle>
          <DialogDescription>
            Varianten haben einen Namen und Preis. Sie können später aktiviert
            werden.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="variant-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} placeholder="z.B. Klein, Groß, 0.5L" />
              <PriceField form={form} />
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
