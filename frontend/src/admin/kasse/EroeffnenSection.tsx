import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { useTSEKonfiguration } from '@/admin/tse/hooks'
import { EuroField } from '@/components/common/FormFields'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { kasseBackend } from './hooks'
import { BetragCentsSchema, BezeichnungSchema } from './Kassensitzung'

export function EroeffnenSection({ onSuccess }: { onSuccess: () => void }) {
  const { tseKonfiguration } = useTSEKonfiguration()
  const [tseDialogOpen, setTseDialogOpen] = useState(false)
  const [pendingData, setPendingData] = useState<FormData | null>(null)

  const FormDataSchema = z.object({
    bezeichnung: BezeichnungSchema,
    betragCents: BetragCentsSchema.gte(1, {
      message: 'Bitte einen Anfangsbestand eingeben.',
    }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: {
      bezeichnung: '',
      betragCents: 0,
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Kassensitzung eröffnen',
    byCode: {
      betreiber_nicht_konfiguriert:
        'Die Betreiber-Stammdaten sind nicht vollständig hinterlegt. Bitte zuerst im Bereich Finanzamt die Betreiber-Stammdaten pflegen, dann die Kassensitzung eröffnen.',
    },
  })

  const eroeffnen = async (data: FormData) => {
    await run(async () => {
      await kasseBackend.kassensitzungEroeffnen(
        data.bezeichnung,
        data.betragCents,
      )
      toast.success('Kassensitzung eröffnet.')
      setTseDialogOpen(false)
      onSuccess()
    })
  }

  const onSubmit = async (data: FormData) => {
    if (tseKonfiguration?.istKonfiguriert) {
      await eroeffnen(data)
      return
    }
    setPendingData(data)
    setTseDialogOpen(true)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensitzung eröffnen</CardTitle>
        <CardDescription>Die Kasse für den Verkaufstag öffnen</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4">
          Trage den Anfangsbestand ein, also das Wechselgeld, das beim Start in
          der Kasse liegt. Die Kassensitzung läuft, bis die Kasse am Ende des
          Tages abgeschlossen wird.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <Field
              data-invalid={!!form.formState.errors.bezeichnung}
              className="gap-1"
            >
              <FieldLabel htmlFor="ks-bezeichnung">Bezeichnung</FieldLabel>
              <Input
                id="ks-bezeichnung"
                {...form.register('bezeichnung')}
                aria-invalid={!!form.formState.errors.bezeichnung}
                placeholder="z.B. Sommerfest Tag 1"
                className="w-72"
              />
              {form.formState.errors.bezeichnung && (
                <FieldError errors={[form.formState.errors.bezeichnung]} />
              )}
            </Field>
            <EuroField
              form={form}
              name="betragCents"
              label="Anfangsbestand"
              withLabel
              placeholder="z.B. 150,00"
              className="w-44"
            />
            <div>
              <Button type="submit" disabled={loading}>
                Kassensitzung eröffnen
              </Button>
            </div>
          </FieldGroup>
        </form>

        <AlertDialog open={tseDialogOpen} onOpenChange={setTseDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Keine TSE konfiguriert</AlertDialogTitle>
              <AlertDialogDescription>
                Es ist keine technische Sicherheitseinrichtung (TSE)
                eingerichtet. Vorgänge dieser Kassensitzung werden nicht nach §
                146a AO signiert. Trotzdem eröffnen?
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={loading}>
                Abbrechen
              </AlertDialogCancel>
              <AlertDialogAction
                disabled={loading}
                onClick={(e) => {
                  e.preventDefault()
                  if (pendingData) void eroeffnen(pendingData)
                }}
              >
                Trotzdem eröffnen
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  )
}
