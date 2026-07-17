import { zodResolver } from '@hookform/resolvers/zod'
import { Calculator } from 'lucide-react'
import { useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { WarnKarte } from '@/admin/components/WarnKarte'
import { useLiveReporting } from '@/admin/reporting/hooks'
import { EuroField } from '@/components/common/FormFields'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogBody,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { BackendError } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'
import { cn, formatEuro, formatEuroMitVorzeichen } from '@/lib/utils'

import { kasseBackend, useKassenbestand } from './hooks'
import {
  type KassenabschlussErgebnis,
  SignaturenAusstehendDetailsSchema,
} from './KasseBackend'
import { BetragCentsSchema } from './Kassensitzung'
import { ZaehlhilfeDialog } from './ZaehlhilfeDialog'

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

// offeneTischeWarnung fasst die noch offenen Tische zusammen: Anzahl und
// Gesamtbetrag aus dem Live-Reporting. Null, wenn kein Tisch offen ist.
function offeneTischeWarnung(
  anzahl: number,
  saldoCents: number,
): string | null {
  if (anzahl <= 0) return null
  const tischWort = anzahl === 1 ? 'Tisch ist' : 'Tische sind'
  return `${String(anzahl)} ${tischWort} noch offen (${formatEuro(saldoCents)}).`
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
  const [zaehlhilfeOpen, setZaehlhilfeOpen] = useState(false)
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
  // Live-Rechnung: der aktuell eingetippte Ist-Bestand (bei jeder Eingabe neu),
  // nicht erst der beim Absenden festgehaltene Wert. useWatch ist memoisierbar
  // (anders als form.watch).
  const gezaehltCents = useWatch({
    control: form.control,
    name: 'istBestandCents',
  })
  // Anzeige-Differenz als Ist − Soll (Kassenperspektive): negativ = Fehlbetrag
  // (fehlendes Geld, in Rot hervorgehoben), positiv = Überschuss. Das
  // gebuchte Event trägt Soll − Ist (siehe kassensitzung_events.go); hier zählt
  // nur die Anzeige, deren Vorzeichen dem Design-Handoff folgt.
  const liveDifferenzCents =
    sollBestandCents === null ? null : gezaehltCents - sollBestandCents

  // Differenz für die Dialog-Vorschau (aus dem beim Absenden festgehaltenen
  // Ist-Bestand), gleiche Ist − Soll-Perspektive wie die Live-Anzeige.
  const differenzCents =
    sollBestandCents === null || istBestandCents === null
      ? null
      : istBestandCents - sollBestandCents

  const offeneTischeAnzahl = liveData?.offeneTische.length ?? 0
  const warnungText = liveData
    ? offeneTischeWarnung(offeneTischeAnzahl, liveData.offeneSaldiCents)
    : null

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
    <>
      <p className="text-muted-foreground text-sm">
        Bargeld zählen, Betrag eintragen — Kassensturz und Tagesabschluss
        (Z-Bon) werden in einem Schritt gebucht. Das lässt sich nicht rückgängig
        machen.
      </p>

      {warnungText !== null && (
        <WarnKarte className="mt-4">
          <span>
            <strong className="text-destructive">{warnungText}</strong> Erst
            abrechnen lassen, dann abschließen — sonst landen die Beträge als
            offene Posten im Bericht.
          </span>
        </WarnKarte>
      )}

      {/* €-Eingabe, Soll/Gezählt/Differenz und die Bestätigung sitzen in einer
          gemeinsamen Gruppe, damit die Bestätigung direkt neben den Zahlen
          steht, die sie bucht (NEU07). */}
      <form
        onSubmit={(e) => {
          e.preventDefault()
          void form.handleSubmit(onSubmit)()
        }}
        className="mt-4 flex flex-col gap-4 rounded-lg border p-4"
      >
        <div className="flex flex-wrap items-end gap-4">
          <EuroField
            form={form}
            name="istBestandCents"
            label="Gezählter Ist-Bestand"
            withLabel
            placeholder="z.B. 342,50"
            className="w-44"
          />
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setZaehlhilfeOpen(true)
            }}
          >
            <Calculator />
            Zählhilfe öffnen
          </Button>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 border-t pt-4">
          <div className="flex gap-6 text-sm">
            <div className="text-right">
              <div className="text-muted-foreground">Soll</div>
              <div className="text-base font-semibold">
                {sollBestandCents !== null ? formatEuro(sollBestandCents) : '—'}
              </div>
            </div>
            <div className="text-right">
              <div className="text-muted-foreground">Gezählt</div>
              <div className="text-base font-semibold">
                {formatEuro(gezaehltCents)}
              </div>
            </div>
            <div className="text-right">
              <div className="text-muted-foreground">Differenz</div>
              <div
                className={cn(
                  'text-base font-bold',
                  liveDifferenzCents !== null &&
                    liveDifferenzCents < 0 &&
                    'text-destructive',
                )}
              >
                {liveDifferenzCents !== null
                  ? formatEuroMitVorzeichen(liveDifferenzCents)
                  : '—'}
              </div>
            </div>
          </div>
          <Button type="submit" variant="warn">
            Kasse endgültig abschließen…
          </Button>
        </div>

        <span className="text-muted-foreground text-xs">
          Kleine Differenzen sind normal und werden automatisch dokumentiert.
        </span>
      </form>

      <ZaehlhilfeDialog
        open={zaehlhilfeOpen}
        onOpenChange={setZaehlhilfeOpen}
        onUebernehmen={(summeCents) => {
          form.setValue('istBestandCents', summeCents, {
            shouldValidate: true,
            shouldDirty: true,
          })
        }}
      />

      <AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Kasse abschließen?</AlertDialogTitle>
            <AlertDialogDescription>
              Kassensturz und Tagesabschluss (Z-Bon) werden in einem Schritt
              gebucht. Dieser Vorgang kann nicht rückgängig gemacht werden.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogBody className="space-y-4 text-sm">
            <dl className="space-y-1">
              <div className="flex justify-between gap-4">
                <dt className="text-muted-foreground">Soll-Bestand</dt>
                <dd>
                  {sollBestandCents !== null
                    ? formatEuro(sollBestandCents)
                    : '—'}
                </dd>
              </div>
              <div className="flex justify-between gap-4">
                <dt className="text-muted-foreground">Ist-Bestand (gezählt)</dt>
                <dd>
                  {istBestandCents !== null ? formatEuro(istBestandCents) : '—'}
                </dd>
              </div>
              <div className="flex justify-between gap-4 font-medium">
                <dt>Differenz</dt>
                <dd>
                  {differenzCents !== null
                    ? formatEuroMitVorzeichen(differenzCents)
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
                      ? formatEuro(liveData.summary.gesamtUmsatzCents)
                      : '—'}
                  </dd>
                </div>
                <div className="flex justify-between gap-4">
                  <dt className="text-muted-foreground">Stornierungen</dt>
                  <dd>
                    {liveData
                      ? formatEuro(liveData.summary.gesamtStornierungenCents)
                      : '—'}
                  </dd>
                </div>
                <div className="flex justify-between gap-4">
                  <dt className="text-muted-foreground">Geldtransit</dt>
                  <dd>
                    {liveData
                      ? formatEuro(liveData.summary.geldtransitCents)
                      : '—'}
                  </dd>
                </div>
              </dl>
            </div>
          </AlertDialogBody>

          <AlertDialogFooter>
            <AlertDialogCancel disabled={loading}>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              variant="warn"
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
    </>
  )
}
