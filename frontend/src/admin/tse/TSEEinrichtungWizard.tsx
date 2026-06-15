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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import type { TSESetupBefund, TSSBefund } from '@/lib/EinstellungenBackend'

import { pruefeTSESetup } from '../settings/hooks'

export function TSEEinrichtungWizard() {
  const [apiKey, setApiKey] = useState('')
  const [apiSecret, setApiSecret] = useState('')
  const [befund, setBefund] = useState<TSESetupBefund | null>(null)

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

  return (
    <Card>
      <CardHeader>
        <CardTitle>Geführte Einrichtung</CardTitle>
        <CardDescription>
          jotti prüft dein fiskaly-Konto und bereitet die TSE-Einrichtung vor.
          Du brauchst nur den API-Key und das API-Secret aus dem
          fiskaly-Dashboard.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {befund ? (
          <BefundSchritt
            befund={befund}
            onZuruck={() => {
              setBefund(null)
            }}
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
  befund,
  onZuruck,
}: {
  befund: TSESetupBefund
  onZuruck: () => void
}) {
  return (
    <div className="grid gap-4">
      <UmgebungAnzeige umgebung={befund.umgebung} />
      <TSSListe tssListe={befund.vorhandeneTss} />
      <Button variant="outline" className="w-fit" onClick={onZuruck}>
        Andere Zugangsdaten
      </Button>
    </div>
  )
}

function UmgebungAnzeige({ umgebung }: { umgebung: 'TEST' | 'LIVE' }) {
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
