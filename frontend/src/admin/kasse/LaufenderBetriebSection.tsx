import { MinusCircle, PlusCircle } from 'lucide-react'
import { useState } from 'react'

import { formatStand } from '@/admin/reporting/utils'
import { Button } from '@/components/ui/button'
import { formatEuro } from '@/lib/utils'

import { GeldtransitDialog } from './GeldtransitDialog'
import { useGeldtransitListe, useKassenbestand } from './hooks'
import {
  type GeldtransitBuchung,
  type GeldtransitRichtung,
} from './Kassensitzung'

// formatBewegungZeit gibt die lokale Uhrzeit HH:MM einer Bewegung aus.
function formatBewegungZeit(zeitpunkt: string): string {
  return new Date(zeitpunkt).toLocaleTimeString('de-DE', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

function AufschluesselungKachel({
  label,
  betragCents,
}: {
  label: string
  betragCents: number | null
}) {
  return (
    <div className="rounded-lg bg-muted/60 px-3 py-2.5">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-0.5 text-sm font-semibold">
        {betragCents !== null ? formatEuro(betragCents) : '—'}
      </div>
    </div>
  )
}

function BewegungZeile({ buchung }: { buchung: GeldtransitBuchung }) {
  const istEntnahme = buchung.richtung === 'entnahme'
  const richtungLabel = istEntnahme ? 'Entnahme' : 'Einlage'
  const vorzeichen = istEntnahme ? '−' : '+'
  return (
    <div className="flex items-center justify-between gap-3 px-3.5 py-2 text-sm">
      <span className="min-w-0 flex-1 truncate text-muted-foreground">
        {formatBewegungZeit(buchung.zeitpunkt)} · {richtungLabel} · „
        {buchung.kommentar}" · {buchung.gebuchtVon}
      </span>
      <span
        className={`shrink-0 font-semibold ${istEntnahme ? 'text-destructive' : ''}`}
      >
        {vorzeichen} {formatEuro(buchung.betragCents)}
      </span>
    </div>
  )
}

// LaufenderBetriebSection ist der Inhalt von Schritt 2 des Kassentag-Steppers:
// der Soll-Bestand groß mit Stand-Zeit, die vier Aufschlüsselungs-Kacheln und die
// Liste der heutigen Kassenbewegungen mit Einlegen-/Entnehmen-Buttons.
export function LaufenderBetriebSection({
  kassensitzungNr,
  onBuchung,
}: {
  kassensitzungNr: number
  onBuchung: () => void
}) {
  const { kassenbestand, dataUpdatedAt } = useKassenbestand(kassensitzungNr)
  const { buchungen } = useGeldtransitListe(kassensitzungNr)
  const [dialogRichtung, setDialogRichtung] =
    useState<GeldtransitRichtung | null>(null)

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">
            So viel Bargeld müsste jetzt rechnerisch in der Kasse liegen:
          </p>
        </div>
        <div className="text-right">
          <div className="text-3xl font-extrabold tracking-tight tabular-nums">
            {kassenbestand !== null
              ? formatEuro(kassenbestand.sollBestandCents)
              : '—'}
          </div>
          <div className="text-xs text-muted-foreground">
            Soll-Bestand
            {dataUpdatedAt ? ` · Stand ${formatStand(dataUpdatedAt)}` : ''}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
        <AufschluesselungKachel
          label="Anfangsbestand"
          betragCents={kassenbestand?.anfangsbestandCents ?? null}
        />
        <AufschluesselungKachel
          label="+ Bareinnahmen"
          betragCents={kassenbestand?.bareinnahmenCents ?? null}
        />
        <AufschluesselungKachel
          label="+ Einlagen"
          betragCents={kassenbestand?.einlagenCents ?? null}
        />
        <AufschluesselungKachel
          label="− Entnahmen"
          betragCents={kassenbestand?.entnahmenCents ?? null}
        />
      </div>

      <div className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="text-sm font-semibold">
            Heutige Kassenbewegungen
          </span>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setDialogRichtung('einlage')
              }}
            >
              <PlusCircle />
              Geld einlegen
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setDialogRichtung('entnahme')
              }}
            >
              <MinusCircle />
              Geld entnehmen
            </Button>
          </div>
        </div>
        {buchungen.length > 0 ? (
          <div className="divide-y overflow-hidden rounded-lg border">
            {buchungen.map((buchung) => (
              <BewegungZeile
                key={`${buchung.zeitpunkt}-${buchung.richtung}-${buchung.gebuchtVon}-${buchung.kommentar}`}
                buchung={buchung}
              />
            ))}
          </div>
        ) : (
          <p className="rounded-lg border px-3.5 py-3 text-sm text-muted-foreground">
            Noch keine Einlagen oder Entnahmen gebucht.
          </p>
        )}
      </div>

      <GeldtransitDialog
        open={dialogRichtung !== null}
        onOpenChange={(open) => {
          if (!open) setDialogRichtung(null)
        }}
        richtung={dialogRichtung}
        onSuccess={onBuchung}
      />
    </div>
  )
}
