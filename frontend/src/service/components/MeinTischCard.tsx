import { ChevronRight } from 'lucide-react'
import { useNavigate } from 'react-router'

import { AuthSingleton } from '@/lib/Auth'
import { cn, formatCents } from '@/lib/utils'

import type { TischSession } from '../table/Tisch'

interface MeinTischCardProps {
  state: TischSession
}

export function MeinTischCard({ state }: MeinTischCardProps) {
  const navigate = useNavigate()

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
      className={cn(
        'flex w-full items-center gap-3 rounded-xl bg-card p-4 text-left ring-1 ring-foreground/10 transition-colors hover:bg-accent/50',
        alleErledigt && 'opacity-75',
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
            {formatCents(state.saldoCents)}&nbsp;€
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
