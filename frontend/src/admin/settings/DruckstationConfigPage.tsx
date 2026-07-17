import { Check, Printer } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { cn } from '@/lib/utils'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { WarnKarte } from '../components/WarnKarte'
import {
  beschreibeFehlBons,
  type Bonmodus,
  type DruckstationConfig,
  type FehlgeschlagenerDruckauftrag,
  formatDruckauftragReferenz,
  formatDruckfehler,
  hatBonmodus,
  type Kategorie,
  KATEGORIE_LABEL,
  validateDruckerIp,
} from './DruckstationBackend'
import { useDruckstationen, useFehlgeschlageneDruckauftraege } from './hooks'

// Kurzbeschreibung und Label je Station (Handoff 1g); das Label kommt aus dem
// geteilten KATEGORIE_LABEL des Backends (Single Source of Truth).
const KATEGORIE_INFO: Record<
  Kategorie,
  { label: string; beschreibung: string }
> = {
  essen: {
    label: KATEGORIE_LABEL.essen,
    beschreibung: 'Bons für die Essensausgabe',
  },
  getraenk: {
    label: KATEGORIE_LABEL.getraenk,
    beschreibung: 'Bons für den Ausschank',
  },
  sonstiges: {
    label: KATEGORIE_LABEL.sonstiges,
    beschreibung: 'Bons für sonstige Positionen',
  },
  kassenbeleg: {
    label: KATEGORIE_LABEL.kassenbeleg,
    beschreibung: 'Beleg für Gäste',
  },
  abholbon: {
    label: KATEGORIE_LABEL.abholbon,
    beschreibung: 'Abholnummern beim Direktverkauf',
  },
}

// Die zwei Bonmodus-Optionen mit erklärendem Untertitel (Handoff 1g).
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
  ]

