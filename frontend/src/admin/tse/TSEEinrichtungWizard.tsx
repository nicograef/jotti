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
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { BackendError } from '@/lib/Backend'

import { pruefeTSESetup, useTSEEinrichtung, useTSEKonfiguration } from './hooks'
import {
  type TSEEinrichtenErgebnis,
  type TSESetupBefund,
  type TSEVerbindungStatus,
  type TSSBefund,
  verbindungIstSigniertfaehig,
} from './TSEBackend'

type Umgebung = 'TEST' | 'LIVE'

// Zustände, aus denen jotti eine vorhandene TSS übernehmen bzw. die Einrichtung
// wiederaufnehmen kann. Ab UNINITIALIZED ist die Admin-PIN nötig.
const UEBERNEHMBARE_ZUSTAENDE = ['CREATED', 'UNINITIALIZED', 'INITIALIZED']

function istUebernehmbar(tss: TSSBefund): boolean {
  return UEBERNEHMBARE_ZUSTAENDE.includes(tss.state.toUpperCase())
}

// Einsatzbereit ohne Arbeit: eine INITIALIZED TSS mit bereits registriertem
// (REGISTERED) Client braucht keine privilegierte fiskaly-Operation und damit
// keine Admin-PIN (F8). jotti speichert dann nur noch die Konfiguration.
function istEinsatzbereit(tss: TSSBefund): boolean {
  return (
    tss.state.toUpperCase() === 'INITIALIZED' &&
    tss.passenderClient?.state.toUpperCase() === 'REGISTERED'
  )
}

function brauchtPin(tss: TSSBefund): boolean {
  return tss.state.toUpperCase() !== 'CREATED' && !istEinsatzbereit(tss)
}

// Fehlertexte, die mehrere Setup-Schritte teilen — einmal zentral, damit sie
// nicht zwischen den Schritten auseinanderlaufen.
const ZUGANGSDATEN_FEHLER = {
  tse_setup_zugangsdaten_ungueltig:
    'API-Key oder API-Secret ist ungültig. Bitte Zugangsdaten prüfen.',
}
const SETUP_FEHLER = {
  ...ZUGANGSDATEN_FEHLER,
  tse_setup_umgebung_abweichung:
    'Die Umgebung der Zugangsdaten hat sich geändert. Bitte das Konto erneut prüfen.',
  tse_konfiguration_kassensitzung_offen:
    'Die TSE kann nicht geändert werden, solange eine Kassensitzung offen ist. Bitte zuerst den Kassenabschluss durchführen.',
}

