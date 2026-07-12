import { Switch } from '@/components/ui/switch'
import { cn, formatCents } from '@/lib/utils'

import { type Tisch, TischStatus } from './Tisch'

interface TischItemProps {
  loading: boolean
  tisch: Tisch
  onEdit: (tischId: number) => void
  onActivate: (tischId: number) => Promise<void>
  onDeactivate: (tischId: number) => Promise<void>
}

// Kompakte Tisch-Kachel (Design-Handoff 1d): Name, Mini-Switch und Statustext.
// Ein offener Saldo zeigt den Betrag statt „aktiv“ und sperrt den Switch — das
// Backend erzwingt den Schutz zusätzlich als Single Source of Truth. Die
// Begründung steht als stets sichtbare Zeile (kein Hover-Tooltip: die
// Servicekräfte bedienen Touch-Handys). Klick auf die Kachel öffnet den
// Bearbeiten-Dialog (Umbenennen, Löschen); der Switch stoppt die Propagierung.
export function TischItem(props: TischItemProps) {
  const isActive = props.tisch.status === TischStatus.ACTIVE
  const hatSaldo = props.tisch.saldoCents > 0

  const oeffneEdit = () => {
    props.onEdit(props.tisch.id)
  }

  return (
    // Die Kachel ist klickbar, enthält aber selbst einen Switch (interaktives
    // Element). Ein <button> im <button> ist ungültiges HTML, deshalb ein
    // role="button"-<div> mit Tastaturbedienung statt eines echten Buttons.
    <div
      role="button"
      tabIndex={0}
      onClick={oeffneEdit}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          oeffneEdit()
        }
      }}
      className={cn(
        'flex cursor-pointer flex-col gap-2 rounded-lg border p-3 text-left transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
        !isActive && 'bg-muted/50 opacity-75',
      )}
    >
      <span className="font-semibold">{props.tisch.name}</span>
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span
          onClick={(e) => {
            // Switch-Klick darf den Kachel-Klick (Bearbeiten-Dialog) nicht auslösen.
            e.stopPropagation()
          }}
        >
          <Switch
            className="cursor-pointer"
            aria-label={isActive ? 'Tisch deaktivieren' : 'Tisch aktivieren'}
            disabled={props.loading || hatSaldo}
            checked={isActive}
            onCheckedChange={(checked) => {
              if (checked) {
                void props.onActivate(props.tisch.id)
              } else {
                void props.onDeactivate(props.tisch.id)
              }
            }}
          />
        </span>
        {hatSaldo ? (
          <span className="font-medium text-primary">
            {formatCents(props.tisch.saldoCents)} € offen
          </span>
        ) : (
          <span>{isActive ? 'aktiv' : 'aus'}</span>
        )}
      </div>
      {hatSaldo && (
        <span className="text-xs text-muted-foreground">
          Erst abrechnen, dann deaktivieren
        </span>
      )}
    </div>
  )
}
