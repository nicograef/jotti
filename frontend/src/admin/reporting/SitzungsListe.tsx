import { NavLink } from 'react-router'

import type { OffeneKassensitzung } from '@/admin/kasse/KasseBackend'
import { cn, formatCents } from '@/lib/utils'

import type { AbgeschlosseneSitzung } from './types'
import { formatDatumKurz } from './utils'

// SitzungsListe ist die linke Spalte der Kassenberichte: die offene Sitzung als
// nicht wählbarer Hinweis, der zur Übersicht führt, darunter die abgeschlossenen
// Sitzungen als wählbare Karten (Datum, Nr., Bezeichnung, Gesamtumsatz). Status-
// Emojis entfallen; der Auswahl-Zustand folgt dem Design-Handoff (Abschnitt 1b).
export function SitzungsListe({
  sitzungen,
  offeneSitzung,
  selectedNr,
  onSelect,
}: {
  sitzungen: AbgeschlosseneSitzung[]
  offeneSitzung: OffeneKassensitzung | null
  selectedNr: number | null
  onSelect: (nr: number) => void
}) {
  return (
    <div className="flex flex-col gap-2">
      {offeneSitzung && (
        <NavLink
          to="/admin/auswertung"
          className="flex flex-col gap-0.5 rounded-lg border p-3 opacity-65 transition-opacity hover:opacity-100"
        >
          <div className="flex items-center justify-between gap-2">
            <span className="text-sm font-semibold">
              Nr. {offeneSitzung.zNr}
            </span>
            <span className="inline-flex items-center gap-1.5 text-xs font-medium text-primary">
              <span className="size-1.5 rounded-full bg-primary" />
              offen
            </span>
          </div>
          <span className="text-xs text-muted-foreground">
            {offeneSitzung.bezeichnung} · läuft — siehe Übersicht
          </span>
        </NavLink>
      )}

      {sitzungen.map((sitzung) => {
        const selected = sitzung.zNr === selectedNr
        return (
          <button
            key={sitzung.zNr}
            type="button"
            onClick={() => {
              onSelect(sitzung.zNr)
            }}
            aria-pressed={selected}
            className={cn(
              'flex flex-col gap-0.5 rounded-lg border p-3 text-left transition-colors',
              selected
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/40',
            )}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm font-semibold">
                {formatDatumKurz(sitzung.datum)} · Nr. {sitzung.zNr}
              </span>
              <span className="text-sm font-semibold">
                {formatCents(sitzung.umsatzGesamtCents)} €
              </span>
            </div>
            <span className="text-xs text-muted-foreground">
              {sitzung.bezeichnung}
            </span>
          </button>
        )
      })}
    </div>
  )
}
