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
import { BackendSingleton } from '@/lib/Backend'
import {
  type Bonmodus,
  DruckerBackend,
  type DruckerKonfig,
  UpdateDruckerConfigSchema,
} from '@/lib/DruckerBackend'
import { useFetch } from '@/lib/useFetch'

const druckerBackend = new DruckerBackend(BackendSingleton)

const KATEGORIE_LABELS: Record<string, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
}

function DruckerRow({
  konfig,
  onSaved,
}: {
  konfig: DruckerKonfig
  onSaved: (updated: DruckerKonfig) => void
}) {
  const [druckerIp, setDruckerIp] = useState(konfig.druckerIp)
  const [bonmodus, setBonmodus] = useState<Bonmodus>(konfig.bonmodus)
  const [saving, setSaving] = useState(false)
  const [ipError, setIpError] = useState<string | null>(null)

  const handleSave = async () => {
    const result = UpdateDruckerConfigSchema.safeParse({
      kategorie: konfig.kategorie,
      druckerIp,
      bonmodus,
    })
    if (!result.success) {
      const ipIssue = result.error.issues.find((i) => i.path[0] === 'druckerIp')
      setIpError(ipIssue?.message ?? 'Ungültige Eingabe')
      return
    }
    setIpError(null)
    setSaving(true)
    try {
      await druckerBackend.updateDruckerConfig(result.data)
      onSaved({ kategorie: konfig.kategorie, druckerIp, bonmodus })
      toast.success(
        `Drucker für „${KATEGORIE_LABELS[konfig.kategorie]}" gespeichert.`,
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
        {KATEGORIE_LABELS[konfig.kategorie]}
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor={`ip-${konfig.kategorie}`}>Drucker-IP</Label>
        <Input
          id={`ip-${konfig.kategorie}`}
          value={druckerIp}
          onChange={(e) => {
            setDruckerIp(e.target.value)
            setIpError(null)
          }}
          placeholder="z.B. 192.168.1.50"
          className={ipError ? 'border-destructive' : ''}
        />
        {druckerIp === '' && !ipError && (
          <p className="text-sm text-muted-foreground">kein Drucker</p>
        )}
        {ipError && <p className="text-sm text-destructive">{ipError}</p>}
      </div>

      <div className="flex flex-col gap-1">
        <Label htmlFor={`bonmodus-${konfig.kategorie}`}>Bonmodus</Label>
        <Select
          value={bonmodus}
          onValueChange={(v) => {
            setBonmodus(v as Bonmodus)
          }}
        >
          <SelectTrigger id={`bonmodus-${konfig.kategorie}`}>
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
  const {
    loading,
    data: drucker,
    setData: setDrucker,
    error,
  } = useFetch(() => druckerBackend.getDruckerConfig(), [] as DruckerKonfig[])

  const handleSaved = (updated: DruckerKonfig) => {
    setDrucker((prev) =>
      prev.map((d) => (d.kategorie === updated.kategorie ? updated : d)),
    )
  }

  if (loading) {
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
        {drucker.map((konfig) => (
          <DruckerRow
            key={konfig.kategorie}
            konfig={konfig}
            onSaved={handleSaved}
          />
        ))}
      </div>
    </div>
  )
}
