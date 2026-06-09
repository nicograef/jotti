import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { getActionErrorMessage } from '@/lib/errorMessages'

import {
  defaultSteuersatzByKategorie,
  Kategorie,
  type Produkt,
} from './Produkt'
import { CreateProduktSchema, ProduktBackend } from './ProduktBackend'

const FormDataSchema = CreateProduktSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewProductDialogProps {
  backend: Pick<ProduktBackend, 'createProdukt'>
  created: (product: Produkt) => void
}

export function NewProductDialog(props: NewProductDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: {
      name: '',
      kategorie: Kategorie.ESSEN,
      steuersatz: defaultSteuersatzByKategorie(Kategorie.ESSEN),
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const kategorie = useWatch({ control: form.control, name: 'kategorie' })

  useEffect(() => {
    form.setValue('steuersatz', defaultSteuersatzByKategorie(kategorie), {
      shouldValidate: true,
    })
  }, [form, kategorie])

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const id = await props.backend.createProdukt(data)
      form.reset()
      setOpen(false)
      props.created({
        id,
        ...data,
        status: 'active',
        varianten: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
    } catch (error: unknown) {
      console.error(error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Produkt anlegen', error }),
      )
    }

    setLoading(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] right-4 md:bottom-16 md:right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <Plus /> Neues Produkt
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neues Produkt anlegen</DialogTitle>
          <DialogDescription>
            Produkte haben einen Namen und eine Kategorie. Varianten mit Preisen
            können später hinzugefügt werden.
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
            {loading ? <Spinner /> : null} Produkt anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
