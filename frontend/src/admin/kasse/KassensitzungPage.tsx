import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
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
import { getActionErrorMessage } from '@/lib/errorMessages'
import { formatCents, parseCents } from '@/lib/utils'

import { kasseBackend, useKassenbestand, useOffeneKassensitzung } from './hooks'
import {
  BezeichnungSchema,
  KassenbewegungArt,
  KassenbewegungArtSchema,
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

          <AnfangsbestandSection onSuccess={() => void refetch()} />
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
    datum: z.string().min(1, { message: 'Datum ist erforderlich.' }),
    bezeichnung: BezeichnungSchema,
  })
  type FormData = z.infer<typeof FormDataSchema>

  const [todayStr] = useState(() => new Date().toISOString().slice(0, 10))
  const form = useForm<FormData>({
    defaultValues: {
      datum: todayStr,
      bezeichnung: '',
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    try {
      await kasseBackend.kassensitzungEroeffnen(data.datum, data.bezeichnung)
      toast.success('Kassensitzung eröffnet.')
      onSuccess()
    } catch (error: unknown) {
      toast.error(
        getActionErrorMessage({ actionLabel: 'Kassensitzung eröffnen', error }),
      )
    }
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
              data-invalid={!!form.formState.errors.datum}
              className="gap-1"
            >
              <FieldLabel htmlFor="ks-datum">Datum</FieldLabel>
              <Input
                id="ks-datum"
                type="date"
                {...form.register('datum')}
                aria-invalid={!!form.formState.errors.datum}
                className="w-48"
              />
              {form.formState.errors.datum && (
                <FieldError errors={[form.formState.errors.datum]} />
              )}
            </Field>
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
            <div>
              <Button type="submit" disabled={form.formState.isSubmitting}>
                Kassensitzung eröffnen
              </Button>
            </div>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}

function AnfangsbestandSection({ onSuccess }: { onSuccess: () => void }) {
  const FormDataSchema = z.object({
    betragEuro: z
      .string()
      .min(1, { message: 'Bitte einen Betrag eingeben.' })
      .refine((val) => !isNaN(parseFloat(val.replace(',', '.'))), {
        message: 'Bitte einen gültigen Betrag eingeben.',
      }),
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: { betragEuro: '' },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    try {
      await kasseBackend.anfangsbestandSetzen(parseCents(data.betragEuro))
      toast.success('Anfangsbestand gesetzt.')
      form.reset()
      onSuccess()
    } catch (error: unknown) {
      toast.error(
        getActionErrorMessage({ actionLabel: 'Anfangsbestand setzen', error }),
      )
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Anfangsbestand setzen</CardTitle>
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
              data-invalid={!!form.formState.errors.betragEuro}
              className="gap-1"
            >
              <FieldLabel htmlFor="anfangsbestand">Betrag (€)</FieldLabel>
              <Input
                id="anfangsbestand"
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
              <Button type="submit" disabled={form.formState.isSubmitting}>
                Anfangsbestand setzen
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
    art: KassenbewegungArtSchema,
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
      art: KassenbewegungArt.GELDTRANSIT,
      betragEuro: '',
      kommentar: '',
    },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const onSubmit = async (data: FormData) => {
    try {
      await kasseBackend.kassenbewegungBuchen(
        data.art,
        parseCents(data.betragEuro),
        data.kommentar,
      )
      toast.success('Kassenbewegung gebucht.')
      form.reset()
      onSuccess()
    } catch (error: unknown) {
      toast.error(
        getActionErrorMessage({ actionLabel: 'Kassenbewegung buchen', error }),
      )
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassenbewegung buchen</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <Field data-invalid={!!form.formState.errors.art} className="gap-1">
              <FieldLabel>Art</FieldLabel>
              <div className="flex gap-4">
                {(
                  [
                    [KassenbewegungArt.GELDTRANSIT, 'Geldtransit'],
                    [KassenbewegungArt.PRIVATENTNAHME, 'Privatentnahme'],
                    [KassenbewegungArt.PRIVATEINLAGE, 'Privateinlage'],
                  ] as const
                ).map(([value, label]) => (
                  <label key={value} className="flex items-center gap-1.5">
                    <input
                      type="radio"
                      value={value}
                      {...form.register('art')}
                    />
                    {label}
                  </label>
                ))}
              </div>
              {form.formState.errors.art && (
                <FieldError errors={[form.formState.errors.art]} />
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
              <Button type="submit" disabled={form.formState.isSubmitting}>
                Kassenbewegung buchen
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

  const onSubmit = async (data: FormData) => {
    try {
      await kasseBackend.kassensturzDurchfuehren(
        parseCents(data.istBestandEuro),
      )
      toast.success('Kassensturz durchgeführt.')
      form.reset()
      onSuccess()
    } catch (error: unknown) {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Kassensturz durchführen',
          error,
        }),
      )
    }
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
              <Button type="submit" disabled={form.formState.isSubmitting}>
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
  const [saving, setSaving] = useState(false)

  const handleTagesabschluss = async () => {
    setSaving(true)
    try {
      await kasseBackend.tagesabschlussErstellen()
      toast.success('Tagesabschluss erstellt.')
      onSuccess()
    } catch (error: unknown) {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Tagesabschluss erstellen',
          error,
        }),
      )
    } finally {
      setSaving(false)
    }
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
          disabled={saving}
        >
          Tagesabschluss erstellen
        </Button>
      </CardContent>
    </Card>
  )
}
