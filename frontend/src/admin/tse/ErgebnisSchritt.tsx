import { useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'

import { useTSEKonfiguration } from './hooks'
import {
  type TSEEinrichtenErgebnis,
  type TSEVerbindungStatus,
} from './TSEBackend'

export function ErgebnisSchritt({
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

  // Eine Neu-Anlage liefert PUK und PIN, ein PUK-Reset nur eine neue PIN, eine
  // reine Übernahme keine — dann entfällt die Verwahr-Bestätigung.
  const hatNeuenPuk = ergebnis.puk !== ''
  const hatNeuePin = ergebnis.adminPin !== ''
  const hatNeueGeheimnisse = hatNeuenPuk || hatNeuePin
  const abschlussFreigegeben = !hatNeueGeheimnisse || verwahrt

  // Die neuen Geheimnisse stehen genau einmal auf dem Schirm und sind nirgends
  // gespeichert — ein Reload macht sie unwiederbringlich. Der Haken „verwahrt"
  // gibt den Vorgang bewusst nicht frei, bis der Schritt verlassen wird.
  useOffenerVorgang(hatNeueGeheimnisse)

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
  if (status.signiertfaehig) {
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
