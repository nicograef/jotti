import { zodResolver } from '@hookform/resolvers/zod'
import { Calculator } from 'lucide-react'
import { useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { WarnKarte } from '@/admin/components/WarnKarte'
import { useLiveReporting } from '@/admin/reporting/hooks'
import type { LiveReportingData } from '@/admin/reporting/types'
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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'
import {
  cn,
  formatAnzahl,
  formatEuro,
  formatEuroMitVorzeichen,
} from '@/lib/utils'

import { kasseBackend, useKassenbestand } from './hooks'
import {
  KasseAbschliessenSchema,
  type KassenabschlussErgebnis,
  SignaturenAusstehendDetailsSchema,
} from './KasseBackend'
import { ZaehlhilfeDialog } from './ZaehlhilfeDialog'

function abschlussErfolgMeldung(ergebnis: KassenabschlussErgebnis): string {
  const hinweise: string[] = []
  if (ergebnis.ausfallResteAnzahl > 0) {
    hinweise.push(
      `${formatAnzahl(ergebnis.ausfallResteAnzahl, 'Vorgang wird', 'Vorgänge werden')} nach Rückkehr der TSE nachsigniert.`,
    )
  }
  if (ergebnis.ohneKonfigurationAnzahl > 0) {
    hinweise.push(
      `${formatAnzahl(ergebnis.ohneKonfigurationAnzahl, 'Vorgang', 'Vorgänge')} ohne TSE-Signatur (keine TSE konfiguriert).`,
    )
  }
  return hinweise.length > 0
    ? `Kasse abgeschlossen. ${hinweise.join(' ')}`
    : 'Kasse abgeschlossen.'
}

function signaturenAusstehendMeldung(anzahl: number): string {
  const kern =
    anzahl <= 0
      ? 'Es sind noch Vorgänge nicht signiert'
      : anzahl === 1
        ? 'Ein Vorgang ist noch nicht signiert'
        : `${String(anzahl)} Vorgänge sind noch nicht signiert`
  return `Der Abschluss wartet: ${kern}. Die TSE holt gerade auf – bitte gleich erneut abschließen.`
}

function offeneTischeWarnung(
  anzahl: number,
  saldoCents: number,
): string | null {
  if (anzahl <= 0) return null
  return `${formatAnzahl(anzahl, 'Tisch ist', 'Tische sind')} noch offen (${formatEuro(saldoCents)}).`
}

function euroOderStrich(
  cents: number | null,
  format: (cents: number) => string = formatEuro,
): string {
  return cents === null ? '—' : format(cents)
}

function VergleichsWert({
  label,
  children,
  className,
}: {
  label: string
  children: string
  className?: string
}) {
  return (
    <div className="text-right">
      <div className="text-muted-foreground">{label}</div>
      <div className={cn('text-base', className ?? 'font-semibold')}>
        {children}
      </div>
    </div>
  )
}

// Live-Rechnung neben dem Eingabefeld; negativ = Fehlbetrag in Rot.
function BestandVergleich({
  sollBestandCents,
  gezaehltCents,
  differenzCents,
}: {
  sollBestandCents: number | null
  gezaehltCents: number
  differenzCents: number | null
}) {
  return (
    <div className="flex gap-6 text-sm">
      <VergleichsWert label="Soll">
        {euroOderStrich(sollBestandCents)}
      </VergleichsWert>
      <VergleichsWert label="Gezählt">
        {formatEuro(gezaehltCents)}
      </VergleichsWert>
      <VergleichsWert
        label="Differenz"
        className={cn(
          'font-bold',
          differenzCents !== null && differenzCents < 0 && 'text-destructive',
        )}
      >
        {euroOderStrich(differenzCents, formatEuroMitVorzeichen)}
      </VergleichsWert>
    </div>
  )
}

function Zeile({
  label,
  wert,
  hervorgehoben = false,
}: {
  label: string
  wert: string
  hervorgehoben?: boolean
}) {
  return (
    <div
      className={cn(
        'flex justify-between gap-4',
        hervorgehoben && 'font-medium',
      )}
    >
      <dt className={hervorgehoben ? undefined : 'text-muted-foreground'}>
        {label}
      </dt>
      <dd>{wert}</dd>
    </div>
  )
}

function AbschlussBestaetigungDialog({
  open,
  onOpenChange,
  loading,
  sollBestandCents,
  istBestandCents,
  differenzCents,
  summary,
  onAbschliessen,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  loading: boolean
  sollBestandCents: number | null
  istBestandCents: number | null
  differenzCents: number | null
  summary: LiveReportingData['summary'] | null
  onAbschliessen: () => void
}) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
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
            <Zeile
              label="Soll-Bestand"
              wert={euroOderStrich(sollBestandCents)}
            />
            <Zeile
              label="Ist-Bestand (gezählt)"
              wert={euroOderStrich(istBestandCents)}
            />
            <Zeile
              label="Differenz"
              wert={euroOderStrich(differenzCents, formatEuroMitVorzeichen)}
              hervorgehoben
            />
          </dl>

          <div>
            <p className="mb-1 font-medium">Z-Bon-Vorschau</p>
            <dl className="space-y-1">
              <Zeile
                label="Umsatz"
                wert={euroOderStrich(summary?.gesamtUmsatzCents ?? null)}
              />
              <Zeile
                label="Stornierungen"
                wert={euroOderStrich(summary?.gesamtStornierungenCents ?? null)}
              />
              <Zeile
                label="Geldtransit"
                wert={euroOderStrich(summary?.geldtransitCents ?? null)}
              />
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
              onAbschliessen()
            }}
          >
            Kasse abschließen
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

