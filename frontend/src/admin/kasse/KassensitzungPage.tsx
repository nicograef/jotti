import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'
import { formatCents, parseCents } from '@/lib/utils'

import { kasseBackend, useKassenbestand, useOffeneKassensitzung } from './hooks'
import {
  BezeichnungSchema,
  GeldtransitRichtung,
  GeldtransitRichtungSchema,
} from './Kassensitzung'

export function KassensitzungPage() {
  const { kassensitzung, isPending, refetch } = useOffeneKassensitzung()
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)

  if (isPending) {
    return (
      <>
        <h1 className="text-2xl font-bold">Kassensitzung</h1>
        <p className="mt-4 text-muted-foreground">Laden…</p>
      </>
    )
  }

  return (
    <>
      <h1 className="text-2xl font-bold">Kassensitzung</h1>

      {kassensitzung ? (
        <div className="mt-4 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Kassensitzung #{String(kassensitzung.zNr)}
                <Badge variant="secondary">{kassensitzung.status}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <p>
                <span className="text-muted-foreground">Datum:</span>{' '}
                {kassensitzung.datum}
              </p>
              <p>
                <span className="text-muted-foreground">Bezeichnung:</span>{' '}
                {kassensitzung.bezeichnung}
              </p>
              {kassenbestand && (
                <p>
                  <span className="text-muted-foreground">
                    Soll-Kassenbestand:
                  </span>{' '}
                  {formatCents(kassenbestand.sollBestandCents)} €
                </p>
              )}
            </CardContent>
          </Card>

          <KassenbewegungSection onSuccess={() => void refetch()} />
          <KassensturzSection onSuccess={() => void refetch()} />
          <TagesabschlussSection onSuccess={() => void refetch()} />
        </div>
      ) : (
        <div className="mt-4 space-y-6">
          <p className="text-muted-foreground">Keine Kassensitzung geöffnet.</p>
          <EroeffnenSection onSuccess={() => void refetch()} />
        </div>
      )}
    </>
  )
}

