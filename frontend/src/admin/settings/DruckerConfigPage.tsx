import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { type Bonmodus, type DruckerConfig } from '@/lib/DruckerBackend'

import { useDruckerConfig } from './hooks'

const KATEGORIE_LABELS: Record<string, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
}

function DruckerRow({
  config,
  onUpdate,
}: {
  config: DruckerConfig
  onUpdate: (config: DruckerConfig) => Promise<void>
}) {
  const [druckerIp, setDruckerIp] = useState(config.druckerIp)
  const [bonmodus, setBonmodus] = useState<Bonmodus>(config.bonmodus)
  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      await onUpdate({ kategorie: config.kategorie, druckerIp, bonmodus })
      toast.success(
        `Drucker für „${KATEGORIE_LABELS[config.kategorie]}" gespeichert.`,
      )
    } catch {
      toast.error('Fehler beim Speichern der Druckerkonfiguration.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="grid grid-cols-[120px_1fr_1fr_auto] items-start gap-4 py-4 border-b last:border-b-0">
      <div className="font-medium pt-2">
        {KATEGORIE_LABELS[config.kategorie]}
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor={`ip-${config.kategorie}`}>Drucker-IP</Label>
        <Input
          id={`ip-${config.kategorie}`}
          value={druckerIp}
          onChange={(e) => {
            setDruckerIp(e.target.value)
          }}
          placeholder="z.B. 192.168.1.50"
        />
        {druckerIp === '' && (
          <p className="text-sm text-muted-foreground">kein Drucker</p>
        )}
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor={`bonmodus-${config.kategorie}`}>Bonmodus</Label>
        <Select
          value={bonmodus}
          onValueChange={(v) => {
            setBonmodus(v as Bonmodus)
          }}
        >
          <SelectTrigger id={`bonmodus-${config.kategorie}`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="pro_position">Pro Position</SelectItem>
            <SelectItem value="pro_bestellung">Pro Bestellung</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="pt-6">
        <Button onClick={() => void handleSave()} disabled={saving} size="sm">
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </div>
  )
}

export function DruckerConfigPage() {
  const { drucker, isPending, error, updateDruckerConfig } = useDruckerConfig()

  if (isPending) {
    return <p className="text-muted-foreground">Lade Druckerkonfiguration…</p>
  }

  if (error) {
    return (
      <p className="text-destructive">
        Fehler beim Laden der Druckerkonfiguration.
      </p>
    )
  }

  return (
    <div className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-4">Druckerkonfiguration</h2>
      <p className="text-muted-foreground text-sm mb-6">
        Jedem Produkttyp kann ein Bondrucker im lokalen Netzwerk zugewiesen
        werden. Leere IP bedeutet kein Drucker für diese Kategorie.
      </p>
      <div className="rounded-md border px-4">
        {drucker.map((config) => (
          <DruckerRow
            key={config.kategorie}
            config={config}
            onUpdate={updateDruckerConfig}
          />
        ))}
      </div>
    </div>
  )
}
