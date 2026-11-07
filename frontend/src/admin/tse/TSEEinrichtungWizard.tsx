import { useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'

import { ErgebnisSchritt } from './ErgebnisSchritt'
import { checkTSESetup, useTSEEinrichtung } from './hooks'
import { SETUP_FEHLER, ZUGANGSDATEN_FEHLER } from './setupFehler'
import {
  type TSEEinrichtenErgebnis,
  type TSESetupBefund,
  type Umgebung,
} from './TSEBackend'
import { TSSListe, UmgebungAnzeige } from './TSSAnzeige'
import { UebernahmeSchritt } from './UebernahmeSchritt'

export function TSEEinrichtungWizard() {
  const [apiKey, setApiKey] = useState('')
  const [apiSecret, setApiSecret] = useState('')
  const [befund, setBefund] = useState<TSESetupBefund | null>(null)
  const [ergebnis, setErgebnis] = useState<TSEEinrichtenErgebnis | null>(null)

  // Die Zugangsdaten sind aus dem fiskaly-HUB abgetippt und werden
  // nirgends gespeichert; ein Reload verlangt den ganzen Weg dorthin erneut.
  useOffenerVorgang(apiKey.trim() !== '' || apiSecret.trim() !== '')

  const { loading, run } = useActionSubmit({
    actionLabel: 'TSE prüfen',
    byCode: {
      ...ZUGANGSDATEN_FEHLER,
      tse_verbindung_fehlgeschlagen:
        'Verbindung zu fiskaly fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const handlePruefen = async () => {
    await run(async () => {
      setBefund(await checkTSESetup({ apiKey, apiSecret }))
    })
  }

  // Leert die Zugangsdaten mit: blieben sie stehen, meldete der Wizard oben
  // einen offenen Vorgang, den es nicht gibt.
  const zurueckZuZugangsdaten = () => {
    setApiKey('')
    setApiSecret('')
    setBefund(null)
    setErgebnis(null)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Geführte Einrichtung</CardTitle>
        <CardDescription>
          jotti prüft dein fiskaly-Konto und richtet die TSE für dich ein. Du
          brauchst nur den API-Key und das API-Secret aus dem fiskaly-HUB.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {ergebnis ? (
          <ErgebnisSchritt
            ergebnis={ergebnis}
            onFertig={zurueckZuZugangsdaten}
          />
        ) : befund ? (
          <BefundSchritt
            apiKey={apiKey}
            apiSecret={apiSecret}
            befund={befund}
            onEingerichtet={setErgebnis}
            onZuruck={zurueckZuZugangsdaten}
          />
        ) : (
          <ZugangsdatenSchritt
            apiKey={apiKey}
            apiSecret={apiSecret}
            loading={loading}
            onApiKeyChange={setApiKey}
            onApiSecretChange={setApiSecret}
            onPruefen={() => void handlePruefen()}
          />
        )}
      </CardContent>
    </Card>
  )
}

function ZugangsdatenSchritt({
  apiKey,
  apiSecret,
  loading,
  onApiKeyChange,
  onApiSecretChange,
  onPruefen,
}: {
  apiKey: string
  apiSecret: string
  loading: boolean
  onApiKeyChange: (value: string) => void
  onApiSecretChange: (value: string) => void
  onPruefen: () => void
}) {
  const eingabeUnvollstaendig = apiKey.trim() === '' || apiSecret.trim() === ''

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="setupApiKey">API-Key</Label>
        <Input
          id="setupApiKey"
          type="password"
          value={apiKey}
          onChange={(event) => {
            onApiKeyChange(event.target.value)
          }}
          placeholder="fiskaly API-Key"
          autoComplete="off"
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="setupApiSecret">API-Secret</Label>
        <Input
          id="setupApiSecret"
          type="password"
          value={apiSecret}
          onChange={(event) => {
            onApiSecretChange(event.target.value)
          }}
          placeholder="fiskaly API-Secret"
          autoComplete="off"
        />
      </div>
      <Button
        className="w-fit"
        onClick={onPruefen}
        disabled={loading || eingabeUnvollstaendig}
      >
        {loading ? 'Prüfe…' : 'fiskaly-Konto prüfen'}
      </Button>
    </div>
  )
}

function BefundSchritt({
  apiKey,
  apiSecret,
  befund,
  onEingerichtet,
  onZuruck,
}: {
  apiKey: string
  apiSecret: string
  befund: TSESetupBefund
  onEingerichtet: (ergebnis: TSEEinrichtenErgebnis) => void
  onZuruck: () => void
}) {
  // Die „eigene" TSS (diese Kasse ist dort schon angemeldet) zuerst, damit der
  // Admin bei mehreren TSS den naheliegenden Übernahme-Kandidaten oben findet.
  const uebernehmbar = befund.vorhandeneTss
    .filter((tss) => tss.uebernehmbar)
    .sort(
      (a, b) =>
        Number(b.passenderClient !== null) - Number(a.passenderClient !== null),
    )
  const istTest = befund.umgebung === 'TEST'

  return (
    <div className="grid gap-4">
      <UmgebungAnzeige umgebung={befund.umgebung} />
      <TSSListe tssListe={befund.vorhandeneTss} />
      {uebernehmbar.length > 0 ? (
        <div className="grid gap-3">
          <p className="text-sm">
            In diesem Konto ist bereits eine TSE vorhanden. Du kannst sie
            übernehmen.
          </p>
          {uebernehmbar.map((tss) => (
            <UebernahmeSchritt
              key={tss.id}
              apiKey={apiKey}
              apiSecret={apiSecret}
              umgebung={befund.umgebung}
              tss={tss}
              onUebernommen={onEingerichtet}
            />
          ))}
          {istTest && (
            <NeueTseTrotzdemAnlegen
              apiKey={apiKey}
              apiSecret={apiSecret}
              onEingerichtet={onEingerichtet}
            />
          )}
        </div>
      ) : befund.neuanlageMoeglich ? (
        <BestaetigungSchritt
          apiKey={apiKey}
          apiSecret={apiSecret}
          umgebung={befund.umgebung}
          onEingerichtet={onEingerichtet}
        />
      ) : (
        <Alert variant="destructive">
          <AlertTitle>TSE in nicht übernehmbarem Zustand</AlertTitle>
          <AlertDescription>
            Die vorhandene TSE lässt sich nicht automatisch übernehmen. Bitte
            wende dich an den fiskaly-Support oder nutze die manuelle
            Einrichtung unten.
          </AlertDescription>
        </Alert>
      )}
      <Button variant="outline" className="w-fit" onClick={onZuruck}>
        Andere Zugangsdaten
      </Button>
    </div>
  )
}

// Nur in TEST: eine neue, frische Test-TSE trotz vorhandener TSS. Bewusst als
// untergeordnete Sekundäraktion, damit die Übernahme der Normalweg bleibt.
function NeueTseTrotzdemAnlegen({
  apiKey,
  apiSecret,
  onEingerichtet,
}: {
  apiKey: string
  apiSecret: string
  onEingerichtet: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [offen, setOffen] = useState(false)

  return (
    <Collapsible open={offen} onOpenChange={setOffen}>
      <CollapsibleTrigger asChild>
        <Button variant="link" className="h-auto w-fit p-0 text-sm">
          Stattdessen neue TSE anlegen
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2">
        <BestaetigungSchritt
          apiKey={apiKey}
          apiSecret={apiSecret}
          umgebung="TEST"
          neuAnlegenTrotzVorhandener
          onEingerichtet={onEingerichtet}
        />
      </CollapsibleContent>
    </Collapsible>
  )
}

function BestaetigungSchritt({
  apiKey,
  apiSecret,
  umgebung,
  neuAnlegenTrotzVorhandener = false,
  onEingerichtet,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  neuAnlegenTrotzVorhandener?: boolean
  onEingerichtet: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [tippBestaetigung, setTippBestaetigung] = useState('')
  const { richteTSEEin } = useTSEEinrichtung()

  // Wer die Tippbestätigung begonnen hat, steht unmittelbar vor einer Anlage,
  // die in LIVE Kosten verursacht und sich nicht rückgängig machen lässt.
  useOffenerVorgang(tippBestaetigung.trim() !== '')

  const { loading, run } = useActionSubmit({
    actionLabel: 'TSE einrichten',
    byCode: {
      ...SETUP_FEHLER,
      tse_bereits_eingerichtet:
        'In diesem Konto existiert bereits eine TSE. Es wird keine neue angelegt.',
      tse_setup_tss_limit_erreicht:
        'In der Test-Umgebung sind bereits fünf TSE vorhanden (fiskaly-Grenze). Alte werden bei Inaktivität automatisch bereinigt – bitte später erneut versuchen oder eine vorhandene übernehmen.',
      tse_einrichtung_fehlgeschlagen:
        'Die Einrichtung ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const istLive = umgebung === 'LIVE'
  const tippFehlt = istLive && tippBestaetigung.trim().toUpperCase() !== 'LIVE'

  const handleEinrichten = async () => {
    await run(async () => {
      onEingerichtet(
        await richteTSEEin({
          apiKey,
          apiSecret,
          umgebung,
          neuAnlegenTrotzVorhandener,
        }),
      )
    })
  }

  return (
    <div className="grid gap-4 rounded-md border p-4">
      <div className="grid gap-1.5">
        <p className="text-sm font-medium">
          {neuAnlegenTrotzVorhandener
            ? 'Neue Test-TSE anlegen'
            : 'Einrichtung starten'}
        </p>
        <p className="text-sm text-muted-foreground">
          {neuAnlegenTrotzVorhandener
            ? 'Statt eine vorhandene TSE zu übernehmen, legt jotti eine zusätzliche, frische Test-TSE an und meldet diese Kasse dort an. Sinnvoll, wenn die Admin-PIN der vorhandenen TSE nicht mehr vorliegt. Du erhältst danach einmalig den Admin-PUK und die Admin-PIN zur Verwahrung.'
            : 'jotti legt jetzt eine neue TSE an, richtet sie ein und meldet diese Kasse an. Du erhältst danach einmalig den Admin-PUK und die Admin-PIN zur Verwahrung.'}
        </p>
      </div>

      {istLive && (
        <div className="grid gap-1.5">
          <Label htmlFor="liveBestaetigung">
            Zur Bestätigung „LIVE“ eingeben
          </Label>
          <Input
            id="liveBestaetigung"
            value={tippBestaetigung}
            onChange={(event) => {
              setTippBestaetigung(event.target.value)
            }}
            placeholder="LIVE"
            autoComplete="off"
          />
          <p className="text-xs text-muted-foreground">
            Die Anlage in der LIVE-Umgebung verursacht Kosten und lässt sich
            nicht rückgängig machen.
          </p>
        </div>
      )}

      <Button
        className="w-fit"
        variant={istLive ? 'destructive' : 'default'}
        onClick={() => void handleEinrichten()}
        disabled={loading || tippFehlt}
      >
        {loading ? 'Richte ein…' : 'TSE einrichten'}
      </Button>
    </div>
  )
}
