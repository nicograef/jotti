import { Printer } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatAnzahl, formatUhrzeit } from '@/lib/utils'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { WarnKarte } from '../components/WarnKarte'
import { BonmodusOptionen, DruckerIpFeld } from './DruckerFelder'
import {
  beschreibeFehlBons,
  type Bonmodus,
  type DruckstationConfig,
  type FehlgeschlagenerDruckauftrag,
  type Kategorie,
  STATION_KATEGORIE_LABEL,
} from './DruckstationBackend'
import { useDruckstationen, useFehlgeschlageneDruckauftraege } from './hooks'
import { ZusatzdruckerAbschnitt } from './ZusatzdruckerKarten'

const KATEGORIE_BESCHREIBUNG: Record<Kategorie, string> = {
  essen: 'Bons für die Essensausgabe',
  getraenk: 'Bons für den Ausschank',
  sonstiges: 'Bons für sonstige Positionen',
  kassenbeleg: 'Beleg für Gäste',
  abholbon: 'Abholnummern beim Direktverkauf',
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
  const label = STATION_KATEGORIE_LABEL[config.kategorie]
  const zeigtBonmodus = config.bonmodusOptionen.length > 0
  const { loading: saving, run: runSave } = useActionSubmit({
    actionLabel: 'Drucker-IP speichern',
  })
  const { loading: testing, run: runTestbon } = useActionSubmit({
    actionLabel: 'Testbon senden',
  })
  const { loading: bonmodusSaving, run: runBonmodus } = useActionSubmit({
    actionLabel: 'Bonmodus speichern',
  })
  // Beide Speicherungen senden die ganze Konfiguration aus `config`; die eine
  // sperrt die andere, bis der Refetch den neuen Stand geliefert hat.

  const speichereIp = (druckerIp: string) =>
    runSave(async () => {
      await onUpdate({ ...config, druckerIp })
      toast.success(`Drucker-IP für „${label}“ gespeichert.`)
    })

  const waehleBonmodus = async (bonmodus: Bonmodus) => {
    if (bonmodus === config.bonmodus) {
      return
    }
    await runBonmodus(async () => {
      await onUpdate({ ...config, bonmodus })
      toast.success(`Bonmodus für „${label}“ geändert.`)
    })
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border p-5">
      <div>
        <span className="text-base font-semibold">{label}</span>{' '}
        <span className="text-xs font-normal text-muted-foreground">
          — {KATEGORIE_BESCHREIBUNG[config.kategorie]}
        </span>
      </div>

      <DruckerIpFeld
        id={`ip-${config.kategorie}`}
        gespeicherteIp={config.druckerIp}
        disabled={saving || bonmodusSaving}
        onSpeichern={speichereIp}
        aktion={
          <Button
            variant="outline"
            size="sm"
            className="h-9"
            disabled={testing || config.druckerIp === ''}
            onClick={() =>
              void runTestbon(async () => {
                await onTestbon(config.kategorie)
                toast.success(`Testbon an „${label}“ gesendet.`)
              })
            }
          >
            <Printer className="size-4" />
            Testbon
          </Button>
        }
      />

      {zeigtBonmodus && (
        <BonmodusOptionen
          optionen={config.bonmodusOptionen}
          aktiv={config.bonmodus}
          disabled={bonmodusSaving || saving}
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
    .map((s) => STATION_KATEGORIE_LABEL[s.kategorie])
    .join(' & ')
  const abholbonDabei = stationen.some((s) => s.kategorie === 'abholbon')

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-dashed bg-muted/30 p-5">
      <span className="text-base font-semibold text-muted-foreground">
        {namen} — kein Drucker
      </span>
      <p className="text-sm leading-relaxed text-muted-foreground">
        Diese Stationen drucken nichts.
        {abholbonDabei &&
          ' Abholbon aktivieren, wenn Gäste am Direktverkauf eine Abholnummer bekommen sollen — sonst gehen Direktverkäufe an die Produktstationen. Mit Bonmodus „Kein Bon“ druckt der Direktverkauf gar nichts.'}
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
          {auftrag.referenzAnzeige} → {auftrag.zielIp}
        </div>
        <div className="mt-0.5 text-xs text-muted-foreground">
          {formatUhrzeit(auftrag.erstelltAm)} Uhr · {auftrag.fehlerMeldung} ·{' '}
          {auftrag.versuche} Versuche
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

  const verwerfenBeschreibung = `${formatAnzahl(anzahl, 'fehlgeschlagener Auftrag wird', 'fehlgeschlagene Aufträge werden')} aus der Warteschlange entfernt.`

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
                  `${formatAnzahl(verworfen, 'Auftrag', 'Aufträge')} verworfen.`,
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

  const anzahl = druckauftraege.length
  const { singular, plural, kuecheBetroffen } = beschreibeFehlBons(
    druckauftraege.map((auftrag) => auftrag.bonArt),
  )
  const kuecheHinweis = !kuecheBetroffen
    ? ''
    : anzahl === 1
      ? ' — die Küche hat ihn nicht!'
      : ' — die Küche hat sie nicht!'
  const titel = `${formatAnzahl(anzahl, `${singular} konnte`, `${plural} konnten`)} nicht gedruckt werden${kuecheHinweis}`

  return (
    <WarnKarte title={titel} className="mb-6">
      {/* Höhenbegrenzt, damit die Stationskonfiguration darunter auch bei
          vielen Fehl-Bons erreichbar bleibt. */}
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

// „Kein Bon" am Abholbon ist eine bewusste Einstellung, auch ohne Drucker-IP.
function istKonfiguriert(config: DruckstationConfig): boolean {
  return config.druckerIp !== '' || config.bonmodus === 'kein_bon'
}

export function DruckstationConfigPage() {
  const {
    druckstationen,
    isPending,
    isError,
    refetch,
    updateDruckstation,
    testbonDrucken,
  } = useDruckstationen()

  let inhalt
  if (isPending) {
    inhalt = <p className="text-muted-foreground">Lade Druckstationen…</p>
  } else if (isError) {
    inhalt = (
      <LadefehlerAlert
        titel="Druckstationen konnten nicht geladen werden"
        onErneutVersuchen={() => void refetch()}
      />
    )
  } else {
    const konfiguriert = druckstationen.filter(istKonfiguriert)
    const nichtKonfiguriert = druckstationen.filter((s) => !istKonfiguriert(s))

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
        <ZusatzdruckerAbschnitt />
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
