import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { useLiveReporting } from '@/admin/reporting/hooks'
import { useTSEKonfiguration } from '@/admin/settings/hooks'
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
import { Badge } from '@/components/ui/badge'
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
import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'
import { formatCents } from '@/lib/utils'

import { kasseBackend, useKassenbestand, useOffeneKassensitzung } from './hooks'
import {
  type KassenabschlussErgebnis,
  SignaturenAusstehendDetailsSchema,
} from './KasseBackend'
import {
  BetragCentsSchema,
  BezeichnungSchema,
  GeldtransitRichtung,
  GeldtransitRichtungSchema,
} from './Kassensitzung'

export function KassensitzungPage() {
  const { kassensitzung, isPending, refetch } = useOffeneKassensitzung()

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
            </CardContent>
          </Card>

          <GeldtransitSection onSuccess={() => void refetch()} />
          <KasseAbschliessenSection
            kassensitzungNr={kassensitzung.zNr}
            onSuccess={() => void refetch()}
          />
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

function GeldtransitSection({ onSuccess }: { onSuccess: () => void }) {
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
        data.richtung,
        data.betragCents,
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

// abschlussErfolgMeldung ergänzt die Erfolgsmeldung um die verbliebenen
// Ausfall-Reste: Vorgänge, die die TSE noch nachsigniert, und Vorgänge ohne
// Signatur mangels TSE-Konfiguration (Tag ohne TSE deutlich ausgewiesen).
function abschlussErfolgMeldung(ergebnis: KassenabschlussErgebnis): string {
  const hinweise: string[] = []
  if (ergebnis.ausfallResteAnzahl > 0) {
    hinweise.push(
      ergebnis.ausfallResteAnzahl === 1
        ? '1 Vorgang wird nach Rückkehr der TSE nachsigniert.'
        : `${String(ergebnis.ausfallResteAnzahl)} Vorgänge werden nach Rückkehr der TSE nachsigniert.`,
    )
  }
  if (ergebnis.ohneKonfigurationAnzahl > 0) {
    hinweise.push(
      ergebnis.ohneKonfigurationAnzahl === 1
        ? '1 Vorgang ohne TSE-Signatur (keine TSE konfiguriert).'
        : `${String(ergebnis.ohneKonfigurationAnzahl)} Vorgänge ohne TSE-Signatur (keine TSE konfiguriert).`,
    )
  }
  return hinweise.length > 0
    ? `Kasse abgeschlossen. ${hinweise.join(' ')}`
    : 'Kasse abgeschlossen.'
}

// signaturenAusstehendMeldung erklärt den Gate-Block: Signaturen stehen noch
// aus, die TSE holt auf, der Abschluss wird gleich erneut angefordert.
function signaturenAusstehendMeldung(anzahl: number): string {
  const kern =
    anzahl <= 0
      ? 'Es sind noch Vorgänge nicht signiert'
      : anzahl === 1
        ? 'Ein Vorgang ist noch nicht signiert'
        : `${String(anzahl)} Vorgänge sind noch nicht signiert`
  return `Der Abschluss wartet: ${kern}. Die TSE holt gerade auf – bitte gleich erneut abschließen.`
}

export function KasseAbschliessenSection({
  kassensitzungNr,
  onSuccess,
}: {
  kassensitzungNr: number
  onSuccess: () => void
}) {
  const { kassenbestand } = useKassenbestand(kassensitzungNr)
  const { liveData } = useLiveReporting()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [istBestandCents, setIstBestandCents] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)

  const FormDataSchema = z.object({
    istBestandCents: BetragCentsSchema,
  })
  type FormData = z.infer<typeof FormDataSchema>

  const form = useForm<FormData>({
    defaultValues: { istBestandCents: 0 },
    resolver: zodResolver(FormDataSchema),
    mode: 'onTouched',
  })

  const sollBestandCents = kassenbestand?.sollBestandCents ?? null
  // Differenz wie im Kassensturz-Event: Soll − Ist (positiv = Fehlbetrag).
  const differenzCents =
    sollBestandCents === null || istBestandCents === null
      ? null
      : sollBestandCents - istBestandCents

  const onSubmit = (data: FormData) => {
    setIstBestandCents(data.istBestandCents)
    setDialogOpen(true)
  }

  const handleAbschliessen = async () => {
    if (istBestandCents === null) return
    setLoading(true)
    try {
      const ergebnis = await kasseBackend.kasseAbschliessen(istBestandCents)
      toast.success(abschlussErfolgMeldung(ergebnis))
      setDialogOpen(false)
      form.reset()
      setIstBestandCents(null)
      onSuccess()
    } catch (error: unknown) {
      // Das Gate blockiert bei noch ausstehenden Signaturen (409). Der Dialog
      // bleibt offen; derselbe Button fordert den Abschluss erneut an.
      if (
        error instanceof BackendError &&
        error.code === 'signaturen_ausstehend'
      ) {
        const details = SignaturenAusstehendDetailsSchema.safeParse(
          error.details,
        )
        toast.warning(
          signaturenAusstehendMeldung(
            details.success ? details.data.anzahl : 0,
          ),
        )
        return
      }
      console.error(error)
      toast.error(
        getActionErrorMessage({ actionLabel: 'Kasse abschließen', error }),
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kasse abschließen</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4">
          Zähle das Bargeld in der Kasse und trage den gezählten Betrag ein.
          Kleine Differenzen zum Soll-Bestand sind normal. Beim Abschließen
          werden Kassensturz und Tagesabschluss (Z-Bon) in einem Schritt
          gebucht.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
          }}
        >
          <FieldGroup>
            <EuroField
              form={form}
              name="istBestandCents"
              label="Gezählter Ist-Bestand"
              withLabel
              placeholder="z.B. 342,50"
              className="w-44"
            />
            <div>
              <Button type="submit" variant="destructive">
                Kasse abschließen
              </Button>
            </div>
          </FieldGroup>
        </form>

        <AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Kasse abschließen?</AlertDialogTitle>
              <AlertDialogDescription>
                Kassensturz und Tagesabschluss (Z-Bon) werden in einem Schritt
                gebucht. Dieser Vorgang kann nicht rückgängig gemacht werden.
              </AlertDialogDescription>
            </AlertDialogHeader>

            <div className="space-y-4 text-sm">
              <dl className="space-y-1">
                <div className="flex justify-between gap-4">
                  <dt className="text-muted-foreground">Soll-Bestand</dt>
                  <dd>
                    {sollBestandCents !== null
                      ? `${formatCents(sollBestandCents)} €`
                      : '—'}
                  </dd>
                </div>
                <div className="flex justify-between gap-4">
                  <dt className="text-muted-foreground">
                    Ist-Bestand (gezählt)
                  </dt>
                  <dd>
                    {istBestandCents !== null
                      ? `${formatCents(istBestandCents)} €`
                      : '—'}
                  </dd>
                </div>
                <div className="flex justify-between gap-4 font-medium">
                  <dt>Differenz</dt>
                  <dd>
                    {differenzCents !== null
                      ? `${formatCents(differenzCents)} €`
                      : '—'}
                  </dd>
                </div>
              </dl>

              <div>
                <p className="mb-1 font-medium">Z-Bon-Vorschau</p>
                <dl className="space-y-1">
                  <div className="flex justify-between gap-4">
                    <dt className="text-muted-foreground">Umsatz</dt>
                    <dd>
                      {liveData
                        ? `${formatCents(liveData.summary.gesamtUmsatzCents)} €`
                        : '—'}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-4">
                    <dt className="text-muted-foreground">Stornierungen</dt>
                    <dd>
                      {liveData
                        ? `${formatCents(liveData.summary.gesamtStornierungenCents)} €`
                        : '—'}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-4">
                    <dt className="text-muted-foreground">Geldtransit</dt>
                    <dd>
                      {liveData
                        ? `${formatCents(liveData.summary.geldtransitCents)} €`
                        : '—'}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>

            <AlertDialogFooter>
              <AlertDialogCancel disabled={loading}>
                Abbrechen
              </AlertDialogCancel>
              <AlertDialogAction
                variant="destructive"
                disabled={loading}
                onClick={(e) => {
                  e.preventDefault()
                  void handleAbschliessen()
                }}
              >
                Kasse abschließen
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  )
}