export function TSEEinrichtungWizard() {
  const [apiKey, setApiKey] = useState('')
  const [apiSecret, setApiSecret] = useState('')
  const [befund, setBefund] = useState<TSESetupBefund | null>(null)
  const [ergebnis, setErgebnis] = useState<TSEEinrichtenErgebnis | null>(null)

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
  // Die „eigene" TSS (diese Kasse ist dort schon angemeldet) zuerst, damit der
  // Admin bei mehreren TSS den naheliegenden Übernahme-Kandidaten oben findet.
  const uebernehmbar = befund.vorhandeneTss
    .filter(istUebernehmbar)
    .sort(
      (a, b) =>
        Number(b.passenderClient !== null) - Number(a.passenderClient !== null),
    )
  const nurDisabledOderLeer = befund.vorhandeneTss.every(
    (tss) => tss.state.toUpperCase() === 'DISABLED',
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
      ) : nurDisabledOderLeer ? (
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
      ...SETUP_FEHLER,
      tse_setup_tss_nicht_gefunden:
        'Diese TSE wurde im Konto nicht mehr gefunden. Bitte das Konto erneut prüfen.',
      tse_setup_pin_erforderlich:
        'Für diese TSE ist die verwahrte Admin-PIN erforderlich.',
      tse_setup_uebernahme_nicht_moeglich:
        'Diese TSE lässt sich nicht übernehmen. Bitte den fiskaly-Support kontaktieren.',
      tse_einrichtung_fehlgeschlagen:
        'Die Übernahme ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const pinErforderlich = brauchtPin(tss)
  const pinFehlt = pinErforderlich && pin.trim() === ''
  const einsatzbereit = istEinsatzbereit(tss)

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
        <p className="text-sm font-medium">TSE übernehmen</p>
        <p className="text-sm text-muted-foreground">
          {einsatzbereit
            ? 'Diese Kasse ist hier bereits angemeldet und einsatzbereit. jotti speichert nur noch die Konfiguration.'
            : 'jotti schließt die Einrichtung dieser TSE ab und meldet diese Kasse an.'}
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
            Diese TSE wurde bereits eingerichtet. Gib die bei der ersten
            Einrichtung verwahrte Admin-PIN ein. Tippe sie sorgfältig – nach
            fünf Fehlversuchen sperrt fiskaly die PIN; dann hilft nur der
            Admin-PUK.
          </p>
          <PinFehltHinweis umgebung={umgebung} />
          <PukReset
            apiKey={apiKey}
            apiSecret={apiSecret}
            umgebung={umgebung}
            tssId={tss.id}
            onUebernommen={onUebernommen}
          />
        </div>
      )}

      {pinUnbekannt && (
        <Alert variant="destructive">
          <AlertTitle>Admin-PIN nicht akzeptiert</AlertTitle>
          <AlertDescription>
            fiskaly hat die Admin-PIN abgelehnt (falsch oder nach fünf
            Fehlversuchen gesperrt). Mögliche Auswege: die verwahrte PIN erneut
            prüfen, mit dem Admin-PUK über „Ich habe den Admin-PUK“ eine neue
            PIN setzen, den fiskaly-Support kontaktieren
            {umgebung === 'TEST'
              ? ' oder unten über „Stattdessen neue TSE anlegen“ eine frische Test-TSE anlegen.'
              : ' oder mit anderen Zugangsdaten bewusst eine neue TSE anlegen.'}
          </AlertDescription>
        </Alert>
      )}

      <Button
        className="w-fit"
        onClick={() => void handleUebernehmen()}
        disabled={loading || pinFehlt}
      >
        {loading ? 'Übernehme…' : 'TSE übernehmen'}
      </Button>
    </div>
  )
}

// Auch wenn das PIN-Feld leer ist (Button deaktiviert), braucht der Admin einen
// sichtbaren Ausweg statt einer Sackgasse. Dieser aufklappbare Hinweis nennt die
// Wege, wenn die bei der Ersteinrichtung verwahrte Admin-PIN nicht vorliegt: den
// PUK-Reset („Ich habe den Admin-PUK"), den fiskaly-Support und – in TEST – die
// Sekundäraktion „Stattdessen neue TSE anlegen".
function PinFehltHinweis({ umgebung }: { umgebung: Umgebung }) {
  const [offen, setOffen] = useState(false)

  return (
    <Collapsible open={offen} onOpenChange={setOffen}>
      <CollapsibleTrigger asChild>
        <Button variant="link" className="h-auto w-fit p-0 text-sm">
          Ich habe die Admin-PIN nicht
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <Alert className="mt-2">
          <AlertTitle>Ohne Admin-PIN keine Übernahme</AlertTitle>
          <AlertDescription>
            Die Admin-PIN wurde bei der ersten Einrichtung dieser TSE einmalig
            angezeigt und von euch verwahrt. Mögliche Wege: die verwahrte PIN in
            euren Unterlagen suchen, mit dem verwahrten Admin-PUK über „Ich habe
            den Admin-PUK“ eine neue PIN setzen, den fiskaly-Support
            kontaktieren
            {umgebung === 'TEST'
              ? ', oder in dieser Test-Umgebung unten über „Stattdessen neue TSE anlegen“ eine frische TSE anlegen.'
              : ', oder über „Andere Zugangsdaten“ mit einem anderen fiskaly-Konto neu beginnen.'}
          </AlertDescription>
        </Alert>
      </CollapsibleContent>
    </Collapsible>
  )
}

