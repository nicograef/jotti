import { useNavigate } from 'react-router'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { AuthSingleton } from '@/lib/Auth'
import { formatCents } from '@/lib/utils'

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

  return (
    <Card
      className="cursor-pointer hover:bg-accent/50 transition-colors"
      onClick={handleClick}
    >
      <CardContent className="pt-4">
        <div className="flex items-start justify-between mb-2">
          <span className="font-semibold text-base">{state.tischName}</span>
          <div className="text-right">
            <div className="text-[11px] font-medium uppercase tracking-[0.04em] text-muted-foreground">
              Offen
            </div>
            <div className="font-bold tabular-nums">
              {formatCents(state.saldoCents)}&nbsp;€
            </div>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          {anzahlOffen > 0 && (
            <>
              <Badge variant="secondary">{anzahlOffen} offen</Badge>
              <span className="text-sm text-muted-foreground">
                davon {anzahlEigeneOffen} von dir
              </span>
            </>
          )}
          {alleErledigt && (
            <span className="text-sm text-green-600 font-medium">
              Alles erledigt
            </span>
          )}
        </div>
      </CardContent>
    </Card>
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
