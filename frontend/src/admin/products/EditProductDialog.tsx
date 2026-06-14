import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useRef } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { z } from 'zod'

import {
  CategoryField,
  NameField,
  SteuersatzField,
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
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { defaultSteuersatzByKategorie, type Produkt } from './Produkt'
import { ProduktBackend, UpdateProduktSchema } from './ProduktBackend'

const FormDataSchema = UpdateProduktSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditProductDialogProps {
  backend: Pick<ProduktBackend, 'updateProdukt'>
  open: boolean
  product: Produkt
  updated: (product: Produkt) => void
  close: () => void
}

export function EditProductDialog(props: EditProductDialogProps) {
  const form = useForm<FormData>({
    defaultValues: {
      name: props.product.name,
      kategorie: props.product.kategorie,
      steuersatz: props.product.steuersatz,
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
      await props.backend.updateProdukt({
        id: props.product.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.product, ...data })
      props.close()
    })
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.product.name}</DialogTitle>
          <DialogDescription>
            Du kannst Name und Kategorie des Produkts ändern.
          </DialogDescription>
        </DialogHeader>
        <form
          id="product-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <NameField form={form} withLabel />
            <CategoryField form={form} withLabel />
            <SteuersatzField form={form} withLabel />
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
