import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { formatCents } from '@/lib/utils'

import {
  type Nennwert,
  NENNWERTE_CENTS,
  summeAusStueckzahlen,
} from './zaehlhilfe'

// nennwertLabel beschriftet einen Nennwert menschenlesbar: Münzen unter 1 €
// als „X ct", ab 1 € als „X €" (ohne Nachkommastellen, weil alle Nennwerte
// glatt sind).
function nennwertLabel(nennwert: Nennwert): string {
  return nennwert < 100
    ? `${String(nennwert)} ct`
    : `${String(nennwert / 100)} €`
}

// ZaehlhilfeInhalt hält die Zählung. Er wird beim Öffnen frisch gemountet
// (bedingt gerendert am Aufrufer), damit jede Zählung leer startet — ohne
// setState im Effekt.
function ZaehlhilfeInhalt({
  onOpenChange,
  onUebernehmen,
}: {
  onOpenChange: (open: boolean) => void
  onUebernehmen: (summeCents: number) => void
}) {
  const [stueckzahlen, setStueckzahlen] = useState<
    Partial<Record<Nennwert, number>>
  >({})

  const summeCents = summeAusStueckzahlen(stueckzahlen)

  const setAnzahl = (nennwert: Nennwert, roh: string) => {
    // Nur nicht-negative Ganzzahlen; leeres Feld ergibt 0 (kein Eintrag).
    const anzahl = Number.parseInt(roh, 10)
    const bereinigt = Number.isNaN(anzahl) || anzahl <= 0 ? 0 : anzahl
    setStueckzahlen((bisher) => ({ ...bisher, [nennwert]: bereinigt }))
  }

  return (
    <>
      <div className="grid grid-cols-2 gap-x-6 gap-y-2">
        {NENNWERTE_CENTS.map((nennwert) => {
          const id = `zaehlhilfe-${String(nennwert)}`
          const anzahl = stueckzahlen[nennwert]
          return (
            <div
              key={nennwert}
              className="flex items-center justify-between gap-3"
            >
              <Label htmlFor={id} className="w-14 shrink-0">
                {nennwertLabel(nennwert)}
              </Label>
              <Input
                id={id}
                type="number"
                min={0}
                step={1}
                inputMode="numeric"
                placeholder="0"
                className="h-8 w-20 text-right"
                value={anzahl !== undefined && anzahl > 0 ? String(anzahl) : ''}
                onChange={(e) => {
                  setAnzahl(nennwert, e.target.value)
                }}
              />
            </div>
          )
        })}
      </div>

      <div className="flex items-center justify-between border-t pt-4 text-sm">
        <span className="text-muted-foreground">Summe</span>
        <span
          className="text-base font-semibold"
          data-testid="zaehlhilfe-summe"
        >
          {formatCents(summeCents)} €
        </span>
      </div>

      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            onOpenChange(false)
          }}
        >
          Abbrechen
        </Button>
        <Button
          type="button"
          onClick={() => {
            onUebernehmen(summeCents)
            onOpenChange(false)
          }}
        >
          Übernehmen
        </Button>
      </DialogFooter>
    </>
  )
}

// ZaehlhilfeDialog ist eine rein clientseitige Zählhilfe: der Nutzer trägt je
// Nennwert eine Stückzahl ein, sieht die laufende Summe und übernimmt sie mit
// „Übernehmen" in das Ist-Bestand-Feld. Keine Netzwerk-Interaktion.
export function ZaehlhilfeDialog({
  open,
  onOpenChange,
  onUebernehmen,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onUebernehmen: (summeCents: number) => void
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Zählhilfe</DialogTitle>
          <DialogDescription>
            Trage je Nennwert die gezählte Stückzahl ein. Die Summe wird
            automatisch berechnet und lässt sich als Ist-Bestand übernehmen.
          </DialogDescription>
        </DialogHeader>
        {/* Frisches Mount pro Öffnung: jede Zählung startet leer. */}
        {open && (
          <ZaehlhilfeInhalt
            onOpenChange={onOpenChange}
            onUebernehmen={onUebernehmen}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
