import { useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'

import { useTSEEinrichtung } from './hooks'
import { SETUP_FEHLER, unvollstaendigAnzeigen } from './setupFehler'
import {
  type TSEEinrichtenErgebnis,
  type TSSBefund,
  type Umgebung,
} from './TSEBackend'

export function UebernahmeSchritt({
  apiKey,
  apiSecret,
  umgebung,
  tss,
  onUebernommen,
  onUnvollstaendig,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  tss: TSSBefund
  onUebernommen: (ergebnis: TSEEinrichtenErgebnis) => void
  onUnvollstaendig: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [pin, setPin] = useState('')
  const [pinUnbekannt, setPinUnbekannt] = useState(false)
  const { uebernimmTSE } = useTSEEinrichtung()

  // Die Admin-PIN wird aus den Unterlagen abgetippt, und jeder Fehlversuch
  // zählt auf die fünf, nach denen fiskaly sperrt.
  useOffenerVorgang(pin.trim() !== '')

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
    // Die Sackgasse „PIN unbekannt" wird unten als bleibende Meldung mit
    // Auswegen gezeigt, nicht als flüchtiger Toast.
    onCode: {
      ...unvollstaendigAnzeigen(onUnvollstaendig),
      tse_setup_pin_unbekannt: () => {
        setPinUnbekannt(true)
      },
    },
  })

  const { pinErforderlich, einsatzbereit } = tss
  const pinFehlt = pinErforderlich && pin.trim() === ''

  const handleUebernehmen = async () => {
    setPinUnbekannt(false)
    await run(async () => {
      onUebernommen(
        await uebernimmTSE({
          apiKey,
          apiSecret,
          umgebung,
          tssId: tss.id,
          pin,
        }),
      )
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
            onUnvollstaendig={onUnvollstaendig}
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

// Sichtbarer Ausweg statt Sackgasse, wenn die verwahrte Admin-PIN fehlt: der
// PUK-Reset („Ich habe den Admin-PUK"), der fiskaly-Support und – in TEST – die
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

// Ist die Admin-PIN verloren oder nach fünf Fehlversuchen gesperrt, setzt jotti
// mit dem verwahrten Admin-PUK eine frische PIN und schließt die Übernahme ab —
// ohne neue, kostenpflichtige TSS. Gilt in TEST und LIVE.
function PukReset({
  apiKey,
  apiSecret,
  umgebung,
  tssId,
  onUebernommen,
  onUnvollstaendig,
}: {
  apiKey: string
  apiSecret: string
  umgebung: Umgebung
  tssId: string
  onUebernommen: (ergebnis: TSEEinrichtenErgebnis) => void
  onUnvollstaendig: (ergebnis: TSEEinrichtenErgebnis) => void
}) {
  const [offen, setOffen] = useState(false)
  const [puk, setPuk] = useState('')
  const [pukUnbekannt, setPukUnbekannt] = useState(false)
  const { uebernimmTSE } = useTSEEinrichtung()

  // Der abgetippte Admin-PUK ist meldepflichtig, das bloße Aufklappen nicht.
  useOffenerVorgang(puk.trim() !== '')

  const { loading, run } = useActionSubmit({
    actionLabel: 'PIN zurücksetzen',
    byCode: {
      ...SETUP_FEHLER,
      tse_setup_tss_nicht_gefunden:
        'Diese TSE wurde im Konto nicht mehr gefunden. Bitte das Konto erneut prüfen.',
      tse_einrichtung_fehlgeschlagen:
        'Das Zurücksetzen ist fehlgeschlagen. Bitte später erneut versuchen.',
    },
    // Der falsche PUK bleibt als bleibende Meldung mit Ausweg stehen, nicht
    // als flüchtiger Toast.
    onCode: {
      ...unvollstaendigAnzeigen(onUnvollstaendig),
      tse_setup_puk_unbekannt: () => {
        setPukUnbekannt(true)
      },
    },
  })

  const pukFehlt = puk.trim() === ''

  const handleReset = async () => {
    setPukUnbekannt(false)
    await run(async () => {
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
