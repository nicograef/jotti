import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useActionSubmit } from '@/hooks/use-action-submit'
import {
  type Betreiber,
  type TSEKonfigurationSpeichern,
  type TSENachsignierAuftrag,
  type TSEVerbindungStatus,
} from '@/lib/EinstellungenBackend'

import {
  useBetreiber,
  useKassenidentitaet,
  useTSEKonfiguration,
  useTSENachsignierAuftraege,
} from './hooks'

function KassenidentitaetSection() {
  const { data: kassenidentitaet, isPending, error } = useKassenidentitaet()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!kassenidentitaet) return
    await navigator.clipboard.writeText(kassenidentitaet.seriennummer)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Kassenidentität</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Seriennummer und Inbetriebnahmedatum identifizieren diese jotti-Instanz
        eindeutig. Beide Angaben werden für die ELSTER-Meldung (§ 146a AO)
        benötigt; die Seriennummer erscheint zusätzlich auf jedem Kassenbeleg.
      </p>
      {isPending && (
        <p className="text-muted-foreground text-sm">Lade Kassenidentität…</p>
      )}
      {error && (
        <p className="text-destructive text-sm">
          Fehler beim Laden der Kassenidentität.
        </p>
      )}
      {kassenidentitaet && (
        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label>Seriennummer</Label>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm">
                {kassenidentitaet.seriennummer}
              </code>
              <Button
                variant="outline"
                size="icon"
                onClick={() => void handleCopy()}
                aria-label="Seriennummer kopieren"
              >
                {copied ? (
                  <Check className="h-4 w-4" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>
          <div className="grid gap-1.5">
            <Label>Inbetriebnahmedatum</Label>
            <p className="text-sm">
              {new Date(kassenidentitaet.angelegtAm).toLocaleDateString(
                'de-DE',
              )}
            </p>
          </div>
        </div>
      )}
    </section>
  )
}

function BetreiberForm({
  initial,
  onSave,
}: {
  initial: Betreiber
  onSave: (b: Betreiber) => Promise<void>
}) {
  const [form, setForm] = useState<Betreiber>(initial)
  const { loading: saving, run } = useActionSubmit({
    actionLabel: 'Betreiber-Stammdaten speichern',
  })

  const handleChange =
    (field: keyof Betreiber) => (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      setForm((prev) => ({
        ...prev,
        [field]:
          value === '' && (field === 'steuernummer' || field === 'ustId')
            ? null
            : value,
      }))
    }

  const handleSave = async () => {
    await run(async () => {
      await onSave(form)
      toast.success('Betreiber-Stammdaten gespeichert.')
    })
  }

  return (
    <div className="grid gap-4">
      <div className="grid gap-1.5">
        <Label htmlFor="vereinsname">Vereinsname *</Label>
        <Input
          id="vereinsname"
          value={form.vereinsname}
          onChange={handleChange('vereinsname')}
          placeholder="z.B. Sportverein Musterstadt e.V."
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="strasse">Straße und Hausnummer *</Label>
        <Input
          id="strasse"
          value={form.strasse}
          onChange={handleChange('strasse')}
          placeholder="z.B. Musterstraße 1"
        />
      </div>
      <div className="grid grid-cols-[120px_1fr] gap-4">
        <div className="grid gap-1.5">
          <Label htmlFor="plz">PLZ *</Label>
          <Input
            id="plz"
            value={form.plz}
            onChange={handleChange('plz')}
            placeholder="12345"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="ort">Ort *</Label>
          <Input
            id="ort"
            value={form.ort}
            onChange={handleChange('ort')}
            placeholder="Musterstadt"
          />
        </div>
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="steuernummer">Steuernummer (optional)</Label>
        <Input
          id="steuernummer"
          value={form.steuernummer ?? ''}
          onChange={handleChange('steuernummer')}
          placeholder="z.B. 12/345/67890"
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="ustId">USt-ID (optional)</Label>
        <Input
          id="ustId"
          value={form.ustId ?? ''}
          onChange={handleChange('ustId')}
          placeholder="z.B. DE123456789"
        />
      </div>
      <div>
        <Button onClick={() => void handleSave()} disabled={saving}>
          {saving ? 'Speichern…' : 'Speichern'}
        </Button>
      </div>
    </div>
  )
}

const emptyBetreiber: Betreiber = {
  vereinsname: '',
  strasse: '',
  plz: '',
  ort: '',
  steuernummer: null,
  ustId: null,
}

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

function TSEKonfigurationSection() {
  const {
    tseKonfiguration,
    isPending,
    error,
    saveTSEKonfiguration,
    clearTSEKonfiguration,
    testTSEVerbindung,
  } = useTSEKonfiguration()

  if (isPending) {
    return (
      <section className="max-w-2xl">
        <p className="text-muted-foreground text-sm">Lade TSE-Konfiguration…</p>
      </section>
    )
  }

  if (error) {
    return (
      <section className="max-w-2xl">
        <p className="text-destructive text-sm">
          Fehler beim Laden der TSE-Konfiguration.
        </p>
      </section>
    )
  }

  const initial: TSEKonfigurationSpeichern = {
    apiKey: '',
    apiSecret: '',
    tssId: tseKonfiguration?.tssId ?? '',
    clientId: tseKonfiguration?.clientId ?? '',
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">TSE-Integration (BYOT)</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Hier hinterlegst du die Zugangsdaten für deine Cloud-TSE. Die
        Seriennummer aus der Kassenidentität oben wird in fiskaly als
        serial_number benötigt.
      </p>
      <TSEKonfigurationForm
        key={`${tseKonfiguration?.tssId ?? ''}-${tseKonfiguration?.clientId ?? ''}-${String(tseKonfiguration?.apiKeyGesetzt ?? false)}-${String(tseKonfiguration?.apiSecretGesetzt ?? false)}`}
        initial={initial}
        apiKeyGesetzt={tseKonfiguration?.apiKeyGesetzt ?? false}
        apiSecretGesetzt={tseKonfiguration?.apiSecretGesetzt ?? false}
        onSave={saveTSEKonfiguration}
        onClear={clearTSEKonfiguration}
        onTestConnection={testTSEVerbindung}
      />
    </section>
  )
}

