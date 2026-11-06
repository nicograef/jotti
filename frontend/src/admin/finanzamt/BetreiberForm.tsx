import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'

import { TextField } from '@/components/common/FormFields'
import { Button } from '@/components/ui/button'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import {
  type BetreiberEingabe,
  BetreiberEingabeSchema,
} from './BetreiberBackend'

export function BetreiberForm({
  initial,
  onSave,
}: {
  initial: BetreiberEingabe
  onSave: (b: BetreiberEingabe) => Promise<void>
}) {
  const form = useForm<BetreiberEingabe>({
    defaultValues: initial,
    resolver: zodResolver(BetreiberEingabeSchema),
    mode: 'onTouched',
  })
  const { loading: saving, run } = useFormActionSubmit({
    form,
    actionLabel: 'Betreiber-Stammdaten speichern',
  })

  const onSubmit = async (daten: BetreiberEingabe) => {
    await run(async () => {
      await onSave(daten)
      toast.success('Betreiber-Stammdaten gespeichert.')
    })
  }

  return (
    <form
      className="grid gap-4"
      onSubmit={(e) => {
        e.preventDefault()
        void form.handleSubmit(onSubmit)()
      }}
    >
      <TextField
        form={form}
        name="vereinsname"
        label="Vereinsname *"
        placeholder="z.B. Sportverein Musterstadt e.V."
      />
      <TextField
        form={form}
        name="strasse"
        label="Straße und Hausnummer *"
        placeholder="z.B. Musterstraße 1"
      />
      <div className="grid grid-cols-[120px_1fr] gap-4">
        <TextField form={form} name="plz" label="PLZ *" placeholder="12345" />
        <TextField
          form={form}
          name="ort"
          label="Ort *"
          placeholder="Musterstadt"
        />
      </div>
      <TextField
        form={form}
        name="steuernummer"
        label="Steuernummer (optional)"
        placeholder="z.B. 12/345/67890"
        autoComplete="off"
      />
      <TextField
        form={form}
        name="ustId"
        label="USt-ID (optional)"
        placeholder="z.B. DE123456789"
        autoComplete="off"
      />
      <div>
        <Button type="submit" disabled={saving}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </form>
  )
}
