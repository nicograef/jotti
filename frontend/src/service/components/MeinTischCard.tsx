import { ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router'

import { AuthSingleton } from '@/lib/Auth'
import { cn, formatEuro } from '@/lib/utils'

import type { TischSession } from '../table/Tisch'

interface MeinTischCardProps {
  state: TischSession
  // Position in der Eintritts-Staffelung; `undefined` = kein animierter Eintritt.
  eintrittIndex?: number
}

export function MeinTischCard({ state, eintrittIndex }: MeinTischCardProps) {
  const navigate = useNavigate()
  // Beim Mount erfasst: so bleibt die Staffelung stabil, wenn die Elternliste
  // während der Animation neu rendert (Übersicht-Query trifft separat ein).
  const [eintritt] = useState(eintrittIndex)

  const handleClick = () => {
    void navigate(`/service/tische/${state.tischId.toString()}`)
  }

  const { anzahlOffen, anzahlEigeneOffen } = countOffenePositionen(state)
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
      style={
        eintritt === undefined
          ? undefined
          : { animationDelay: `${(eintritt * 60).toString()}ms` }
      }
      className={cn(
        'flex w-full items-center gap-3 rounded-xl bg-card p-4 text-left ring-1 ring-foreground/10 transition-colors hover:bg-accent/50',
        alleErledigt && 'opacity-75',
        eintritt !== undefined &&
          'animate-fade-up [animation-duration:450ms] [animation-timing-function:cubic-bezier(0.2,0.7,0.3,1)]',
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

function countOffenePositionen(state: TischSession) {
  const myUserId = AuthSingleton.userId
  const offeneIds = new Set<string>()
  const eigeneOffeneIds = new Set<string>()

  for (const position of state.unbezahltePositionen) {
    offeneIds.add(position.positionId)
    if (position.bestellerUserId === myUserId) {
      eigeneOffeneIds.add(position.positionId)
    }
  }

  return {
    anzahlOffen: offeneIds.size,
    anzahlEigeneOffen: eigeneOffeneIds.size,
  }
}
