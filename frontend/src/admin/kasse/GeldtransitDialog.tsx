import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { EuroField } from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { kasseBackend } from './hooks'
import {
  BetragCentsSchema,
  type GeldtransitRichtung,
  KommentarSchema,
} from './Kassensitzung'

// GeldtransitDialog bucht eine einzelne Bargeldbewegung mit fest vorgegebener
// Richtung (die Buttons „+ Geld einlegen" / „− Geld entnehmen" wählen sie). Die
// Richtung wird nicht mehr im Formular gewählt — das ist der Unterschied zur
// früheren GeldtransitSection mit Richtungs-Umschalter.
export function GeldtransitDialog({
  open,
  onOpenChange,
  richtung,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  richtung: GeldtransitRichtung | null
  onSuccess: () => void
}) {
  // geldtransitId pro logischem Vorgang (nicht pro Retry). Neue ID nach Erfolg.
  const [geldtransitId, setGeldtransitId] = useState(() => crypto.randomUUID())

  const FormDataSchema = z.object({
    betragCents: BetragCentsSchema.gte(1, {
      message: 'Bitte einen Betrag größer als 0 eingeben.',
    }),
    kommentar: KommentarSchema,
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: { betragCents: 0, kommentar: '' },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  // Bei jedem Öffnen ein sauberes Formular (Betrag/Kommentar leer).
  useEffect(() => {
    if (open) form.reset({ betragCents: 0, kommentar: '' })
  }, [open, form])

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Kassenbewegung buchen',
  })

  const istEinlage = richtung === 'einlage'
  const titel = istEinlage ? 'Geld einlegen' : 'Geld entnehmen'

  const onSubmit = async (data: FormData) => {
    if (richtung === null) return
    await run(async () => {
      await kasseBackend.geldtransitBuchen(
        geldtransitId,
        richtung,
        data.betragCents,
        data.kommentar,
      )
      toast.success('Kassenbewegung gebucht.')
      setGeldtransitId(crypto.randomUUID())
      onOpenChange(false)
      onSuccess()
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{titel}</DialogTitle>
          <DialogDescription>
            {istEinlage
              ? 'Zusätzliches Bargeld in die Kasse legen (z.B. Wechselgeld). Der Soll-Bestand steigt entsprechend.'
              : 'Bargeld aus der Kasse nehmen (z.B. Abschöpfung in den Tresor). Der Soll-Bestand sinkt entsprechend.'}
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <EuroField
              form={form}
              name="betragCents"
              label="Betrag"
              withLabel
              placeholder="z.B. 25,00"
              className="w-44"
            />
            <Field
              data-invalid={!!form.formState.errors.kommentar}
              className="gap-1"
            >
              <FieldLabel htmlFor="bewegung-kommentar">Kommentar</FieldLabel>
              <Input
                id="bewegung-kommentar"
                {...form.register('kommentar')}
                aria-invalid={!!form.formState.errors.kommentar}
                placeholder={
                  istEinlage
                    ? 'z.B. Wechselgeld Nachschub'
                    : 'z.B. Abschöpfung in den Tresor'
                }
              />
              {form.formState.errors.kommentar && (
                <FieldError errors={[form.formState.errors.kommentar]} />
              )}
            </Field>
          </FieldGroup>
          <DialogFooter className="mt-4">
            <Button
              type="button"
              variant="outline"
              disabled={loading}
              onClick={() => {
                onOpenChange(false)
              }}
            >
              Abbrechen
            </Button>
            <Button type="submit" disabled={loading}>
              {titel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
