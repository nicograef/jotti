import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { EuroField } from '@/components/common/FormFields'
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
import {
  BetragCentsSchema,
  GeldtransitRichtung,
  GeldtransitRichtungSchema,
} from './Kassensitzung'

export function GeldtransitSection({ onSuccess }: { onSuccess: () => void }) {
  // geldtransitId pro logischem Vorgang (nicht pro Retry). Neue ID nach Erfolg.
  const [geldtransitId, setGeldtransitId] = useState(() => crypto.randomUUID())

  const FormDataSchema = z.object({
    richtung: GeldtransitRichtungSchema,
    betragCents: BetragCentsSchema.gte(1, {
      message: 'Bitte einen Betrag größer als 0 eingeben.',
    }),
    kommentar: z
      .string()
      .min(3, { message: 'Kommentar muss mindestens 3 Zeichen lang sein.' })
      .max(200, { message: 'Kommentar darf maximal 200 Zeichen lang sein.' }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: {
      richtung: GeldtransitRichtung.EINLAGE,
      betragCents: 0,
      kommentar: '',
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Kassenbewegung buchen',
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await kasseBackend.geldtransitBuchen(
        geldtransitId,
        data.richtung,
        data.betragCents,
        data.kommentar,
      )
      toast.success('Kassenbewegung gebucht.')
      setGeldtransitId(crypto.randomUUID())
      form.reset()
      onSuccess()
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Geldtransit buchen</CardTitle>
        <CardDescription>
          Bargeld in die Kasse legen oder daraus entnehmen
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4">
          Buche zusätzliches Wechselgeld als Einlage oder Bargeld, das du aus
          der Kasse nimmst, als Entnahme. Der Soll-Bestand wird dabei
          automatisch angepasst.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <Field
              data-invalid={!!form.formState.errors.richtung}
              className="gap-1"
            >
              <FieldLabel>Richtung</FieldLabel>
              <div className="grid grid-cols-2 gap-3">
                {(
                  [
                    [
                      GeldtransitRichtung.EINLAGE,
                      'Einlage',
                      'Geld in die Kasse legen',
                    ],
                    [
                      GeldtransitRichtung.ENTNAHME,
                      'Entnahme',
                      'Geld aus der Kasse nehmen',
                    ],
                  ] as const
                ).map(([value, label, hint]) => (
                  <label
                    key={value}
                    className="flex cursor-pointer flex-col gap-0.5 rounded-lg border p-4 text-center transition-colors has-[:checked]:border-primary has-[:checked]:bg-primary/5 has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring"
                  >
                    <input
                      type="radio"
                      value={value}
                      className="sr-only"
                      {...form.register('richtung')}
                    />
                    <span className="font-medium">{label}</span>
                    <span className="text-muted-foreground text-xs">
                      {hint}
                    </span>
                  </label>
                ))}
              </div>
              {form.formState.errors.richtung && (
                <FieldError errors={[form.formState.errors.richtung]} />
              )}
            </Field>
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
                placeholder="z.B. Wechselgeld Nachschub"
                className="w-72"
              />
              {form.formState.errors.kommentar && (
                <FieldError errors={[form.formState.errors.kommentar]} />
              )}
            </Field>
            <div>
              <Button type="submit" disabled={loading}>
                Geldtransit buchen
              </Button>
            </div>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
