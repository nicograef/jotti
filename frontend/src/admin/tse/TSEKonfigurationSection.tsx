import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'

import { useTSEKonfiguration } from './hooks'
import {
  type TSEKonfigurationSpeichern,
  type TSEVerbindungStatus,
  verbindungIstSigniertfaehig,
} from './TSEBackend'

const emptyTSEKonfiguration: TSEKonfigurationSpeichern = {
  apiKey: '',
  apiSecret: '',
  tssId: '',
  clientId: '',
}

function TSEKonfigurationForm({
  initial,
  apiKeyGesetzt,
  apiSecretGesetzt,
  onSave,
  onClear,
  onTestConnection,
}: {
  initial: TSEKonfigurationSpeichern
  apiKeyGesetzt: boolean
  apiSecretGesetzt: boolean
  onSave: (config: TSEKonfigurationSpeichern) => Promise<void>
  onClear: () => Promise<void>
  onTestConnection: () => Promise<TSEVerbindungStatus>
}) {
  const [form, setForm] = useState<TSEKonfigurationSpeichern>(initial)
  const [verbindungStatus, setVerbindungStatus] =
    useState<TSEVerbindungStatus | null>(null)

  const { loading: saving, run: runSave } = useActionSubmit({
    actionLabel: 'TSE-Konfiguration speichern',
    byCode: {
      validation_error:
        'Bitte alle vier Felder ausfüllen und auf gültige Länge prüfen.',
      tse_konfiguration_kassensitzung_offen:
        'Die TSE-Konfiguration kann nicht geändert werden, solange eine Kassensitzung offen ist. Bitte zuerst den Kassenabschluss durchführen.',
    },
  })
  const { loading: clearing, run: runClear } = useActionSubmit({
    actionLabel: 'TSE-Konfiguration leeren',
    byCode: {
      tse_konfiguration_kassensitzung_offen:
        'Die TSE-Konfiguration kann nicht geleert werden, solange eine Kassensitzung offen ist. Bitte zuerst den Kassenabschluss durchführen.',
    },
  })
  const { loading: testing, run: runTestConnection } = useActionSubmit({
    actionLabel: 'TSE-Verbindung testen',
    byCode: {
      tse_nicht_konfiguriert:
        'Bitte zuerst eine vollständige TSE-Konfiguration speichern.',
      tse_verbindung_fehlgeschlagen:
        'Verbindung zur TSE fehlgeschlagen. Bitte Zugangsdaten und TSS prüfen.',
    },
  })

  const handleSave = async () => {
    await runSave(async () => {
      await onSave(form)
      setForm((prev) => ({
        ...prev,
        apiKey: '',
        apiSecret: '',
      }))
      setVerbindungStatus(null)
      toast.success('TSE-Konfiguration gespeichert.')
    })
  }

  const handleClear = async () => {
    await runClear(async () => {
      await onClear()
      setForm(emptyTSEKonfiguration)
      setVerbindungStatus(null)
      toast.success('TSE-Konfiguration geleert.')
    })
  }

  const handleTestConnection = async () => {
    await runTestConnection(async () => {
      const status = await onTestConnection()
      setVerbindungStatus(status)
      if (verbindungIstSigniertfaehig(status)) {
        toast.success('TSE-Verbindung erfolgreich getestet.')
      } else {
        toast.error('Verbindung steht, aber die TSE ist nicht signierfähig.')
      }
    })
  }

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="tseApiKey">API-Key *</Label>
        <Input
          id="tseApiKey"
          type="password"
          value={form.apiKey}
          onChange={(event) => {
            setForm((prev) => ({ ...prev, apiKey: event.target.value }))
          }}
          placeholder="fiskaly API-Key"
          autoComplete="off"
        />
        {apiKeyGesetzt && (
          <p className="text-sm text-muted-foreground">
            Ein API-Key ist bereits hinterlegt. Beim Speichern wird der hier
            eingegebene Wert übernommen.
          </p>
        )}
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="tseApiSecret">API-Secret *</Label>
        <Input
          id="tseApiSecret"
          type="password"
          value={form.apiSecret}
          onChange={(event) => {
            setForm((prev) => ({
              ...prev,
              apiSecret: event.target.value,
            }))
          }}
          placeholder="fiskaly API-Secret"
          autoComplete="off"
        />
        {apiSecretGesetzt && (
          <p className="text-sm text-muted-foreground">
            Ein API-Secret ist bereits hinterlegt. Beim Speichern wird der hier
            eingegebene Wert übernommen.
          </p>
        )}
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="tseTssId">TSS-ID *</Label>
        <Input
          id="tseTssId"
          value={form.tssId}
          onChange={(event) => {
            setForm((prev) => ({ ...prev, tssId: event.target.value }))
          }}
          placeholder="z.B. 123e4567-e89b-12d3-a456-426614174000"
        />
      </div>

      <div className="grid gap-1.5">
        <Label htmlFor="tseClientId">Client-ID *</Label>
        <Input
          id="tseClientId"
          value={form.clientId}
          onChange={(event) => {
            setForm((prev) => ({ ...prev, clientId: event.target.value }))
          }}
          placeholder="z.B. KASSE-1"
        />
      </div>

      <div className="flex flex-wrap gap-2">
        <Button onClick={() => void handleSave()} disabled={saving || clearing}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
        <Button
          variant="secondary"
          onClick={() => void handleTestConnection()}
          disabled={saving || clearing || testing}
        >
          {testing ? 'Teste Verbindung…' : 'Verbindung testen'}
        </Button>
        <Button
          variant="outline"
          onClick={() => void handleClear()}
          disabled={saving || clearing || testing}
        >
          {clearing ? 'Leeren…' : 'Alle Felder leeren'}
        </Button>
      </div>

      {verbindungStatus ? (
        <VerbindungStatusAnzeige status={verbindungStatus} />
      ) : (
        <p className="text-sm text-muted-foreground">
          Das Ergebnis erscheint nach einem Verbindungstest.
        </p>
      )}
    </div>
  )
}

