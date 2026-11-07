import { Check } from 'lucide-react'
import { type ReactNode, useState } from 'react'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'

import { type Bonmodus, validateDruckerIp } from './DruckstationBackend'

// Welche Kacheln ein Drucker zeigt, sagt `optionen` des Aufrufers.
const BONMODUS_OPTIONEN: { wert: Bonmodus; titel: string; hinweis: string }[] =
  [
    {
      wert: 'pro_position',
      titel: 'Pro Position',
      hinweis: 'je Gericht ein Abreiß-Bon',
    },
    {
      wert: 'pro_bestellung',
      titel: 'Pro Bestellung',
      hinweis: 'ein Sammelbon',
    },
    {
      wert: 'pro_stueck',
      titel: 'Pro Stück',
      hinweis: 'je Einheit ein Bon',
    },
    {
      wert: 'kein_bon',
      titel: 'Kein Bon',
      hinweis: 'Direktverkauf druckt nichts',
    },
  ]

export function BonmodusOptionen({
  optionen,
  aktiv,
  disabled,
  onSelect,
}: {
  optionen: readonly Bonmodus[]
  aktiv: Bonmodus | ''
  disabled: boolean
  onSelect: (bonmodus: Bonmodus) => void
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-muted-foreground">
        Wie sollen Bons gedruckt werden?
      </Label>
      <div className="grid grid-cols-2 gap-2">
        {BONMODUS_OPTIONEN.filter((option) =>
          optionen.includes(option.wert),
        ).map((option) => {
          const gewaehlt = aktiv === option.wert
          return (
            <button
              key={option.wert}
              type="button"
              disabled={disabled}
              aria-pressed={gewaehlt}
              onClick={() => {
                onSelect(option.wert)
              }}
              className={cn(
                'rounded-lg border p-3 text-left transition-colors disabled:opacity-60',
                gewaehlt ? 'border-primary bg-primary/5' : 'hover:bg-accent/50',
              )}
            >
              <div className="text-sm font-semibold">{option.titel}</div>
              <div className="text-xs text-muted-foreground">
                {option.hinweis}
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}

// Speichert nur bei Änderung, beim Verlassen des Felds und per Enter (kein
// Speichern-Button). `onSpeichern` meldet, ob das Speichern gelang.
export function DruckerIpFeld({
  id,
  gespeicherteIp,
  disabled,
  onSpeichern,
  aktion,
}: {
  id: string
  gespeicherteIp: string
  disabled: boolean
  onSpeichern: (druckerIp: string) => Promise<boolean>
  aktion: ReactNode
}) {
  const [druckerIp, setDruckerIp] = useState(gespeicherteIp)
  const [ipError, setIpError] = useState<string | null>(null)
  const [ipGespeichert, setIpGespeichert] = useState(false)

  const speichereIp = async () => {
    if (druckerIp === gespeicherteIp) {
      setIpError(null)
      return
    }
    const fehler = validateDruckerIp(druckerIp)
    if (fehler !== null) {
      setIpError(fehler)
      return
    }
    setIpError(null)
    if (await onSpeichern(druckerIp)) {
      setIpGespeichert(true)
      setTimeout(() => {
        setIpGespeichert(false)
      }, 2000)
    }
  }

  return (
    <>
      <div className="flex items-end gap-2.5">
        <div className="flex flex-1 flex-col gap-1">
          <Label htmlFor={id} className="text-muted-foreground">
            Drucker-IP
            {ipGespeichert && (
              <span className="flex items-center gap-1 text-primary">
                <Check className="size-3.5" aria-hidden />
                Gespeichert
              </span>
            )}
          </Label>
          <Input
            id={id}
            value={druckerIp}
            disabled={disabled}
            onChange={(e) => {
              setDruckerIp(e.target.value)
              if (ipError !== null) {
                setIpError(null)
              }
              if (ipGespeichert) {
                setIpGespeichert(false)
              }
            }}
            onBlur={() => void speichereIp()}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.currentTarget.blur()
              }
            }}
            aria-invalid={ipError !== null}
            placeholder="z.B. 192.168.1.50"
          />
        </div>
        {aktion}
      </div>
      {ipError !== null && (
        <p className="text-sm text-destructive">{ipError}</p>
      )}
    </>
  )
}
