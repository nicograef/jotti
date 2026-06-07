import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  type Betreiber,
  type BondruckEinstellungen,
} from '@/lib/EinstellungenBackend'
import { getActionErrorMessage } from '@/lib/errorMessages'

import {
  useBetreiber,
  useBondruckEinstellungen,
  useKassenidentitaet,
} from './hooks'

function KassenidentitaetSection() {
  const { data: kassenidentitaet, isPending, error } = useKassenidentitaet()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!kassenidentitaet) return
    await navigator.clipboard.writeText(kassenidentitaet.seriennummer)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Kassenidentität</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Seriennummer und Inbetriebnahmedatum identifizieren diese jotti-Instanz
        eindeutig. Beide Angaben werden für die ELSTER-Meldung (§ 146a AO)
        benötigt; die Seriennummer erscheint zusätzlich auf jedem Kassenbeleg.
      </p>
      {isPending && (
        <p className="text-muted-foreground text-sm">Lade Kassenidentität…</p>
      )}
      {error && (
        <p className="text-destructive text-sm">
          Fehler beim Laden der Kassenidentität.
        </p>
      )}
      {kassenidentitaet && (
        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label>Seriennummer</Label>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm">
                {kassenidentitaet.seriennummer}
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
          </div>
          <div className="grid gap-1.5">
            <Label>Inbetriebnahmedatum</Label>
            <p className="text-sm">
              {new Date(kassenidentitaet.angelegtAm).toLocaleDateString(
                'de-DE',
              )}
            </p>
          </div>
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
    } catch (error) {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Betreiber-Stammdaten speichern',
          error,
        }),
      )
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

const emptyBondruckEinstellungen: BondruckEinstellungen = {
  kassenbelegDruckerIp: '',
}

function BondruckEinstellungenForm({
  initial,
  onSave,
}: {
  initial: BondruckEinstellungen
  onSave: (b: BondruckEinstellungen) => Promise<void>
}) {
  const [form, setForm] = useState<BondruckEinstellungen>(initial)
  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      await onSave(form)
      toast.success('Bondruck-Einstellungen gespeichert.')
    } catch (error) {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'Bondruck-Einstellungen speichern',
          error,
          byCode: {
            validation_error: 'Bitte eine gültige IPv4-Adresse eingeben.',
          },
        }),
      )
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="kassenbelegDruckerIp">Kassenbeleg-Drucker-IP</Label>
        <Input
          id="kassenbelegDruckerIp"
          value={form.kassenbelegDruckerIp}
          onChange={(event) => {
            setForm({ kassenbelegDruckerIp: event.target.value })
          }}
          placeholder="z.B. 192.168.1.80"
        />
        {form.kassenbelegDruckerIp === '' && (
          <p className="text-sm text-muted-foreground">kein Drucker</p>
        )}
      </div>
      <div>
        <Button onClick={() => void handleSave()} disabled={saving}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </div>
  )
}

function BondruckEinstellungenSection() {
  const { bondruckEinstellungen, isPending, error, saveBondruckEinstellungen } =
    useBondruckEinstellungen()

  if (isPending) {
    return (
      <section className="max-w-2xl">
        <p className="text-muted-foreground text-sm">
          Lade Bondruck-Einstellungen…
        </p>
      </section>
    )
  }

  if (error) {
    return (
      <section className="max-w-2xl">
        <p className="text-destructive text-sm">
          Fehler beim Laden der Bondruck-Einstellungen.
        </p>
      </section>
    )
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Kassenbeleg-Druck</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Hier konfigurierst du den Drucker für den gesetzlichen Kassenbeleg. Ist
        keine IP gesetzt, ist Belegdruck im Service nicht möglich.
      </p>
      <BondruckEinstellungenForm
        initial={bondruckEinstellungen ?? emptyBondruckEinstellungen}
        onSave={saveBondruckEinstellungen}
      />
    </section>
  )
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
      <KassenidentitaetSection />
      <hr />
      <BondruckEinstellungenSection />
      <hr />
      <BetreiberSection />
    </div>
  )
}
