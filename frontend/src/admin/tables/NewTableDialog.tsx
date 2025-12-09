import { zodResolver } from '@hookform/resolvers/zod'
import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
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
import { BackendError } from '@/lib/Backend'

import type { Table } from './Table'
import { CreateTableSchema, TableBackend } from './TableBackend'

const FormDataSchema = CreateTableSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewTableDialogProps {
  backend: Pick<TableBackend, 'createTable'>
  created: (table: Table) => void
}

export function NewTableDialog(props: NewTableDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: { name: '' },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const id = await props.backend.createTable(data)
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

      if (error instanceof BackendError) {
        if (error.code === 'table_already_exists') {
          form.setError('name', {
            type: 'custom',
            message: 'Dieser Name ist bereits vergeben.',
          })
        }
      }
    }

    setLoading(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-16 right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <Plus /> Neuer Tisch
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Tisch anlegen</DialogTitle>
          <DialogDescription>
            Den Namen kannst du später jederzeit ändern.
          </DialogDescription>
        </DialogHeader>
        <form
          id="table-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
            return false
          }}
        >
          <FieldGroup>
            <NameField form={form} withLabel placeholder="z.B. Tisch 34" />
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
            form="table-form"
            disabled={loading || !form.formState.isValid}
          >
            {loading ? <Spinner /> : <></>} Tisch anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