type FormData = z.infer<typeof KasseAbschliessenSchema>

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

  const form = useForm<FormData>({
    defaultValues: { istBestandCents: 0 },
    resolver: zodResolver(KasseAbschliessenSchema),
    mode: 'onTouched',
  })

  // Eingetippter Ist-Bestand und offene Rückfrage sind offene Vorgänge; den
  // laufenden Abschluss meldet useActionSubmit.
  useOffenerVorgang(form.formState.isDirty || dialogOpen)

  const abgeschlossen = () => {
    setDialogOpen(false)
    form.reset()
    setIstBestandCents(null)
    onSuccess()
  }

  const { loading, run } = useActionSubmit({
    actionLabel: 'Kasse abschließen',
    onCode: {
      // Gate-Block bei noch ausstehenden Signaturen (409): der Dialog bleibt
      // offen, derselbe Button fordert den Abschluss erneut an.
      signaturen_ausstehend: (error) => {
        const details = SignaturenAusstehendDetailsSchema.safeParse(
          error.details,
        )
        toast.warning(
          signaturenAusstehendMeldung(
            details.success ? details.data.anzahl : 0,
          ),
        )
      },
      // Retry nach verlorener Antwort: der erste Versuch hat schon abgeschlossen.
      kasse_nicht_geoeffnet: () => {
        toast.info('Die Kasse ist bereits abgeschlossen.')
        abgeschlossen()
      },
    },
    // Ein abgebrochener Abschluss kann die Sitzung in wird_abgeschlossen
    // zurücklassen; erst der neu geladene Status zeigt das an.
    onError: onSuccess,
  })

  const sollBestandCents = kassenbestand?.sollBestandCents ?? null
  // Live-Rechnung mit dem aktuell eingetippten Wert; useWatch ist memoisierbar
  // (anders als form.watch).
  const gezaehltCents = useWatch({
    control: form.control,
    name: 'istBestandCents',
  })
  // Anzeige-Differenz als Ist − Soll: negativ = Fehlbetrag, positiv = Überschuss.
  // Das gebuchte Event trägt Soll − Ist (`abschlussBestand.differenzCents`); nur das Anzeige-Vorzeichen ist gedreht.
  const liveDifferenzCents =
    sollBestandCents === null ? null : gezaehltCents - sollBestandCents

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
    await run(async () => {
      const ergebnis = await kasseBackend.kasseAbschliessen(istBestandCents)
      toast.success(abschlussErfolgMeldung(ergebnis))
      abgeschlossen()
    })
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
            abrechnen lassen — solange Tische offen sind, lehnt jotti den
            Abschluss ab.
          </span>
        </WarnKarte>
      )}

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
          <BestandVergleich
            sollBestandCents={sollBestandCents}
            gezaehltCents={gezaehltCents}
            differenzCents={liveDifferenzCents}
          />
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

      <AbschlussBestaetigungDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        loading={loading}
        sollBestandCents={sollBestandCents}
        istBestandCents={istBestandCents}
        differenzCents={differenzCents}
        summary={liveData?.summary ?? null}
        onAbschliessen={() => void handleAbschliessen()}
      />
    </>
  )
}