function BonmodusOptionen({
  aktiv,
  disabled,
  onSelect,
}: {
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
        {BONMODUS_OPTIONEN.map((option) => {
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

function DruckstationCard({
  config,
  onUpdate,
  onTestbon,
}: {
  config: DruckstationConfig
  onUpdate: (config: DruckstationConfig) => Promise<void>
  onTestbon: (kategorie: Kategorie) => Promise<void>
}) {
  const info = KATEGORIE_INFO[config.kategorie]
  const zeigtBonmodus = hatBonmodus(config.kategorie)
  const [druckerIp, setDruckerIp] = useState(config.druckerIp)
  const [ipError, setIpError] = useState<string | null>(null)
  const [ipGespeichert, setIpGespeichert] = useState(false)
  const { loading: saving, run: runSave } = useActionSubmit({
    actionLabel: 'Drucker-IP speichern',
  })
  const { loading: testing, run: runTestbon } = useActionSubmit({
    actionLabel: 'Testbon senden',
  })
  const { loading: bonmodusSaving, run: runBonmodus } = useActionSubmit({
    actionLabel: 'Bonmodus speichern',
  })

  // Speichert die Drucker-IP nur, wenn sie sich geändert und die Validierung
  // besteht. Wird on-blur und per Enter ausgelöst (kein Speichern-Button).
  const speichereIp = async () => {
    if (druckerIp === config.druckerIp) {
      setIpError(null)
      return
    }
    const fehler = validateDruckerIp(druckerIp)
    if (fehler !== null) {
      setIpError(fehler)
      return
    }
    setIpError(null)
    await runSave(async () => {
      await onUpdate({ ...config, druckerIp })
      toast.success(`Drucker-IP für „${info.label}“ gespeichert.`)
      // Kurze Inline-Bestätigung am Feld für ~2 Sekunden (Muster der TSE-
      // Kopier-Bestätigung); der Toast bleibt zusätzlich bestehen.
      setIpGespeichert(true)
      setTimeout(() => {
        setIpGespeichert(false)
      }, 2000)
    })
  }

  const waehleBonmodus = async (bonmodus: Bonmodus) => {
    if (bonmodus === config.bonmodus) {
      return
    }
    await runBonmodus(async () => {
      await onUpdate({ ...config, bonmodus })
      toast.success(`Bonmodus für „${info.label}“ geändert.`)
    })
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border p-5">
      <div>
        <span className="text-base font-semibold">{info.label}</span>{' '}
        <span className="text-xs font-normal text-muted-foreground">
          — {info.beschreibung}
        </span>
      </div>

      <div className="flex items-end gap-2.5">
        <div className="flex flex-1 flex-col gap-1">
          <Label
            htmlFor={`ip-${config.kategorie}`}
            className="text-muted-foreground"
          >
            Drucker-IP
            {ipGespeichert && (
              <span className="flex items-center gap-1 text-primary">
                <Check className="size-3.5" aria-hidden />
                Gespeichert
              </span>
            )}
          </Label>
          <Input
            id={`ip-${config.kategorie}`}
            value={druckerIp}
            disabled={saving}
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
        <Button
          variant="outline"
          size="sm"
          className="h-9"
          disabled={testing || config.druckerIp === ''}
          onClick={() =>
            void runTestbon(async () => {
              await onTestbon(config.kategorie)
              toast.success(`Testbon an „${info.label}“ gesendet.`)
            })
          }
        >
          <Printer className="size-4" />
          Testbon
        </Button>
      </div>
      {ipError !== null && (
        <p className="text-sm text-destructive">{ipError}</p>
      )}

      {zeigtBonmodus && (
        <BonmodusOptionen
          aktiv={config.bonmodus}
          disabled={bonmodusSaving}
          onSelect={(bonmodus) => void waehleBonmodus(bonmodus)}
        />
      )}
    </div>
  )
}

function NichtKonfigurierteKarte({
  stationen,
  onUpdate,
  onTestbon,
}: {
  stationen: DruckstationConfig[]
  onUpdate: (config: DruckstationConfig) => Promise<void>
  onTestbon: (kategorie: Kategorie) => Promise<void>
}) {
  const [zuweisen, setZuweisen] = useState(false)

  if (zuweisen) {
    return (
      <>
        {stationen.map((config) => (
          <DruckstationCard
            key={config.kategorie}
            config={config}
            onUpdate={onUpdate}
            onTestbon={onTestbon}
          />
        ))}
      </>
    )
  }

  const namen = stationen
    .map((s) => KATEGORIE_INFO[s.kategorie].label)
    .join(' & ')

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-dashed bg-muted/30 p-5">
      <span className="text-base font-semibold text-muted-foreground">
        {namen} — kein Drucker
      </span>
      <p className="text-sm leading-relaxed text-muted-foreground">
        Diese Stationen drucken nichts. Abholbon aktivieren, wenn Gäste am
        Direktverkauf eine Abholnummer bekommen sollen — sonst gehen
        Direktverkäufe an die Produktstationen.
      </p>
      <Button
        variant="outline"
        size="sm"
        className="w-fit"
        onClick={() => {
          setZuweisen(true)
        }}
      >
        Drucker zuweisen
      </Button>
    </div>
  )
}

function FehlgeschlagenerDruckauftragRow({
  auftrag,
  onErneutVersuchen,
  onVerwerfen,
}: {
  auftrag: FehlgeschlagenerDruckauftrag
  onErneutVersuchen: (id: number) => Promise<void>
  onVerwerfen: (id: number) => Promise<void>
}) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Druckauftrag aktualisieren',
  })

  return (
    <div className="flex flex-wrap items-center gap-3 rounded-lg border bg-background p-3">
      <div className="min-w-[200px] flex-1">
        <div className="text-sm font-semibold" title={auftrag.referenz}>
          {formatDruckauftragReferenz(auftrag.referenz)} → {auftrag.zielIp}
        </div>
        <div className="mt-0.5 text-xs text-muted-foreground">
          {new Date(auftrag.erstelltAm).toLocaleTimeString('de-DE', {
            hour: '2-digit',
            minute: '2-digit',
          })}{' '}
          Uhr · {formatDruckfehler(auftrag.letzterFehler)} · {auftrag.versuche}{' '}
          Versuche
        </div>
      </div>
      <div className="flex gap-2">
        <Button
          size="sm"
          disabled={loading}
          onClick={() =>
            void run(async () => {
              await onErneutVersuchen(auftrag.id)
              toast.success('Druckauftrag erneut eingereiht.')
            })
          }
        >
          Nochmal drucken
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={loading}
          onClick={() =>
            void run(async () => {
              await onVerwerfen(auftrag.id)
              toast.success('Druckauftrag verworfen.')
            })
          }
        >
          Verwerfen
        </Button>
      </div>
    </div>
  )
}

function AlleVerwerfenDialog({
  anzahl,
  alleVerwerfen,
}: {
  anzahl: number
  alleVerwerfen: () => Promise<number>
}) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Druckaufträge verwerfen',
  })

  const verwerfenBeschreibung =
    anzahl === 1
      ? '1 fehlgeschlagener Auftrag wird aus der Warteschlange entfernt.'
      : `${String(anzahl)} fehlgeschlagene Aufträge werden aus der Warteschlange entfernt.`

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="text-destructive"
          disabled={loading}
        >
          Alle verwerfen
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Alle fehlgeschlagenen Druckaufträge verwerfen?
          </AlertDialogTitle>
          <AlertDialogDescription>
            {verwerfenBeschreibung} Noch benötigte Bons vorher einzeln über
            „Nochmal drucken“ nachdrucken.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Abbrechen</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive-solid"
            onClick={() =>
              void run(async () => {
                const verworfen = await alleVerwerfen()
                toast.success(
                  verworfen === 1
                    ? '1 Auftrag verworfen.'
                    : `${String(verworfen)} Aufträge verworfen.`,
                )
              })
            }
          >
            Alle verwerfen
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