// Wenn die Admin-PIN verloren oder nach fünf Fehlversuchen gesperrt ist, der
// Admin aber den Admin-PUK verwahrt hat, setzt jotti damit eine frische PIN und
// schließt die Übernahme ab – ohne neue, kostenpflichtige TSS. Gilt in TEST und
// LIVE. Erfolg endet wie die übrigen Wege im ErgebnisSchritt (mit einmaliger
// Anzeige der neuen PIN); ein falscher PUK bleibt als Meldung mit Ausweg stehen.
function PukReset({
  apiKey,
  apiSecret,
  umgebung,
  tssId,
  onUebernommen,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  tssId: string
  onUebernommen: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [offen, setOffen] = useState(false)
  const [puk, setPuk] = useState('')
  const [pukUnbekannt, setPukUnbekannt] = useState(false)
  const { uebernimmTSE } = useTSEEinrichtung()

  const { loading, run } = useActionSubmit({
    actionLabel: 'PIN zurücksetzen',
    byCode: {
      ...SETUP_FEHLER,
      tse_setup_tss_nicht_gefunden:
        'Diese TSE wurde im Konto nicht mehr gefunden. Bitte das Konto erneut prüfen.',
      tse_einrichtung_fehlgeschlagen:
        'Das Zurücksetzen ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
  })

  const pukFehlt = puk.trim() === ''

  const handleReset = async () => {
    setPukUnbekannt(false)
    await run(async () => {
      try {
        onUebernommen(
          await uebernimmTSE({
            apiKey,
            apiSecret,
            umgebung,
            tssId,
            pin: '',
            puk,
          }),
        )
      } catch (error) {
        // Der falsche PUK bleibt als bleibende Meldung mit Ausweg stehen, nicht
        // als flüchtiger Toast.
        if (
          error instanceof BackendError &&
          error.code === 'tse_setup_puk_unbekannt'
        ) {
          setPukUnbekannt(true)
          return
        }
        throw error
      }
    })
  }

  return (
    <Collapsible open={offen} onOpenChange={setOffen}>
      <CollapsibleTrigger asChild>
        <Button variant="link" className="h-auto w-fit p-0 text-sm">
          Ich habe den Admin-PUK
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2">
        <div className="grid gap-1.5">
          <p className="text-sm text-muted-foreground">
            Mit dem verwahrten Admin-PUK setzt jotti eine neue Admin-PIN und
            übernimmt die TSE damit – auch wenn die alte PIN verloren oder
            gesperrt ist. Der Admin-PUK bleibt unverändert. Du erhältst danach
            einmalig die neue Admin-PIN zur Verwahrung.
          </p>
          <Label htmlFor={`puk-${tssId}`}>Admin-PUK</Label>
          <Input
            id={`puk-${tssId}`}
            type="password"
            value={puk}
            onChange={(event) => {
              setPuk(event.target.value)
            }}
            placeholder="Verwahrter Admin-PUK"
            autoComplete="off"
          />
          {pukUnbekannt && (
            <Alert variant="destructive">
              <AlertTitle>Admin-PUK nicht akzeptiert</AlertTitle>
              <AlertDescription>
                fiskaly hat den eingegebenen Admin-PUK abgelehnt. Bitte den
                verwahrten PUK genau prüfen. Sind Admin-PUK und Admin-PIN beide
                verloren, hilft nur der fiskaly-Support.
              </AlertDescription>
            </Alert>
          )}
          <Button
            className="w-fit"
            onClick={() => void handleReset()}
            disabled={loading || pukFehlt}
          >
            {loading ? 'Setze zurück…' : 'PIN zurücksetzen und übernehmen'}
          </Button>
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

// In TEST darf der Admin trotz vorhandener (ggf. PIN-loser, nicht übernehmbarer)
// TSS bewusst eine neue, frische Test-TSE anlegen (F2). Klar untergeordnet als
// aufklappbare Sekundäraktion, damit die Übernahme der Normalweg bleibt.
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

  // Welche Geheimnisse neu sind, steuert die Anzeige: eine Neu-Anlage liefert
  // PUK und PIN, ein PUK-Reset nur eine neue PIN (der PUK bleibt unverändert),
  // eine reine Übernahme keine — dann entfällt die Verwahr-Bestätigung.
  const hatNeuenPuk = ergebnis.puk !== ''
  const hatNeuePin = ergebnis.adminPin !== ''
  const hatNeueGeheimnisse = hatNeuenPuk || hatNeuePin
  const abschlussFreigegeben = !hatNeueGeheimnisse || verwahrt

  const handleAbschluss = async () => {
    await run(async () => {
      setStatus(await testTSEVerbindung())
    })
  }

  return (
    <div className="grid gap-4">
      {hatNeuenPuk ? (
        <Alert>
          <AlertTitle>TSE erfolgreich eingerichtet</AlertTitle>
          <AlertDescription>
            Notiere PUK und PIN jetzt und verwahre sie sicher außerhalb von
            jotti. Sie werden nicht gespeichert und können nicht erneut
            angezeigt werden.
          </AlertDescription>
        </Alert>
      ) : hatNeuePin ? (
        <Alert>
          <AlertTitle>Neue Admin-PIN gesetzt</AlertTitle>
          <AlertDescription>
            Notiere die neue Admin-PIN jetzt und verwahre sie sicher außerhalb
            von jotti. Dein bereits verwahrter Admin-PUK bleibt unverändert
            gültig. Die PIN wird nicht gespeichert und kann nicht erneut
            angezeigt werden.
          </AlertDescription>
        </Alert>
      ) : (
        <Alert>
          <AlertTitle>TSE übernommen</AlertTitle>
          <AlertDescription>
            Die vorhandene TSE wurde übernommen und die Konfiguration
            gespeichert. Deine bereits verwahrten Admin-PUK und Admin-PIN gelten
            unverändert weiter.
          </AlertDescription>
        </Alert>
      )}

      {hatNeueGeheimnisse && (
        <>
          <div className="grid gap-3 rounded-md border p-4">
            {hatNeuenPuk && <Geheimnis label="Admin-PUK" wert={ergebnis.puk} />}
            {hatNeuePin && (
              <Geheimnis label="Admin-PIN" wert={ergebnis.adminPin} />
            )}
          </div>

          <label className="flex items-start gap-2 text-sm">
            <Checkbox
              checked={verwahrt}
              onCheckedChange={(checked) => {
                setVerwahrt(checked === true)
              }}
              className="mt-0.5"
            />
            <span>
              {hatNeuenPuk
                ? 'Ich habe Admin-PUK und Admin-PIN sicher verwahrt.'
                : 'Ich habe die neue Admin-PIN sicher verwahrt.'}
            </span>
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
  if (verbindungIstSigniertfaehig(status)) {
    return (
      <Alert>
        <AlertTitle>Verbindung bestätigt</AlertTitle>
        <AlertDescription>
          Die TSE ist einsatzbereit (Umgebung {status.umgebung}, TSE-Zustand{' '}
          {status.tssState}, Kassen-Anmeldung {status.clientState}).
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert variant="destructive">
      <AlertTitle>Verbindung mit Auffälligkeiten</AlertTitle>
      <AlertDescription>
        Umgebung {status.umgebung}, TSE-Zustand {status.tssState},
        Kassen-Anmeldung {status.clientState}, Seriennummern-Abgleich{' '}
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
          Dies ist die echte Produktivumgebung. Hier angelegte TSE verursachen
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

// Übersetzt den technischen fiskaly-Zustand in eine laienverständliche Angabe.
// Der rohe Zustand bleibt daneben als Kennung für den Support sichtbar.
function tssZustandKlartext(state: string): string {
  switch (state.toUpperCase()) {
    case 'CREATED':
      return 'neu angelegt'
    case 'UNINITIALIZED':
      return 'angelegt, noch nicht einsatzbereit'
    case 'INITIALIZED':
      return 'einsatzbereit'
    case 'DISABLED':
      return 'stillgelegt'
    default:
      return state
  }
}

function TSSListe({ tssListe }: { tssListe: TSSBefund[] }) {
  if (tssListe.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        In diesem Konto wurde noch keine TSE gefunden.
      </p>
    )
  }

  return (
    <div className="grid gap-3">
      <p className="text-sm font-medium">Gefundene TSE</p>
      {tssListe.map((tss) => (
        <div key={tss.id} className="grid gap-1 rounded-md border p-3 text-sm">
          <div className="flex items-center justify-between gap-2">
            <Badge variant="outline">{tssZustandKlartext(tss.state)}</Badge>
            <span className="font-mono text-xs text-muted-foreground">
              {tss.state}
            </span>
          </div>
          {tss.passenderClient ? (
            <p className="text-muted-foreground">
              Diese Kasse ist hier bereits angemeldet.
            </p>
          ) : (
            <p className="text-muted-foreground">
              Diese Kasse ist hier noch nicht angemeldet.
            </p>
          )}
          <p className="text-muted-foreground text-xs break-all">
            Technische Kennung: {tss.id}
          </p>
        </div>
      ))}
    </div>
  )
}
