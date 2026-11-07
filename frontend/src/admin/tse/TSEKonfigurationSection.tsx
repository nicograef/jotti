import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'

import { TextField } from '@/components/common/FormFields'
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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useFormActionSubmit } from '@/hooks/use-form-action-submit'

import { useTSEKonfiguration } from './hooks'
import { ZUGANGSDATEN_FEHLER } from './setupFehler'
import {
  type TSEKonfigurationSpeichern,
  TSEKonfigurationSpeichernSchema,
  type TSEVerbindungStatus,
} from './TSEBackend'

// Speichern und Leeren scheitern am selben Backend-Guard: offene Kassensitzung
// oder unterbrochener Abschluss.
const KASSENSITZUNG_AKTIV_GRUND =
  'solange eine Kassensitzung aktiv ist (offen oder Abschluss unterbrochen). Bitte zuerst den Kassenabschluss durchführen bzw. wiederholen.'

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
  const form = useForm<TSEKonfigurationSpeichern>({
    defaultValues: initial,
    resolver: zodResolver(TSEKonfigurationSpeichernSchema),
    mode: 'onTouched',
  })
  const [verbindungStatus, setVerbindungStatus] =
    useState<TSEVerbindungStatus | null>(null)

  const { loading: saving, run: runSave } = useFormActionSubmit({
    form,
    actionLabel: 'TSE-Konfiguration speichern',
    byCode: {
      // Ohne zentrale Meldung.
      tse_konfiguration_kassensitzung_offen: `Die TSE-Konfiguration kann nicht geändert werden, ${KASSENSITZUNG_AKTIV_GRUND}`,
      // Vor dem Speichern liest jotti die Stammdaten der TSS bei fiskaly.
      ...ZUGANGSDATEN_FEHLER,
      tse_verbindung_fehlgeschlagen:
        'Die TSS-Daten konnten bei fiskaly nicht gelesen werden. Bitte TSS-ID prüfen und erneut versuchen; gespeichert wurde nichts.',
    },
  })
  const { loading: clearing, run: runClear } = useActionSubmit({
    actionLabel: 'TSE-Konfiguration leeren',
    byCode: {
      // Ohne zentrale Meldung.
      tse_konfiguration_kassensitzung_offen: `Die TSE-Konfiguration kann nicht geleert werden, ${KASSENSITZUNG_AKTIV_GRUND}`,
    },
  })
  const { loading: testing, run: runTestConnection } = useActionSubmit({
    actionLabel: 'TSE-Verbindung testen',
    byCode: {
      // Die zentrale Meldung verweist auf den Bereich Finanzamt — auf dem der
      // Nutzer hier schon steht; der nächste Schritt ist das Speichern.
      tse_nicht_konfiguriert:
        'Bitte zuerst eine vollständige TSE-Konfiguration speichern.',
    },
  })

  const onSubmit = async (config: TSEKonfigurationSpeichern) => {
    await runSave(async () => {
      await onSave(config)
      form.reset({ ...config, apiKey: '', apiSecret: '' })
      setVerbindungStatus(null)
      toast.success('TSE-Konfiguration gespeichert.')
    })
  }

  const handleClear = async () => {
    await runClear(async () => {
      await onClear()
      form.reset(emptyTSEKonfiguration)
      setVerbindungStatus(null)
      toast.success('TSE-Konfiguration geleert.')
    })
  }

  const handleTestConnection = async () => {
    await runTestConnection(async () => {
      const status = await onTestConnection()
      setVerbindungStatus(status)
      if (status.signiertfaehig) {
        toast.success('TSE-Verbindung erfolgreich getestet.')
      } else {
        toast.error('Verbindung steht, aber die TSE ist nicht signierfähig.')
      }
    })
  }

  return (
    <form
      className="grid gap-4"
      onSubmit={(e) => {
        e.preventDefault()
        void form.handleSubmit(onSubmit)()
      }}
    >
      <TextField
        form={form}
        name="apiKey"
        label="API-Key *"
        type="password"
        placeholder="fiskaly API-Key"
        autoComplete="off"
        description={
          apiKeyGesetzt
            ? 'Ein API-Key ist bereits hinterlegt. Beim Speichern wird der hier eingegebene Wert übernommen.'
            : undefined
        }
      />
      <TextField
        form={form}
        name="apiSecret"
        label="API-Secret *"
        type="password"
        placeholder="fiskaly API-Secret"
        autoComplete="off"
        description={
          apiSecretGesetzt
            ? 'Ein API-Secret ist bereits hinterlegt. Beim Speichern wird der hier eingegebene Wert übernommen.'
            : undefined
        }
      />
      <TextField
        form={form}
        name="tssId"
        label="TSS-ID *"
        placeholder="z.B. 123e4567-e89b-12d3-a456-426614174000"
      />
      <TextField
        form={form}
        name="clientId"
        label="Client-ID *"
        placeholder="z.B. KASSE-1"
      />

      <div className="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving || clearing}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => void handleTestConnection()}
          disabled={saving || clearing || testing}
        >
          {testing ? 'Teste Verbindung…' : 'Verbindung testen'}
        </Button>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              type="button"
              variant="outline"
              disabled={saving || clearing || testing}
            >
              {clearing ? 'Leeren…' : 'Alle Felder leeren'}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>TSE-Konfiguration leeren?</AlertDialogTitle>
              <AlertDialogDescription>
                API-Key, API-Secret, TSS-ID und Client-ID werden gelöscht. Ohne
                TSE werden Vorgänge nicht nach § 146a AO signiert.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
              <AlertDialogAction
                variant="destructive-solid"
                disabled={clearing}
                onClick={() => void handleClear()}
              >
                Leeren
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {verbindungStatus ? (
        <VerbindungStatusAnzeige status={verbindungStatus} />
      ) : (
        <p className="text-sm text-muted-foreground">
          Das Ergebnis erscheint nach einem Verbindungstest.
        </p>
      )}
    </form>
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
    isError,
    refetch,
    saveTSEKonfiguration,
    clearTSEKonfiguration,
    testTSEVerbindung,
  } = useTSEKonfiguration()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade TSE-Konfiguration…</p>
    )
  } else if (isError) {
    inhalt = (
      <LadefehlerAlert
        titel="TSE-Konfiguration konnte nicht geladen werden"
        onErneutVersuchen={() => void refetch()}
      />
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
