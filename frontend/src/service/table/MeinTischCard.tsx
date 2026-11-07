import { ChevronRight } from 'lucide-react'
import { useNavigate } from 'react-router'

import { useEintritt } from '@/hooks/use-eintritt'
import { cn, formatEuro } from '@/lib/utils'

import type { TischSession } from './Tisch'

interface MeinTischCardProps {
  state: TischSession
  // Position in der Eintritts-Staffelung; `undefined` = kein animierter Eintritt.
  eintrittIndex?: number
}

export function MeinTischCard({ state, eintrittIndex }: MeinTischCardProps) {
  const navigate = useNavigate()
  const eintritt = useEintritt(eintrittIndex)

  const handleClick = () => {
    void navigate(`/service/tische/${state.tischId.toString()}`)
  }

  const anzahlOffen = state.unbezahltePositionen.length
  const { anzahlEigeneOffen } = state
  const alleErledigt = anzahlOffen === 0

  // Farbsemantik: eigene offene Positionen fordern zur Aktion auf (amber), nur
  // fremde offene sind neutral wartend (muted), alles erledigt ist grün. Rot
  // bleibt Storno-/Fehlerzuständen vorbehalten (docs/decisions.md D04).
  const statusFarbe = alleErledigt
    ? 'bg-green-600'
    : anzahlEigeneOffen > 0
      ? 'bg-amber-500'
      : 'bg-muted-foreground'

  return (
    <button
      type="button"
      onClick={handleClick}
      style={eintritt.style}
      className={cn(
        'flex w-full items-center gap-3 rounded-xl bg-card p-4 text-left ring-1 ring-foreground/10 transition-colors hover:bg-accent/50',
        alleErledigt && 'opacity-75',
        eintritt.className,
      )}
    >
      <span
        aria-hidden
        className={cn('size-2.5 shrink-0 rounded-full', statusFarbe)}
      />
      <div className="min-w-0 flex-1">
        <div className="text-base font-semibold">{state.tischName}</div>
        {alleErledigt ? (
          <div className="text-[13px] font-medium text-green-600">
            Alles bezahlt
          </div>
        ) : (
          <div className="text-[13px] text-muted-foreground">
            {anzahlOffen} offen · {anzahlEigeneOffen} von dir
          </div>
        )}
      </div>
      {!alleErledigt && (
        <div className="shrink-0 text-right">
          <div className="text-[11px] font-medium uppercase tracking-[0.04em] text-muted-foreground">
            Offen
          </div>
          <div className="font-bold tabular-nums">
            {formatEuro(state.saldoCents)}
          </div>
        </div>
      )}
      <ChevronRight className="size-5 shrink-0 text-muted-foreground" />
    </button>
  )
}
