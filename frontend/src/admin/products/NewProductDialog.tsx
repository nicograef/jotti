import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { FieldGroup } from '@/components/ui/field'
import { Spinner } from '@/components/ui/spinner'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import { Kategorie } from '@/lib/produktSchemas'

import { produktBackend } from './hooks'
import { PfandField } from './PfandField'
import { defaultSteuersatzByKategorie } from './Produkt'
import { CreateProduktSchema } from './ProduktBackend'
import { SteuersatzField } from './SteuersatzField'

const FormDataSchema = CreateProduktSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewProductDialogProps {
  created: (product: FormData) => void
}

export function NewProductDialog(props: NewProductDialogProps) {
  const [open, setOpen] = useState(false)
  const form = useForm<FormData>({
    defaultValues: {
      name: '',
      kategorie: Kategorie.ESSEN,
      steuersatz: defaultSteuersatzByKategorie(Kategorie.ESSEN),
      pfand: false,
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

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Produkt anlegen',
    fieldErrorsByCode: {
      produkt_already_exists: { name: 'Dieser Name ist bereits vergeben.' },
    },
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await produktBackend.createProdukt(data)
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
      <DialogTrigger asChild>
        <Button>
          <Plus /> Neues Produkt
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neues Produkt anlegen</DialogTitle>
          <DialogDescription>
            Produkte haben einen Namen und eine Kategorie. Varianten mit Preisen
            können später hinzugefügt werden.
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
              <NameField form={form} placeholder="Produktname eingeben" />
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
            {loading ? <Spinner /> : null} Produkt anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
