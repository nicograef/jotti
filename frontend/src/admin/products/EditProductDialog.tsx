import { zodResolver } from '@hookform/resolvers/zod'
import { DialogDescription } from '@radix-ui/react-dialog'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import {
  CategoryField,
  DescriptionField,
  NameField,
  NetPriceField,
} from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import type { Product } from '@/lib/product/Product'
import {
  ProductBackend,
  UpdateProductSchema,
} from '@/lib/product/ProductBackend'

const FormDataSchema = UpdateProductSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditProductDialogProps {
  backend: Pick<ProductBackend, 'updateProduct'>
  open: boolean
  product: Product
  updated: (product: Product) => void
  close: () => void
}

export function EditProductDialog(props: Readonly<EditProductDialogProps>) {
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: props.product,
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
      const updatedProduct = await props.backend.updateProduct({
        id: props.product.id,
        ...data,
      })
      form.reset()
      props.updated(updatedProduct)
      props.close()
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.product.name}</DialogTitle>
          <DialogDescription>
            Du kannst Name, Benutzername, Rolle und Status des Benutzers ändern.
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
            <DescriptionField form={form} withLabel />
            <CategoryField form={form} withLabel />
            <NetPriceField form={form} withLabel />
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
