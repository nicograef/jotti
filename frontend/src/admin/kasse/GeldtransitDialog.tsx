import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'

import { EuroField } from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
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
import { useVorgangId } from '@/hooks/use-vorgang-id'

import { GELDTRANSIT_LISTE_KEY, kasseBackend, KASSENBESTAND_KEY } from './hooks'
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
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  richtung: GeldtransitRichtung | null
}) {
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

  // Beim Schließen Bewegungsliste und Kassenbestand nachladen: Ging die Antwort
  // einer Buchung verloren, zeigte die stehengebliebene Liste sie nicht — der
  // Admin hielte die Buchung für gescheitert und bliebe ohne die Information,
  // die ihn vom zweiten Versuch abhält. Deckt Abbrechen, Escape und den
  // Erfolgspfad gleichermaßen ab.
  const queryClient = useQueryClient()
  const schliessen = () => {
    void queryClient.invalidateQueries({ queryKey: [GELDTRANSIT_LISTE_KEY] })
    void queryClient.invalidateQueries({ queryKey: [KASSENBESTAND_KEY] })
    onOpenChange(false)
  }

  // Idempotenz-Schlüssel des Vorgangs, wie an den sechs Aufrufstellen im
  // Service-Pfad. Leer ist hier das Betragsfeld ohne Betrag: EuroInput lässt nur
  // Ziffern und ein Komma zu, und parseCents bildet ein leeres Feld auf 0 ab —
  // `betragCents === 0` ist damit genau der Zustand „kein Betrag eingegeben",
  // und zugleich der einzige, aus dem heraus nicht gebucht werden kann (das
  // Schema verlangt mindestens 1 Cent). Ein Wiederholversuch behält den
  // Schlüssel, auch mit korrigiertem Betrag: Genau diese Abweichung meldet der
  // Server als `vorgang_daten_abweichend`, statt ein zweites Mal zu buchen.
  // useWatch ist memoisierbar (anders als form.watch).
  const betragCents = useWatch({ control: form.control, name: 'betragCents' })
  const geldtransitId = useVorgangId(betragCents === 0)

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
      // Der Erfolgspfad stellt den Leerzustand selbst her: Er allein beendet den
      // Vorgang, und erst der Leerzustand löst die Rotation des Schlüssels aus.
      // Ohne ihn trüge die nächste Bewegung denselben Schlüssel — eine zweite
      // Entnahme über denselben Betrag würde als Duplikat verschluckt.
      // Er ist zugleich die einzige Stelle, die leert: Der Dialog bleibt über
      // Schließen und Wiederöffnen montiert, und ein Reset beim Öffnen führte
      // jeden Vorgang durch den Leerzustand — der Wiederholversuch nach einer
      // verlorenen Antwort (schließen, in der Liste nachsehen, erneut öffnen)
      // bekäme einen neuen Schlüssel und würde zur zweiten Buchung.
      form.reset({ betragCents: 0, kommentar: '' })
      schliessen()
    })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(naechsteOffen) => {
        if (!naechsteOffen) schliessen()
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{titel}</DialogTitle>
          <DialogDescription>
            {istEinlage
              ? 'Zusätzliches Bargeld in die Kasse legen (z.B. Wechselgeld). Der Soll-Bestand steigt entsprechend.'
              : 'Bargeld aus der Kasse nehmen (z.B. Abschöpfung in den Tresor). Der Soll-Bestand sinkt entsprechend.'}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <form
            id="geldtransit-form"
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
          </form>
        </DialogBody>
        <DialogFooter className="mt-4">
          <Button
            type="button"
            variant="outline"
            disabled={loading}
            onClick={schliessen}
          >
            Abbrechen
          </Button>
          <Button type="submit" form="geldtransit-form" disabled={loading}>
            {titel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
