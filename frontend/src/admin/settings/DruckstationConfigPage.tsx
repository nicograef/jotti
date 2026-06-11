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
import { useActionSubmit } from '@/hooks/use-action-submit'
import {
  type Bonmodus,
  type DruckstationConfig,
  hatBonmodus,
  type Kategorie,
  validateDruckerIp,
} from '@/lib/DruckstationBackend'

import { useDruckstationen } from './hooks'

const KATEGORIE_INFO: Record<
  Kategorie,
  { label: string; beschreibung: string }
> = {
  essen: {
    label: 'Essen',
    beschreibung: 'Arbeitsbons für bestellte Essens-Positionen.',
  },
  getraenk: {
    label: 'Getränk',
    beschreibung: 'Arbeitsbons für bestellte Getränke-Positionen.',
  },
  sonstiges: {
    label: 'Sonstiges',
    beschreibung: 'Arbeitsbons für sonstige Positionen.',
  },
  kassenbeleg: {
    label: 'Kassenbeleg',
    beschreibung: 'Drucker für den Kassenbeleg (Zahlungsbeleg).',
  },
  abholbon: {
    label: 'Abholbon',
    beschreibung:
      'Ist hier ein Drucker konfiguriert, werden beim Direktverkauf Abholbons an diese Station gedruckt (je nach Bonmodus einer pro Position oder ein Sammelbon) — sonst gehen Direktverkäufe an die Produktstationen.',
  },
}

function DruckstationRow({
  config,
  onUpdate,
}: {
  config: DruckstationConfig
  onUpdate: (config: DruckstationConfig) => Promise<void>
}) {
  const info = KATEGORIE_INFO[config.kategorie]
  const zeigtBonmodus = hatBonmodus(config.kategorie)
  const [druckerIp, setDruckerIp] = useState(config.druckerIp)
  const [bonmodus, setBonmodus] = useState<Bonmodus | ''>(config.bonmodus)
  const [ipError, setIpError] = useState<string | null>(null)
  const { loading: saving, run } = useActionSubmit({
    actionLabel: 'Druckstation speichern',
  })

  const handleSave = async () => {
    const fehler = validateDruckerIp(druckerIp)
    if (fehler !== null) {
      setIpError(fehler)
      return
    }
    setIpError(null)

    await run(async () => {
      await onUpdate({ kategorie: config.kategorie, druckerIp, bonmodus })
      toast.success(`Druckstation „${info.label}“ gespeichert.`)
    })
  }

  return (
    <div className="flex flex-col gap-3 py-4 border-b last:border-b-0">
      <div>
        <div className="font-medium">{info.label}</div>
        <p className="text-sm text-muted-foreground">{info.beschreibung}</p>
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <div className="flex flex-col gap-1 min-w-[200px]">
          <Label htmlFor={`ip-${config.kategorie}`}>Drucker-IP</Label>
          <Input
            id={`ip-${config.kategorie}`}
            value={druckerIp}
            onChange={(e) => {
              setDruckerIp(e.target.value)
              if (ipError !== null) {
                setIpError(null)
              }
            }}
            aria-invalid={ipError !== null}
            placeholder="z.B. 192.168.1.50"
          />
          {ipError !== null && (
            <p className="text-sm text-destructive">{ipError}</p>
          )}
          {ipError === null && druckerIp === '' && (
            <p className="text-sm text-muted-foreground">kein Drucker</p>
          )}
        </div>

        {zeigtBonmodus && (
          <div className="flex flex-col gap-1 min-w-[180px]">
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
        )}

        <Button onClick={() => void handleSave()} disabled={saving} size="sm">
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </div>
  )
}

export function DruckstationConfigPage() {
  const { druckstationen, isPending, error, updateDruckstation } =
    useDruckstationen()

  if (isPending) {
    return <p className="text-muted-foreground">Lade Druckstationen…</p>
  }

  if (error) {
    return (
      <p className="text-destructive">Fehler beim Laden der Druckstationen.</p>
    )
  }

  return (
    <div className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-4">Druckstationen</h2>
      <p className="text-muted-foreground text-sm mb-6">
        Jede Druckstation kann einen Drucker im lokalen Netzwerk zugewiesen
        bekommen. Eine leere IP bedeutet, dass für diese Station nicht gedruckt
        wird.
      </p>
      <div className="rounded-md border px-4">
        {druckstationen.map((config) => (
          <DruckstationRow
            key={config.kategorie}
            config={config}
            onUpdate={updateDruckstation}
          />
        ))}
      </div>
    </div>
  )
}
