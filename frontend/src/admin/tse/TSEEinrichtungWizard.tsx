import { useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { BackendError } from '@/lib/Backend'
import type {
  TSEEinrichtenErgebnis,
  TSESetupBefund,
  TSEVerbindungStatus,
  TSSBefund,
} from '@/lib/EinstellungenBackend'

import {
  pruefeTSESetup,
  useTSEEinrichtung,
  useTSEKonfiguration,
} from '../settings/hooks'

type Umgebung = 'TEST' | 'LIVE'

// Zustände, aus denen jotti eine vorhandene TSS übernehmen bzw. die Einrichtung
// wiederaufnehmen kann. Ab UNINITIALIZED ist die Admin-PIN nötig.
const UEBERNEHMBARE_ZUSTAENDE = ['CREATED', 'UNINITIALIZED', 'INITIALIZED']

function istUebernehmbar(tss: TSSBefund): boolean {
  return UEBERNEHMBARE_ZUSTAENDE.includes(tss.state.toUpperCase())
}

function brauchtPin(tss: TSSBefund): boolean {
  return tss.state.toUpperCase() !== 'CREATED'
}

export function TSEEinrichtungWizard() {
  const [apiKey, setApiKey] = useState('')
  const [apiSecret, setApiSecret] = useState('')
  const [befund, setBefund] = useState<TSESetupBefund | null>(null)
  const [ergebnis, setErgebnis] = useState<TSEEinrichtenErgebnis | null>(null)

  const { loading, run } = useActionSubmit({
    actionLabel: 'TSE prüfen',
    byCode: {
      tse_setup_zugangsdaten_ungueltig:
        'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
      tse_verbindung_fehlgeschlagen:
        'Verbindung zu fiskaly fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const handlePruefen = async () => {
    await run(async () => {
      setBefund(await pruefeTSESetup({ apiKey, apiSecret }))
    })
  }

  const zurueckZuZugangsdaten = () => {
    setBefund(null)
    setErgebnis(null)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Geführte Einrichtung</CardTitle>
        <CardDescription>
          jotti prüft dein fiskaly-Konto und richtet die TSE für dich ein. Du
          brauchst nur den API-Key und das API-Secret aus dem fiskaly-Dashboard.
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
  const uebernehmbar = befund.vorhandeneTss.filter(istUebernehmbar)
  const nurDisabledOderLeer = befund.vorhandeneTss.every(
    (tss) => tss.state.toUpperCase() === 'DISABLED',
  )

  return (
    <div className="grid gap-4">
      <UmgebungAnzeige umgebung={befund.umgebung} />
      <TSSListe tssListe={befund.vorhandeneTss} />
      {uebernehmbar.length > 0 ? (
        <div className="grid gap-3">
          <p className="text-sm">
            In diesem Konto ist bereits eine TSS vorhanden. Du kannst sie
            übernehmen, statt eine neue anzulegen.
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
        </div>
      ) : nurDisabledOderLeer ? (
        <BestaetigungSchritt
          apiKey={apiKey}
          apiSecret={apiSecret}
          umgebung={befund.umgebung}
          onEingerichtet={onEingerichtet}
        />
      ) : (
        <Alert variant="destructive">
          <AlertTitle>TSS in nicht übernehmbarem Zustand</AlertTitle>
          <AlertDescription>
            Die vorhandene TSS lässt sich nicht automatisch übernehmen. Bitte
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

function UebernahmeSchritt({
  apiKey,
  apiSecret,
  umgebung,
  tss,
  onUebernommen,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  tss: TSSBefund
  onUebernommen: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [pin, setPin] = useState('')
  const [pinUnbekannt, setPinUnbekannt] = useState(false)
  const { uebernimmTSE } = useTSEEinrichtung()

  const { loading, run } = useActionSubmit({
    actionLabel: 'TSE übernehmen',
    byCode: {
      tse_setup_zugangsdaten_ungueltig:
        'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
      tse_setup_umgebung_abweichung:
        'Die Umgebung der Zugangsdaten hat sich geändert. Bitte das Konto erneut prüfen.',
      tse_setup_tss_nicht_gefunden:
        'Diese TSS wurde im Konto nicht mehr gefunden. Bitte das Konto erneut prüfen.',
      tse_setup_pin_erforderlich:
        'Für diese TSS ist die verwahrte Admin-PIN erforderlich.',
      tse_setup_uebernahme_nicht_moeglich:
        'Diese TSS lässt sich nicht übernehmen. Bitte den fiskaly-Support kontaktieren.',
      tse_einrichtung_fehlgeschlagen:
        'Die Übernahme ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const pinErforderlich = brauchtPin(tss)
  const pinFehlt = pinErforderlich && pin.trim() === ''
  const passenderClient = tss.passenderClient

  const handleUebernehmen = async () => {
    setPinUnbekannt(false)
    await run(async () => {
      try {
        onUebernommen(
          await uebernimmTSE({
            apiKey,
            apiSecret,
            umgebung,
            tssId: tss.id,
            pin,
          }),
        )
      } catch (error) {
        // Die Sackgasse „PIN unbekannt" wird unten als bleibende Meldung mit
        // Auswegen gezeigt, nicht als flüchtiger Toast.
        if (
          error instanceof BackendError &&
          error.code === 'tse_setup_pin_unbekannt'
        ) {
          setPinUnbekannt(true)
          return
        }
        throw error
      }
    })
  }

  return (
    <div className="grid gap-4 rounded-md border p-4">
      <div className="grid gap-1.5">
        <p className="text-sm font-medium">TSS übernehmen</p>
        <p className="text-sm text-muted-foreground">
          {passenderClient
            ? 'Diese Kasse ist bereits als Client registriert. jotti schließt die Einrichtung ab und speichert die Konfiguration.'
            : 'jotti vollendet die Einrichtung dieser TSS und registriert diese Kasse als Client.'}
        </p>
      </div>

      {pinErforderlich && (
        <div className="grid gap-1.5">
          <Label htmlFor={`pin-${tss.id}`}>Admin-PIN</Label>
          <Input
            id={`pin-${tss.id}`}
            type="password"
            value={pin}
            onChange={(event) => {
              setPin(event.target.value)
            }}
            placeholder="Verwahrte Admin-PIN"
            autoComplete="off"
          />
          <p className="text-xs text-muted-foreground">
            Diese TSS ist bereits personalisiert. Gib die bei der ersten
            Einrichtung verwahrte Admin-PIN ein.
          </p>
        </div>
      )}

      {pinUnbekannt && (
        <Alert variant="destructive">
          <AlertTitle>Admin-PIN nicht akzeptiert</AlertTitle>
          <AlertDescription>
            fiskaly hat die eingegebene Admin-PIN abgelehnt. Ohne die korrekte
            PIN lässt sich diese TSS nicht übernehmen. Mögliche Auswege: die
            verwahrte PIN erneut prüfen, den fiskaly-Support kontaktieren oder
            mit anderen Zugangsdaten bewusst eine neue TSS anlegen.
          </AlertDescription>
        </Alert>
      )}

      <Button
        className="w-fit"
        onClick={() => void handleUebernehmen()}
        disabled={loading || pinFehlt}
      >
        {loading ? 'Übernehme…' : 'TSS übernehmen'}
      </Button>
    </div>
  )
}

function BestaetigungSchritt({
  apiKey,
  apiSecret,
  umgebung,
  onEingerichtet,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  onEingerichtet: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [tippBestaetigung, setTippBestaetigung] = useState('')
  const { richteTSEEin } = useTSEEinrichtung()

  const { loading, run } = useActionSubmit({
    actionLabel: 'TSE einrichten',
    byCode: {
      tse_setup_zugangsdaten_ungueltig:
        'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
      tse_setup_umgebung_abweichung:
        'Die Umgebung der Zugangsdaten hat sich geändert. Bitte das Konto erneut prüfen.',
      tse_bereits_eingerichtet:
        'In diesem Konto existiert bereits eine TSS. Es wird keine neue angelegt.',
      tse_einrichtung_fehlgeschlagen:
        'Die Einrichtung ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const istLive = umgebung === 'LIVE'
  const tippFehlt = istLive && tippBestaetigung.trim().toUpperCase() !== 'LIVE'

  const handleEinrichten = async () => {
    await run(async () => {
      onEingerichtet(await richteTSEEin({ apiKey, apiSecret, umgebung }))
    })
  }

  return (
    <div className="grid gap-4 rounded-md border p-4">
      <div className="grid gap-1.5">
        <p className="text-sm font-medium">Einrichtung starten</p>
        <p className="text-sm text-muted-foreground">
          jotti legt jetzt eine neue TSS an, initialisiert sie und registriert
          diese Kasse als Client. Du erhältst danach einmalig den Admin-PUK und
          die Admin-PIN zur Verwahrung.
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

function ErgebnisSchritt({
  ergebnis,
  onFertig,
}: {
  ergebnis: TSEEinrichtenErgebnis
  onFertig: () => void
}) {
  const [verwahrt, setVerwahrt] = useState(false)
  const [status, setStatus] = useState<TSEVerbindungStatus | null>(null)
  const { testTSEVerbindung } = useTSEKonfiguration()

  const { loading, run } = useActionSubmit({
    actionLabel: 'Verbindung testen',
    byCode: {
      tse_nicht_konfiguriert:
        'Die TSE ist nicht konfiguriert. Bitte die Einrichtung erneut starten.',
      tse_verbindung_fehlgeschlagen:
        'Verbindung zur TSE fehlgeschlagen. Bitte später erneut testen.',
    },
  })

  // Bei der Übernahme einer bereits personalisierten TSS liefert das Backend
  // keine neuen Geheimnisse — der Admin nutzt die bei der Ersteinrichtung
  // verwahrten PUK/PIN weiter. Dann entfällt die Verwahr-Bestätigung.
  const hatNeueGeheimnisse = ergebnis.puk !== ''
  const abschlussFreigegeben = !hatNeueGeheimnisse || verwahrt

  const handleAbschluss = async () => {
    await run(async () => {
      setStatus(await testTSEVerbindung())
    })
  }

  return (
    <div className="grid gap-4">
      {hatNeueGeheimnisse ? (
        <Alert>
          <AlertTitle>TSE erfolgreich eingerichtet</AlertTitle>
          <AlertDescription>
            Notiere PUK und PIN jetzt und verwahre sie sicher außerhalb von
            jotti. Sie werden nicht gespeichert und können nicht erneut
            angezeigt werden.
          </AlertDescription>
        </Alert>
      ) : (
        <Alert>
          <AlertTitle>TSE übernommen</AlertTitle>
          <AlertDescription>
            Die vorhandene TSS wurde übernommen und die Konfiguration
            gespeichert. Deine bereits verwahrten Admin-PUK und Admin-PIN gelten
            unverändert weiter.
          </AlertDescription>
        </Alert>
      )}

      {hatNeueGeheimnisse && (
        <>
          <div className="grid gap-3 rounded-md border p-4">
            <Geheimnis label="Admin-PUK" wert={ergebnis.puk} />
            <Geheimnis label="Admin-PIN" wert={ergebnis.adminPin} />
          </div>

          <label className="flex items-start gap-2 text-sm">
            <Checkbox
              checked={verwahrt}
              onCheckedChange={(checked) => {
                setVerwahrt(checked === true)
              }}
              className="mt-0.5"
            />
            <span>Ich habe Admin-PUK und Admin-PIN sicher verwahrt.</span>
          </label>
        </>
      )}

      {status ? (
        <AbschlussTest status={status} />
      ) : (
        <Button
          className="w-fit"
          onClick={() => void handleAbschluss()}
          disabled={!abschlussFreigegeben || loading}
        >
          {loading ? 'Teste Verbindung…' : 'Verbindung testen & abschließen'}
        </Button>
      )}

      {status && (
        <Button variant="outline" className="w-fit" onClick={onFertig}>
          Fertig
        </Button>
      )}
    </div>
  )
}

function AbschlussTest({ status }: { status: TSEVerbindungStatus }) {
  const inOrdnung =
    status.tssState.toUpperCase() === 'INITIALIZED' &&
    status.clientState.toUpperCase() === 'REGISTERED' &&
    status.seriennummerKorrekt

  if (inOrdnung) {
    return (
      <Alert>
        <AlertTitle>Verbindung bestätigt</AlertTitle>
        <AlertDescription>
          Die TSE ist einsatzbereit (Umgebung {status.umgebung}, TSS{' '}
          {status.tssState}, Client {status.clientState}).
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert variant="destructive">
      <AlertTitle>Verbindung mit Auffälligkeiten</AlertTitle>
      <AlertDescription>
        Umgebung {status.umgebung}, TSS {status.tssState}, Client{' '}
        {status.clientState}, Seriennummern-Abgleich{' '}
        {status.seriennummerKorrekt ? 'korrekt' : 'abweichend'}. Bitte die
        Einrichtung unten prüfen.
      </AlertDescription>
    </Alert>
  )
}

function Geheimnis({ label, wert }: { label: string; wert: string }) {
  return (
    <div className="grid gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="font-mono text-base font-semibold break-all select-all">
        {wert}
      </span>
    </div>
  )
}

function UmgebungAnzeige({ umgebung }: { umgebung: Umgebung }) {
  if (umgebung === 'LIVE') {
    return (
      <Alert variant="destructive">
        <AlertTitle className="flex items-center gap-2">
          Umgebung: <Badge variant="destructive">LIVE</Badge>
        </AlertTitle>
        <AlertDescription>
          Dies ist die echte Produktivumgebung. Hier angelegte TSS verursachen
          laufende Kosten und lassen sich nicht löschen.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert>
      <AlertTitle className="flex items-center gap-2">
        Umgebung: <Badge variant="secondary">TEST</Badge>
      </AlertTitle>
      <AlertDescription>
        Test-Umgebung zum Ausprobieren. Hier signierte Belege sind steuerlich
        nicht gültig.
      </AlertDescription>
    </Alert>
  )
}

function TSSListe({ tssListe }: { tssListe: TSSBefund[] }) {
  if (tssListe.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        In diesem Konto wurde noch keine TSS gefunden.
      </p>
    )
  }

  return (
    <div className="grid gap-3">
      <p className="text-sm font-medium">Gefundene TSS</p>
      {tssListe.map((tss) => (
        <div key={tss.id} className="grid gap-1 rounded-md border p-3 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="font-mono text-xs break-all">{tss.id}</span>
            <Badge variant="outline">{tss.state}</Badge>
          </div>
          {tss.passenderClient ? (
            <p className="text-muted-foreground">
              Passender Client vorhanden ({tss.passenderClient.state}).
            </p>
          ) : (
            <p className="text-muted-foreground">
              Kein Client mit dieser Kassen-Seriennummer.
            </p>
          )}
        </div>
      ))}
    </div>
  )
}