const NACHSIGNIER_STATUS_LABEL: Record<
  TSENachsignierAuftrag['status'],
  string
> = {
  offen: 'Offen',
  erledigt: 'Erledigt',
  fehlgeschlagen: 'Fehlgeschlagen',
  verworfen: 'Verworfen',
}

function NachsignierAuftragRow({
  auftrag,
  onZuruecksetzen,
  onVerwerfen,
}: {
  auftrag: TSENachsignierAuftrag
  onZuruecksetzen: (id: number) => Promise<void>
  onVerwerfen: (id: number) => Promise<void>
}) {
  const { loading, run } = useActionSubmit({
    actionLabel: 'Nachsignier-Auftrag aktualisieren',
  })

  const zeitraum = `${new Date(auftrag.erstelltAm).toLocaleString('de-DE')} – ${
    auftrag.erledigtAm !== null
      ? new Date(auftrag.erledigtAm).toLocaleString('de-DE')
      : 'offen'
  }`

  return (
    <div className="flex flex-col gap-2 py-4 border-b last:border-b-0">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div className="font-medium">
          {NACHSIGNIER_STATUS_LABEL[auftrag.status]} · {auftrag.processType}
        </div>
        <div className="text-sm text-muted-foreground">{zeitraum}</div>
      </div>
      <div className="text-sm text-muted-foreground">
        Transaktion: {auftrag.txId}
        {auftrag.versuche > 0 && <> · {auftrag.versuche} Versuche</>}
      </div>
      {auftrag.letzterFehler !== '' && (
        <p className="text-sm text-destructive">{auftrag.letzterFehler}</p>
      )}
      {auftrag.status === 'fehlgeschlagen' && (
        <div className="flex gap-2">
          <Button
            size="sm"
            disabled={loading}
            onClick={() =>
              void run(async () => {
                await onZuruecksetzen(auftrag.id)
                toast.success('Nachsignier-Auftrag wieder eingereiht.')
              })
            }
          >
            Zurücksetzen
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={loading}
            onClick={() =>
              void run(async () => {
                await onVerwerfen(auftrag.id)
                toast.success('Nachsignier-Auftrag verworfen.')
              })
            }
          >
            Verwerfen
          </Button>
        </div>
      )}
    </div>
  )
}

function TSENachsignierSection() {
  const { auftraege, isPending, error, zuruecksetzen, verwerfen } =
    useTSENachsignierAuftraege()

  let inhalt
  if (isPending) {
    inhalt = (
      <p className="text-muted-foreground text-sm">Lade Nachsignierungen…</p>
    )
  } else if (error) {
    inhalt = (
      <p className="text-destructive text-sm">
        Fehler beim Laden der Nachsignierungen.
      </p>
    )
  } else if (auftraege.length === 0) {
    inhalt = (
      <p className="text-muted-foreground text-sm">
        Keine Nachsignierungen — bisher wurden alle Vorgänge direkt signiert.
      </p>
    )
  } else {
    inhalt = (
      <div className="rounded-md border px-4">
        {auftraege.map((auftrag) => (
          <NachsignierAuftragRow
            key={auftrag.id}
            auftrag={auftrag}
            onZuruecksetzen={zuruecksetzen}
            onVerwerfen={verwerfen}
          />
        ))}
      </div>
    )
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">TSE-Nachsignierungen</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Vorgänge, die während eines TSE-Ausfalls erfasst wurden, werden hier
        automatisch nachsigniert. Die Liste dokumentiert zugleich die
        Ausfallzeiten (Beginn, Ende, Grund) für die gesetzliche
        Ausfalldokumentation. Fehlgeschlagene Aufträge können wieder eingereiht
        oder verworfen werden.
      </p>
      {inhalt}
    </section>
  )
}

function BetreiberSection() {
  const { betreiber, isPending, error, saveBetreiber } = useBetreiber()

  if (isPending) {
    return (
      <section className="max-w-2xl">
        <p className="text-muted-foreground text-sm">Lade Betreiber-Daten…</p>
      </section>
    )
  }

  if (error) {
    return (
      <section className="max-w-2xl">
        <p className="text-destructive text-sm">
          Fehler beim Laden der Betreiber-Daten.
        </p>
      </section>
    )
  }

  return (
    <section className="max-w-2xl">
      <h2 className="text-xl font-semibold mb-1">Betreiber-Stammdaten</h2>
      <p className="text-muted-foreground text-sm mb-4">
        Name und Adresse des Vereins erscheinen auf jedem Kassenbeleg (§ 6
        KassenSichV). Eine Kassensitzung kann erst eröffnet werden, wenn
        mindestens der Vereinsname gesetzt ist.
      </p>
      <BetreiberForm
        initial={betreiber ?? emptyBetreiber}
        onSave={saveBetreiber}
      />
    </section>
  )
}

export function EinstellungenPage() {
  return (
    <div className="flex flex-col gap-10">
      <KassenidentitaetSection />
      <hr />
      <TSEKonfigurationSection />
      <hr />
      <TSENachsignierSection />
      <hr />
      <BetreiberSection />
    </div>
  )
}
