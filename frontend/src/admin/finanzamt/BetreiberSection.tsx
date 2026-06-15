import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { type Betreiber } from '@/lib/EinstellungenBackend'

import { useBetreiber } from '../settings/hooks'

function BetreiberForm({
  initial,
  onSave,
}: {
  initial: Betreiber
  onSave: (b: Betreiber) => Promise<void>
}) {
  const [form, setForm] = useState<Betreiber>(initial)
  const { loading: saving, run } = useActionSubmit({
    actionLabel: 'Betreiber-Stammdaten speichern',
  })

  const handleChange =
    (field: keyof Betreiber) => (e: React.ChangeEvent<HTMLInputElement>) => {
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

const emptyBetreiber: Betreiber = {
  vereinsname: '',
  strasse: '',
  plz: '',
  ort: '',
  steuernummer: null,
  ustId: null,
}

export function BetreiberSection() {
  const { betreiber, isPending, error, saveBetreiber } = useBetreiber()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Betreiber-Stammdaten</CardTitle>
        <CardDescription>
          Name und Adresse des Vereins erscheinen auf jedem Kassenbeleg (§ 6
          KassenSichV). Eine Kassensitzung kann erst eröffnet werden, wenn
          mindestens der Vereinsname gesetzt ist.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isPending && (
          <p className="text-muted-foreground text-sm">Lade Betreiber-Daten…</p>
        )}
        {error && (
          <p className="text-destructive text-sm">
            Fehler beim Laden der Betreiber-Daten.
          </p>
        )}
        {!isPending && !error && (
          <BetreiberForm
            initial={betreiber ?? emptyBetreiber}
            onSave={saveBetreiber}
          />
        )}
      </CardContent>
    </Card>
  )
}
