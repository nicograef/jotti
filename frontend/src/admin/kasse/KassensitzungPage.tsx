import { useState } from 'react'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { formatCents, parseCents } from '@/lib/utils'

import { kasseBackend, useKassenbestand, useOffeneKassensitzung } from './hooks'

export function KassensitzungPage() {
  const { kassensitzung, loading, reload } = useOffeneKassensitzung()
  const { kassenbestand } = useKassenbestand(kassensitzung?.zNr ?? null)

  if (loading) {
    return (
      <>
        <h1 className="text-2xl font-bold">Kassensitzung</h1>
        <p className="mt-4 text-muted-foreground">Laden…</p>
      </>
    )
  }

  return (
    <>
      <h1 className="text-2xl font-bold">Kassensitzung</h1>

      {kassensitzung ? (
        <div className="mt-4 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Kassensitzung #{String(kassensitzung.zNr)}
                <Badge variant="secondary">{kassensitzung.status}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <p>
                <span className="text-muted-foreground">Datum:</span>{' '}
                {kassensitzung.datum}
              </p>
              {kassensitzung.bezeichnung && (
                <p>
                  <span className="text-muted-foreground">Bezeichnung:</span>{' '}
                  {kassensitzung.bezeichnung}
                </p>
              )}
              {kassenbestand && (
                <p>
                  <span className="text-muted-foreground">
                    Soll-Kassenbestand:
                  </span>{' '}
                  {formatCents(kassenbestand.sollBestandCents)} €
                </p>
              )}
            </CardContent>
          </Card>

          <AnfangsbestandSection onSuccess={reload} />
          <KassenbewegungSection onSuccess={reload} />
          <KassensturzSection onSuccess={reload} />
          <TagesabschlussSection onSuccess={reload} />
        </div>
      ) : (
        <div className="mt-4 space-y-6">
          <p className="text-muted-foreground">Keine Kassensitzung geöffnet.</p>
          <EroeffnenSection onSuccess={reload} />
        </div>
      )}
    </>
  )
}

