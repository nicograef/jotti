import { zodResolver } from '@hookform/resolvers/zod'
import { DialogDescription } from '@radix-ui/react-dialog'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
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
import type { Table } from '@/table/Table'
import { TableBackend, UpdateTableRequestSchema } from '@/table/TableBackend'

const FormDataSchema = UpdateTableRequestSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditTableDialogProps {
  backend: Pick<TableBackend, 'updateTable'>
  open: boolean
  table: Table
  updated: (table: Table) => void
  close: () => void
}

export function EditTableDialog(props: Readonly<EditTableDialogProps>) {
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: props.table,
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
      const updatedTable = await props.backend.updateTable({
        id: props.table.id,
        ...data,
      })
      form.reset()
      props.updated(updatedTable)
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
          <DialogTitle>{props.table.name}</DialogTitle>
          <DialogDescription>
            Du kannst Namen und Status des Tisches ändern.
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
            <NameField form={form} withLabel />
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
            {loading ? <Spinner /> : <></>} Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
