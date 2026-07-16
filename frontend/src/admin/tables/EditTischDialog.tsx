import { zodResolver } from '@hookform/resolvers/zod'
import { Trash2 } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

import { NameField } from '@/components/common/FormFields'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import { formatCents } from '@/lib/utils'

import { type Tisch } from './Tisch'
import { TischBackend, UpdateTischSchema } from './TischBackend'

const FormDataSchema = UpdateTischSchema.omit({ id: true })
type FormData = z.infer<typeof FormDataSchema>

interface EditTischDialogProps {
  backend: Pick<TischBackend, 'updateTisch' | 'deleteTisch'>
  open: boolean
  tisch: Tisch
  updated: (tisch: Tisch) => void
  deleted: (tischId: number) => void
  close: () => void
}

export function EditTischDialog(props: EditTischDialogProps) {
  const hatSaldo = props.tisch.saldoCents > 0

  const form = useForm<FormData>({
    defaultValues: props.tisch,
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Tisch speichern',
    fieldErrorsByCode: {
      tisch_already_exists: { name: 'Dieser Name ist bereits vergeben.' },
    },
  })

  const { loading: deleteLoading, run: runDelete } = useActionSubmit({
    actionLabel: 'Tisch löschen',
  })

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset()
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await props.backend.updateTisch({
        id: props.tisch.id,
        ...data,
      })
      form.reset()
      props.updated({ ...props.tisch, ...data })
      props.close()
    })
  }

  const onDelete = async () => {
    await runDelete(async () => {
      await props.backend.deleteTisch(props.tisch.id)
      props.deleted(props.tisch.id)
      props.close()
    })
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.tisch.name}</DialogTitle>
          <DialogDescription>
            Du kannst den Namen des Tisches ändern.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="table-form"
            onSubmit={(e) => {
              e.preventDefault()
              void form.handleSubmit(onSubmit)()
            }}
          >
            <FieldGroup>
              <NameField form={form} withLabel />
            </FieldGroup>
          </form>
        </DialogBody>

        <DialogFooter className="mt-4 sm:justify-between">
          {hatSaldo ? (
            // Löschen ist gesperrt, solange der Tisch einen offenen Saldo trägt
            // (das Backend erzwingt es zusätzlich als Single Source of Truth).
            // Die Begründung steht als stets sichtbare Zeile — auf den
            // Touch-Handys gibt es kein Hover für einen Tooltip.
            <div className="space-y-1">
              {/* Dauerhaft deaktiviert: nutzt das gemeinsame Disabled-Token der
                  Primäraktion (neutrale Fläche + AA-Text) statt einer
                  hand-gerollten text-destructive-Abblendung. Der aktive
                  Löschen-Einstieg (rot) steht im else-Zweig. */}
              <Button className="w-full sm:w-auto" disabled>
                <Trash2 /> Tisch löschen
              </Button>
              <p className="text-xs text-muted-foreground">
                Offener Saldo: {formatCents(props.tisch.saldoCents)} € — erst
                abrechnen, dann lässt sich der Tisch löschen.
              </p>
            </div>
          ) : (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  className="text-destructive hover:text-destructive"
                  disabled={loading || deleteLoading}
                >
                  <Trash2 /> Tisch löschen
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Tisch löschen?</AlertDialogTitle>
                  <AlertDialogDescription>
                    Der Tisch &ldquo;{props.tisch.name}&rdquo; wird
                    unwiderruflich gelöscht.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Abbrechen</AlertDialogCancel>
                  <AlertDialogAction
                    className="bg-destructive text-white hover:bg-destructive/90"
                    onClick={(e) => {
                      e.preventDefault()
                      void onDelete()
                    }}
                    disabled={deleteLoading}
                  >
                    Löschen
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
          <div className="flex flex-col-reverse gap-2 sm:flex-row">
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
              {loading ? <Spinner /> : null} Speichern
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