function EroeffnenSection({ onSuccess }: { onSuccess: () => void }) {
  const FormDataSchema = z.object({
    bezeichnung: BezeichnungSchema,
    betragEuro: z
      .string()
      .min(1, { message: 'Bitte einen Betrag eingeben.' })
      .refine((val) => !isNaN(parseFloat(val.replace(',', '.'))), {
        message: 'Bitte einen gültigen Betrag eingeben.',
      }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: {
      bezeichnung: '',
      betragEuro: '',
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Kassensitzung eröffnen',
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await kasseBackend.kassensitzungEroeffnen(
        data.bezeichnung,
        parseCents(data.betragEuro),
      )
      toast.success('Kassensitzung eröffnet.')
      onSuccess()
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensitzung eröffnen</CardTitle>
      </CardHeader>
      <CardContent>
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
            <Field
              data-invalid={!!form.formState.errors.betragEuro}
              className="gap-1"
            >
              <FieldLabel htmlFor="ks-anfangsbestand">
                Anfangsbestand (€)
              </FieldLabel>
              <Input
                id="ks-anfangsbestand"
                type="text"
                inputMode="decimal"
                {...form.register('betragEuro')}
                aria-invalid={!!form.formState.errors.betragEuro}
                placeholder="z.B. 150,00"
                className="w-40"
              />
              {form.formState.errors.betragEuro && (
                <FieldError errors={[form.formState.errors.betragEuro]} />
              )}
            </Field>
            <div>
              <Button type="submit" disabled={loading}>
                Kassensitzung eröffnen
              </Button>
            </div>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}

function KassenbewegungSection({ onSuccess }: { onSuccess: () => void }) {
  const FormDataSchema = z.object({
    richtung: GeldtransitRichtungSchema,
    betragEuro: z
      .string()
      .min(1, { message: 'Bitte einen Betrag eingeben.' })
      .refine(
        (val) => {
          const parsed = parseFloat(val.replace(',', '.'))
          return !isNaN(parsed) && parsed > 0
        },
        { message: 'Bitte einen Betrag größer als 0 eingeben.' },
      ),
    kommentar: z
      .string()
      .min(3, { message: 'Kommentar muss mindestens 3 Zeichen lang sein.' })
      .max(200, { message: 'Kommentar darf maximal 200 Zeichen lang sein.' }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: {
      richtung: GeldtransitRichtung.EINLAGE,
      betragEuro: '',
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
        data.richtung,
        parseCents(data.betragEuro),
        data.kommentar,
      )
      toast.success('Kassenbewegung gebucht.')
      form.reset()
      onSuccess()
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Geldtransit buchen</CardTitle>
      </CardHeader>
      <CardContent>
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
              <div className="flex gap-4">
                {(
                  [
                    [GeldtransitRichtung.EINLAGE, 'Einlage (in Tageskasse)'],
                    [GeldtransitRichtung.ENTNAHME, 'Entnahme (aus Tageskasse)'],
                  ] as const
                ).map(([value, label]) => (
                  <label key={value} className="flex items-center gap-1.5">
                    <input
                      type="radio"
                      value={value}
                      {...form.register('richtung')}
                    />
                    {label}
                  </label>
                ))}
              </div>
              {form.formState.errors.richtung && (
                <FieldError errors={[form.formState.errors.richtung]} />
              )}
            </Field>
            <Field
              data-invalid={!!form.formState.errors.betragEuro}
              className="gap-1"
            >
              <FieldLabel htmlFor="bewegung-betrag">Betrag (€)</FieldLabel>
              <Input
                id="bewegung-betrag"
                type="text"
                inputMode="decimal"
                {...form.register('betragEuro')}
                aria-invalid={!!form.formState.errors.betragEuro}
                placeholder="z.B. 25,00"
                className="w-40"
              />
              {form.formState.errors.betragEuro && (
                <FieldError errors={[form.formState.errors.betragEuro]} />
              )}
            </Field>
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

function KassensturzSection({ onSuccess }: { onSuccess: () => void }) {
  const FormDataSchema = z.object({
    istBestandEuro: z
      .string()
      .min(1, { message: 'Bitte einen Betrag eingeben.' })
      .refine((val) => !isNaN(parseFloat(val.replace(',', '.'))), {
        message: 'Bitte einen gültigen Betrag eingeben.',
      }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: { istBestandEuro: '' },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const { loading, run } = useFormActionSubmit({
    form,
    actionLabel: 'Kassensturz durchführen',
  })

  const onSubmit = async (data: FormData) => {
    await run(async () => {
      await kasseBackend.kassensturzDurchfuehren(
        parseCents(data.istBestandEuro),
      )
      toast.success('Kassensturz durchgeführt.')
      form.reset()
      onSuccess()
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensturz</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <Field
              data-invalid={!!form.formState.errors.istBestandEuro}
              className="gap-1"
            >
              <FieldLabel htmlFor="kassensturz-ist">Ist-Bestand (€)</FieldLabel>
              <Input
                id="kassensturz-ist"
                type="text"
                inputMode="decimal"
                {...form.register('istBestandEuro')}
                aria-invalid={!!form.formState.errors.istBestandEuro}
                placeholder="z.B. 342,50"
                className="w-40"
              />
              {form.formState.errors.istBestandEuro && (
                <FieldError errors={[form.formState.errors.istBestandEuro]} />
              )}
            </Field>
            <div>
              <Button type="submit" disabled={loading}>
                Kassensturz durchführen
              </Button>
            </div>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}

function TagesabschlussSection({ onSuccess }: { onSuccess: () => void }) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Tagesabschluss erstellen',
  })

  const handleTagesabschluss = async () => {
    await run(async () => {
      await kasseBackend.tagesabschlussErstellen()
      toast.success('Tagesabschluss erstellt.')
      onSuccess()
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Tagesabschluss</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Schliesst die Kassensitzung ab. Voraussetzungen: Kassensturz
          durchgeführt, alle Tische auf Saldo 0.
        </p>
        <Button
          variant="destructive"
          onClick={() => void handleTagesabschluss()}
          disabled={loading}
        >
          Tagesabschluss erstellen
        </Button>
      </CardContent>
    </Card>
  )
}
