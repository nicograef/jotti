import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { Betreiber } from '@/lib/EinstellungenBackend'

import { useBetreiber, useSeriennummer } from './hooks'

function SeriennummerSection() {
  const { data: seriennummer, isPending, error } = useSeriennummer()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!seriennummer) return
    await navigator.clipboard.writeText(seriennummer)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Seriennummer der Kasse</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Diese UUID ist die eindeutige Kennung dieser jotti-Instanz. Sie wird für
        die ELSTER-Meldung (§ 146a AO) benötigt und erscheint auf jedem
        Kassenbeleg.
      </p>
      {isPending && (
        <p className="text-muted-foreground text-sm">Lade Seriennummer…</p>
      )}
      {error && (
        <p className="text-destructive text-sm">
          Fehler beim Laden der Seriennummer.
        </p>
      )}
      {seriennummer && (
        <div className="flex items-center gap-2">
          <code className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm">
            {seriennummer}
          </code>
          <Button
            variant="outline"
            size="icon"
            onClick={() => void handleCopy()}
            aria-label="Seriennummer kopieren"
          >
            {copied ? (
              <Check className="h-4 w-4" />
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </Button>
        </div>
      )}
    </section>
  )
}

function BetreiberForm({
  initial,
  onSave,
}: {
  initial: Betreiber
  onSave: (b: Betreiber) => Promise<void>
}) {
  const [form, setForm] = useState<Betreiber>(initial)
  const [saving, setSaving] = useState(false)

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
    setSaving(true)
    try {
      await onSave(form)
      toast.success('Betreiber-Stammdaten gespeichert.')
    } catch {
      toast.error('Fehler beim Speichern der Betreiber-Stammdaten.')
    } finally {
      setSaving(false)
    }
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
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="ustId">USt-ID (optional)</Label>
        <Input
          id="ustId"
          value={form.ustId ?? ''}
          onChange={handleChange('ustId')}
          placeholder="z.B. DE123456789"
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

function BetreiberSection() {
  const { betreiber, isPending, error, saveBetreiber } = useBetreiber()

  if (isPending) {
    return (
      <section className="max-w-2xl">
        <p className="text-muted-foreground text-sm">Lade Betreiber-Daten…</p>
      </section>
    )
  }

  if (error) {
    return (
      <section className="max-w-2xl">
        <p className="text-destructive text-sm">
          Fehler beim Laden der Betreiber-Daten.
        </p>
      </section>
    )
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Betreiber-Stammdaten</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Name und Adresse des Vereins erscheinen auf jedem Kassenbeleg (§ 6
        KassenSichV). Eine Kassensitzung kann erst eröffnet werden, wenn
        mindestens der Vereinsname gesetzt ist.
      </p>
      <BetreiberForm
        initial={betreiber ?? emptyBetreiber}
        onSave={saveBetreiber}
      />
    </section>
  )
}

export function EinstellungenPage() {
  return (
    <div className="flex flex-col gap-10">
      <SeriennummerSection />
      <hr />
      <BetreiberSection />
    </div>
  )
}