function EroeffnenSection({ onSuccess }: { onSuccess: () => void }) {
  const [datum, setDatum] = useState(() =>
    new Date().toISOString().slice(0, 10),
  )
  const [bezeichnung, setBezeichnung] = useState('')
  const [saving, setSaving] = useState(false)

  const handleEroeffnen = async () => {
    if (!datum) {
      toast.error('Bitte ein Datum eingeben.')
      return
    }
    setSaving(true)
    try {
      await kasseBackend.kassensitzungEroeffnen(datum, bezeichnung)
      toast.success('Kassensitzung eröffnet.')
      onSuccess()
    } catch {
      toast.error('Fehler beim Eröffnen der Kassensitzung.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensitzung eröffnen</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="ks-datum">Datum</Label>
          <Input
            id="ks-datum"
            type="date"
            value={datum}
            onChange={(e) => {
              setDatum(e.target.value)
            }}
            className="mt-1 w-48"
          />
        </div>
        <div>
          <Label htmlFor="ks-bezeichnung">Bezeichnung (optional)</Label>
          <Input
            id="ks-bezeichnung"
            value={bezeichnung}
            onChange={(e) => {
              setBezeichnung(e.target.value)
            }}
            placeholder="z.B. Sommerfest Tag 1"
            className="mt-1 w-72"
          />
        </div>
        <Button onClick={() => void handleEroeffnen()} disabled={saving}>
          Kassensitzung eröffnen
        </Button>
      </CardContent>
    </Card>
  )
}

function AnfangsbestandSection({ onSuccess }: { onSuccess: () => void }) {
  const [betragEuro, setBetragEuro] = useState('')
  const [saving, setSaving] = useState(false)

  const handleSetzen = async () => {
    const cents = parseCents(betragEuro)
    if (cents < 0) {
      toast.error('Bitte einen gültigen Betrag eingeben.')
      return
    }
    setSaving(true)
    try {
      await kasseBackend.anfangsbestandSetzen(cents)
      toast.success('Anfangsbestand gesetzt.')
      setBetragEuro('')
      onSuccess()
    } catch {
      toast.error('Fehler beim Setzen des Anfangsbestands.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Anfangsbestand setzen</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="anfangsbestand">Betrag (€)</Label>
          <Input
            id="anfangsbestand"
            type="text"
            inputMode="decimal"
            value={betragEuro}
            onChange={(e) => {
              setBetragEuro(e.target.value)
            }}
            placeholder="z.B. 150,00"
            className="mt-1 w-40"
          />
        </div>
        <Button onClick={() => void handleSetzen()} disabled={saving}>
          Anfangsbestand setzen
        </Button>
      </CardContent>
    </Card>
  )
}

function KassenbewegungSection({ onSuccess }: { onSuccess: () => void }) {
  const [art, setArt] = useState('einnahme')
  const [betragEuro, setBetragEuro] = useState('')
  const [kommentar, setKommentar] = useState('')
  const [saving, setSaving] = useState(false)

  const handleBuchen = async () => {
    const cents = parseCents(betragEuro)
    if (cents <= 0) {
      toast.error('Bitte einen gültigen Betrag eingeben.')
      return
    }
    setSaving(true)
    try {
      await kasseBackend.kassenbewegungBuchen(art, cents, kommentar)
      toast.success('Kassenbewegung gebucht.')
      setBetragEuro('')
      setKommentar('')
      onSuccess()
    } catch {
      toast.error('Fehler beim Buchen der Kassenbewegung.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassenbewegung buchen</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-4">
          <label className="flex items-center gap-1.5">
            <input
              type="radio"
              name="bewegung-art"
              value="einnahme"
              checked={art === 'einnahme'}
              onChange={() => {
                setArt('einnahme')
              }}
            />
            Einnahme
          </label>
          <label className="flex items-center gap-1.5">
            <input
              type="radio"
              name="bewegung-art"
              value="ausgabe"
              checked={art === 'ausgabe'}
              onChange={() => {
                setArt('ausgabe')
              }}
            />
            Ausgabe
          </label>
        </div>
        <div>
          <Label htmlFor="bewegung-betrag">Betrag (€)</Label>
          <Input
            id="bewegung-betrag"
            type="text"
            inputMode="decimal"
            value={betragEuro}
            onChange={(e) => {
              setBetragEuro(e.target.value)
            }}
            placeholder="z.B. 25,00"
            className="mt-1 w-40"
          />
        </div>
        <div>
          <Label htmlFor="bewegung-kommentar">Kommentar</Label>
          <Input
            id="bewegung-kommentar"
            value={kommentar}
            onChange={(e) => {
              setKommentar(e.target.value)
            }}
            placeholder="z.B. Wechselgeld Nachschub"
            className="mt-1 w-72"
          />
        </div>
        <Button onClick={() => void handleBuchen()} disabled={saving}>
          Kassenbewegung buchen
        </Button>
      </CardContent>
    </Card>
  )
}

function KassensturzSection({ onSuccess }: { onSuccess: () => void }) {
  const [istBestandEuro, setIstBestandEuro] = useState('')
  const [saving, setSaving] = useState(false)

  const handleKassensturz = async () => {
    const cents = parseCents(istBestandEuro)
    if (cents < 0) {
      toast.error('Bitte einen gültigen Betrag eingeben.')
      return
    }
    setSaving(true)
    try {
      await kasseBackend.kassensturzDurchfuehren(cents)
      toast.success('Kassensturz durchgeführt.')
      setIstBestandEuro('')
      onSuccess()
    } catch {
      toast.error('Fehler beim Kassensturz.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassensturz</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="kassensturz-ist">Ist-Bestand (€)</Label>
          <Input
            id="kassensturz-ist"
            type="text"
            inputMode="decimal"
            value={istBestandEuro}
            onChange={(e) => {
              setIstBestandEuro(e.target.value)
            }}
            placeholder="z.B. 342,50"
            className="mt-1 w-40"
          />
        </div>
        <Button onClick={() => void handleKassensturz()} disabled={saving}>
          Kassensturz durchführen
        </Button>
      </CardContent>
    </Card>
  )
}

function TagesabschlussSection({ onSuccess }: { onSuccess: () => void }) {
  const [saving, setSaving] = useState(false)

  const handleTagesabschluss = async () => {
    setSaving(true)
    try {
      await kasseBackend.tagesabschlussErstellen()
      toast.success('Tagesabschluss erstellt.')
      onSuccess()
    } catch {
      toast.error('Fehler beim Erstellen des Tagesabschlusses.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Tagesabschluss</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Schliesst die Kassensitzung ab. Voraussetzungen: Kassensturz
          durchgeführt, alle Tische auf Saldo 0.
        </p>
        <Button
          variant="destructive"
          onClick={() => void handleTagesabschluss()}
          disabled={saving}
        >
          Tagesabschluss erstellen
        </Button>
      </CardContent>
    </Card>
  )
}
