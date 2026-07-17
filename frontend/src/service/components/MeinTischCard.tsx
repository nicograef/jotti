import { ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router'

import { AuthSingleton } from '@/lib/Auth'
import { cn, formatEuro } from '@/lib/utils'

import type { TischSession } from '../table/Tisch'

interface MeinTischCardProps {
  state: TischSession
  // Position in der Eintritts-Staffelung (0-basiert) oder `undefined`, wenn die
  // Karte nicht animiert eintreten soll (z. B. nach einem Refetch).
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

  const statusFarbe = alleErledigt
    ? 'bg-green-600'
    : anzahlEigeneOffen > 0
      ? 'bg-destructive'
      : 'bg-amber-500'

  return (
    <button
      type="button"
      onClick={handleClick}
      // Listen-Eintritt (Handoff): fadeUp 450 ms, 60 ms Stagger je Karte, nur
      // beim ersten Aufbau. Der Verzögerungswert ist dynamisch und steht daher
      // inline; die weiche Kurve überschreibt die kanonische 250-ms-ease-Utility.
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

// countOffenePositionen zählt die noch offenen (unbezahlten) Positionen am Tisch
// (je Position einmal über die positionId) und zusätzlich, wie viele davon von
// der angemeldeten Servicekraft bestellt wurden.
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
