import { useNavigate } from 'react-router'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
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

  const hatAusstehende = state.ausstehendePositionen.length > 0
  const hatUnbezahlte = state.unbezahltePositionen.length > 0
  const hatAuszahlung = state.saldoCents < 0
  const alleErledigt =
    !hatAusstehende && !hatUnbezahlte && state.saldoCents >= 0

  return (
    <Card
      className="cursor-pointer hover:bg-accent/50 transition-colors"
      onClick={handleClick}
    >
      <CardContent className="pt-4">
        <div className="flex items-center justify-between mb-2">
          <span className="font-semibold text-base">{state.tischName}</span>
          <span
            className={
              state.saldoCents < 0
                ? 'font-medium text-destructive'
                : 'font-medium'
            }
          >
            {formatCents(state.saldoCents)} €
          </span>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {hatAusstehende && (
            <Badge variant="secondary">
              {state.ausstehendePositionen.length} ausstehend
            </Badge>
          )}
          {hatUnbezahlte && (
            <Badge variant="outline">
              {state.unbezahltePositionen.length} unbezahlt
            </Badge>
          )}
          {hatAuszahlung && (
            <Badge variant="destructive">
              Auszahlung: {formatCents(Math.abs(state.saldoCents))} €
            </Badge>
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
