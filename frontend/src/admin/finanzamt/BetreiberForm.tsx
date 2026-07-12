import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { type BetreiberEingabe } from './BetreiberBackend'

// Formular für die Betreiber-Stammdaten (Adresse). Wird in der
// Einrichtungs-Checkliste unter „Bearbeiten" ausgeklappt.
export function BetreiberForm({
  initial,
  onSave,
}: {
  initial: BetreiberEingabe
  onSave: (b: BetreiberEingabe) => Promise<void>
}) {
  const [form, setForm] = useState<BetreiberEingabe>(initial)
  const { loading: saving, run } = useActionSubmit({
    actionLabel: 'Betreiber-Stammdaten speichern',
  })

  const handleChange =
    (field: keyof BetreiberEingabe) =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      setForm((prev) => ({
        ...prev,
        [field]:
          value === '' && (field === 'steuernummer' || field === 'ustId')
            ? null
            : value,
      }))
    }

  const handleSave = async () => {
    await run(async () => {
      await onSave(form)
      toast.success('Betreiber-Stammdaten gespeichert.')
    })
  }

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="vereinsname">Vereinsname *</Label>
        <Input
          id="vereinsname"
          value={form.vereinsname}
          onChange={handleChange('vereinsname')}
          placeholder="z.B. Sportverein Musterstadt e.V."
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="strasse">Straße und Hausnummer *</Label>
        <Input
          id="strasse"
          value={form.strasse}
          onChange={handleChange('strasse')}
          placeholder="z.B. Musterstraße 1"
        />
      </div>
      <div className="grid grid-cols-[120px_1fr] gap-4">
        <div className="grid gap-1.5">
          <Label htmlFor="plz">PLZ *</Label>
          <Input
            id="plz"
            value={form.plz}
            onChange={handleChange('plz')}
            placeholder="12345"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="ort">Ort *</Label>
          <Input
            id="ort"
            value={form.ort}
            onChange={handleChange('ort')}
            placeholder="Musterstadt"
          />
        </div>
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="steuernummer">Steuernummer (optional)</Label>
        <Input
          id="steuernummer"
          value={form.steuernummer ?? ''}
          onChange={handleChange('steuernummer')}
          placeholder="z.B. 12/345/67890"
          autoComplete="off"
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="ustId">USt-ID (optional)</Label>
        <Input
          id="ustId"
          value={form.ustId ?? ''}
          onChange={handleChange('ustId')}
          placeholder="z.B. DE123456789"
          autoComplete="off"
        />
      </div>
      <div>
        <Button onClick={() => void handleSave()} disabled={saving}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </div>
  )
}
