import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useRef } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { z } from 'zod'

import { CategoryField, NameField } from '@/components/common/FormFields'
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
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import type { Produkt } from '@/lib/produktSchemas'

import { produktBackend } from './hooks'
import { PfandField } from './PfandField'
import { defaultSteuersatzByKategorie } from './Produkt'
import { UpdateProduktSchema } from './ProduktBackend'
import { SteuersatzField } from './SteuersatzField'

const FormDataSchema = UpdateProduktSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditProductDialogProps {
  product: Produkt
  updated: () => void
  close: () => void
}

export function EditProductDialog(props: EditProductDialogProps) {
  const form = useForm<FormData>({
    defaultValues: {
      name: props.product.name,
      kategorie: props.product.kategorie,
      steuersatz: props.product.steuersatz,
      pfand: props.product.pfand,
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })
  const previousKategorie = useRef(props.product.kategorie)
  const kategorie = useWatch({ control: form.control, name: 'kategorie' })

  useEffect(() => {
    if (kategorie !== previousKategorie.current) {
      previousKategorie.current = kategorie
      form.setValue('steuersatz', defaultSteuersatzByKategorie(kategorie), {
        shouldValidate: true,
      })
    }
  }, [form, kategorie])

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Produkt speichern',
    fieldErrorsByCode: {
      produkt_already_exists: { name: 'Dieser Name ist bereits vergeben.' },
    },
  })

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset()
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await produktBackend.updateProdukt({
        id: props.product.id,
        ...data,
      })
      props.updated()
      props.close()
    })
  }

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.product.name}</DialogTitle>
          <DialogDescription>
            Du kannst Name, Kategorie, Steuersatz und Pfand des Produkts ändern.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="product-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} />
              <CategoryField form={form} />
              <SteuersatzField form={form} />
              <PfandField form={form} />
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
            form="product-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : null} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
