import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { useLiveReporting } from '@/admin/reporting/hooks'
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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { FieldGroup } from '@/components/ui/field'
import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'
import { formatCents } from '@/lib/utils'

import { kasseBackend, useKassenbestand } from './hooks'
import {
  type KassenabschlussErgebnis,
  SignaturenAusstehendDetailsSchema,
} from './KasseBackend'
import { BetragCentsSchema } from './Kassensitzung'

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
