import { NavLink } from 'react-router'

import type { AktiveKassensitzung } from '@/admin/kasse/KasseBackend'
import { StatusDot } from '@/admin/components/StatusDot'
import { cn, formatEuro } from '@/lib/utils'

import type { AbgeschlosseneSitzung } from './types'
import { formatDatumKurz } from './utils'

export function SitzungsListe({
  sitzungen,
  aktiveSitzung,
  abschlussUnterbrochen,
  selectedNr,
  onSelect,
}: {
  sitzungen: AbgeschlosseneSitzung[]
  aktiveSitzung: AktiveKassensitzung | null
  abschlussUnterbrochen: boolean
  selectedNr: number | null
  onSelect: (nr: number) => void
}) {
  return (
    <div className="flex flex-col gap-2">
      {aktiveSitzung && (
        <NavLink
          to="/admin/auswertung"
          className="flex flex-col gap-0.5 rounded-lg border p-3 opacity-65 transition-opacity hover:opacity-100"
        >
          <div className="flex items-center justify-between gap-2">
            <span className="text-sm font-semibold">
              Nr. {aktiveSitzung.zNr}
            </span>
            <span
              className={cn(
                'inline-flex items-center gap-1.5 text-xs font-medium',
                abschlussUnterbrochen ? 'text-destructive' : 'text-primary',
              )}
            >
              <StatusDot
                zustand={abschlussUnterbrochen ? 'fehler' : 'ok'}
                label={
                  abschlussUnterbrochen
                    ? 'Kassensitzung: Abschluss unterbrochen'
                    : 'Kassensitzung offen'
                }
              />
              {abschlussUnterbrochen ? 'Abschluss unterbrochen' : 'offen'}
            </span>
          </div>
          <span className="text-xs text-muted-foreground">
            {aktiveSitzung.bezeichnung} ·{' '}
            {abschlussUnterbrochen
              ? 'siehe Übersicht'
              : 'läuft — siehe Übersicht'}
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
                {formatEuro(sitzung.umsatzGesamtCents)}
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
