import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { CategoryField, NameField } from '@/components/common/FormFields'
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

import type { Produkt } from './Product'
import { ProductBackend, UpdateProductSchema } from './ProductBackend'

const FormDataSchema = UpdateProductSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditProductDialogProps {
  backend: Pick<ProductBackend, 'updateProduct'>
  open: boolean
  product: Produkt
  updated: (product: Produkt) => void
  close: () => void
}

export function EditProductDialog(props: EditProductDialogProps) {
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: {
      name: props.product.name,
      kategorie: props.product.kategorie,
    },
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
      await props.backend.updateProduct({
        id: props.product.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.product, ...data })
      props.close()
    } catch (error: unknown) {
      console.error(error)
      toast.error('Aktion fehlgeschlagen')
    }

    setLoading(false)
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
            return false
          }}
        >
          <FieldGroup>
            <NameField form={form} withLabel />
            <CategoryField form={form} withLabel />
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
            {loading ? <Spinner /> : <></>} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