function VerbindungStatusAnzeige({ status }: { status: TSEVerbindungStatus }) {
  const clientRegistriert = status.clientState === 'REGISTERED'
  return (
    <div className="grid gap-1 text-sm">
      <StatusZeile label="Umgebung" wert={status.umgebung} />
      <StatusZeile label="TSS-Zustand" wert={status.tssState} />
      <StatusZeile
        label="Client-Zustand"
        wert={status.clientState}
        fehler={
          clientRegistriert
            ? undefined
            : 'Der Client ist nicht registriert (REGISTERED). Mit diesem Client kann nicht signiert werden.'
        }
      />
      <StatusZeile
        label="Seriennummern-Abgleich"
        wert={status.seriennummerKorrekt ? 'stimmt überein' : 'weicht ab'}
        fehler={
          status.seriennummerKorrekt
            ? undefined
            : `Die Client-Seriennummer (${status.clientSerialNumber}) stimmt nicht mit der Kassen-Seriennummer überein.`
        }
      />
    </div>
  )
}

function StatusZeile({
  label,
  wert,
  fehler,
}: {
  label: string
  wert: string
  fehler?: string
}) {
  return (
    <div>
      <span className="text-muted-foreground">{label}: </span>
      <strong className={fehler ? 'text-destructive' : undefined}>
        {wert}
      </strong>
      {fehler && <p className="text-destructive">{fehler}</p>}
    </div>
  )
}

export function TSEKonfigurationSection() {
  const {
    tseKonfiguration,
    isPending,
    isLoadingError,
    saveTSEKonfiguration,
    clearTSEKonfiguration,
    testTSEVerbindung,
  } = useTSEKonfiguration()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade TSE-Konfiguration…</p>
    )
  } else if (isLoadingError) {
    // Nur das gescheiterte Erstladen ersetzt das Formular. Ein gescheiterter
    // Hintergrund-Refetch darf es nicht wegreißen: Die eingetippten Zugangsdaten
    // liegen in lokalem State und wären mit dem Unmount verloren.
    inhalt = (
      <p className="text-destructive text-sm">
        Fehler beim Laden der TSE-Konfiguration.
      </p>
    )
  } else {
    const initial: TSEKonfigurationSpeichern = {
      apiKey: '',
      apiSecret: '',
      tssId: tseKonfiguration?.tssId ?? '',
      clientId: tseKonfiguration?.clientId ?? '',
    }

    inhalt = (
      <TSEKonfigurationForm
        key={`${tseKonfiguration?.tssId ?? ''}-${tseKonfiguration?.clientId ?? ''}-${String(tseKonfiguration?.apiKeyGesetzt ?? false)}-${String(tseKonfiguration?.apiSecretGesetzt ?? false)}`}
        initial={initial}
        apiKeyGesetzt={tseKonfiguration?.apiKeyGesetzt ?? false}
        apiSecretGesetzt={tseKonfiguration?.apiSecretGesetzt ?? false}
        onSave={saveTSEKonfiguration}
        onClear={clearTSEKonfiguration}
        onTestConnection={testTSEVerbindung}
      />
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Manuelle Konfiguration (Experten)</CardTitle>
        <CardDescription>
          Fallback für eine bereits außerhalb von jotti angelegte TSE: Hier
          hinterlegst du die vier Zugangsdaten direkt. Die Seriennummer aus der
          Kassenidentität wird in fiskaly als serial_number benötigt.
        </CardDescription>
      </CardHeader>
      <CardContent>{inhalt}</CardContent>
    </Card>
  )
}
