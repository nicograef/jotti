import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { type Product, ProductCategory } from '@/lib/product/Product'
import {
  CreateProductSchema,
  ProductBackend,
} from '@/lib/product/ProductBackend'

const FormDataSchema = CreateProductSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewProductDialogProps {
  backend: Pick<ProductBackend, 'createProduct'>
  created: (product: Product) => void
}

export function NewProductDialog(props: NewProductDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: {
      name: '',
      description: '',
      netPriceCents: 0,
      category: ProductCategory.FOOD,
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const id = await props.backend.createProduct(data)
      form.reset()
      setOpen(false)
      props.created({
        id,
        ...data,
        status: 'inactive',
        createdAt: new Date().toISOString(),
      })
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-16 right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <Plus /> Neues Produkt
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neues Produkt anlegen</DialogTitle>
          <DialogDescription>
            Du kannst alle Angaben später auch ändern.
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
            <NameField
              form={form}
              withLabel
              placeholder="Produktname eingeben"
            />
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
            {loading ? <Spinner /> : <></>} Produkt anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
