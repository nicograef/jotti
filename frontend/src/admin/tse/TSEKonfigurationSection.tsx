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
import {
  type TSEKonfigurationSpeichern,
  type TSEVerbindungStatus,
} from '@/lib/EinstellungenBackend'

import { useTSEKonfiguration } from '../settings/hooks'

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
    },
  })
  const { loading: clearing, run: runClear } = useActionSubmit({
    actionLabel: 'TSE-Konfiguration leeren',
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
      toast.success('TSE-Verbindung erfolgreich getestet.')
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

      <div className="flex gap-2">
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
        <p className="text-sm text-muted-foreground">
          Verbunden mit <strong>{verbindungStatus.umgebung}</strong> ·
          TSS-Status: {verbindungStatus.tssState}
        </p>
      ) : (
        <p className="text-sm text-muted-foreground">
          Umgebung wird nach einem erfolgreichen Verbindungstest angezeigt.
        </p>
      )}
    </div>
  )
}

export function TSEKonfigurationSection() {
  const {
    tseKonfiguration,
    isPending,
    error,
    saveTSEKonfiguration,
    clearTSEKonfiguration,
    testTSEVerbindung,
  } = useTSEKonfiguration()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade TSE-Konfiguration…</p>
    )
  } else if (error) {
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
        <CardTitle>TSE-Integration (fiskaly)</CardTitle>
        <CardDescription>
          Hier hinterlegst du die Zugangsdaten für deine Cloud-TSE. Die
          Seriennummer aus der Kassenidentität wird in fiskaly als serial_number
          benötigt.
        </CardDescription>
      </CardHeader>
      <CardContent>{inhalt}</CardContent>
    </Card>
  )
}
