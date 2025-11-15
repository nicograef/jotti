import { zodResolver } from '@hookform/resolvers/zod'
import { UserPlus } from 'lucide-react'
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
import { BackendSingleton } from '@/lib/backend'
import {
  CreateTableRequestSchema,
  type Table,
  TableBackend,
} from '@/lib/TableBackend'

const FormDataSchema = CreateTableRequestSchema
type FormData = z.infer<typeof FormDataSchema>

interface NewTableDialogProps {
  created: (table: Table) => void
}

export function NewTableDialog({ created }: NewTableDialogProps) {
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
      const table = await new TableBackend(BackendSingleton).createTable(data)
      form.reset()
      setOpen(false)
      created(table)
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
            <UserPlus /> Neuer Tisch
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