function AlarmKarte() {
  const { druckauftraege, erneutVersuchen, verwerfen, alleVerwerfen } =
    useFehlgeschlageneDruckauftraege()

  if (druckauftraege.length === 0) {
    return null
  }

  // Der Warntext folgt der tatsächlichen Bon-Art: nur Arbeitsbons landen in der
  // Küche/an der Theke; ein Kassenbeleg (Gäste-Beleg) oder Testbon darf nicht
  // als Küchenproblem beschrieben werden (NEU02).
  const anzahl = druckauftraege.length
  const { singular, plural, kuecheBetroffen } = beschreibeFehlBons(
    druckauftraege.map((auftrag) => auftrag.bonArt),
  )
  const kuecheHinweis = !kuecheBetroffen
    ? ''
    : anzahl === 1
      ? ' — die Küche hat ihn nicht!'
      : ' — die Küche hat sie nicht!'
  const titel =
    anzahl === 1
      ? `1 ${singular} konnte nicht gedruckt werden${kuecheHinweis}`
      : `${String(anzahl)} ${plural} konnten nicht gedruckt werden${kuecheHinweis}`

  return (
    <WarnKarte title={titel} className="mb-6">
      {/* Höhenbegrenzt und scrollbar, damit die Stationskonfiguration darunter
          auch bei vielen Fehl-Bons ohne langes Scrollen erreichbar bleibt.
          Der Cap greift erst über der Schwelle; wenige Einträge bleiben kompakt. */}
      <div className="mt-2 flex max-h-80 flex-col gap-2 overflow-y-auto">
        {druckauftraege.map((auftrag) => (
          <FehlgeschlagenerDruckauftragRow
            key={auftrag.id}
            auftrag={auftrag}
            onErneutVersuchen={erneutVersuchen}
            onVerwerfen={verwerfen}
          />
        ))}
      </div>
      <div className="mt-3 flex items-center justify-between gap-3">
        <span className="text-xs text-muted-foreground">
          Tipp: Erst Drucker prüfen (Strom, Netzwerk, Papier), dann „Nochmal
          drucken“.
        </span>
        {druckauftraege.length > 1 && (
          <AlleVerwerfenDialog
            anzahl={druckauftraege.length}
            alleVerwerfen={alleVerwerfen}
          />
        )}
      </div>
    </WarnKarte>
  )
}

export function DruckstationConfigPage() {
  const {
    druckstationen,
    isPending,
    error,
    updateDruckstation,
    testbonDrucken,
  } = useDruckstationen()

  let inhalt
  if (isPending) {
    inhalt = <p className="text-muted-foreground">Lade Druckstationen…</p>
  } else if (error) {
    inhalt = (
      <p className="text-destructive">Fehler beim Laden der Druckstationen.</p>
    )
  } else {
    const konfiguriert = druckstationen.filter((s) => s.druckerIp !== '')
    const nichtKonfiguriert = druckstationen.filter((s) => s.druckerIp === '')

    inhalt = (
      <>
        <AlarmKarte />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {konfiguriert.map((config) => (
            <DruckstationCard
              key={config.kategorie}
              config={config}
              onUpdate={updateDruckstation}
              onTestbon={testbonDrucken}
            />
          ))}
          {nichtKonfiguriert.length > 0 && (
            <NichtKonfigurierteKarte
              stationen={nichtKonfiguriert}
              onUpdate={updateDruckstation}
              onTestbon={testbonDrucken}
            />
          )}
        </div>
      </>
    )
  }

  return (
    <div className="max-w-4xl">
      <AdminPageHeader
        titel="Bondrucker"
        unterzeile="Jede Station bekommt einen Drucker im WLAN/LAN zugewiesen. Ohne Drucker wird für die Station nichts gedruckt."
      />
      {inhalt}
    </div>
  )
}
