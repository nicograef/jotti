import { CheckCircle2 } from 'lucide-react'
import { useNavigate } from 'react-router'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import type { EigeneUebersicht } from '../table/Tisch'

interface EigeneUebersichtProps {
  uebersicht: EigeneUebersicht
  loading: boolean
}

export function EigeneUebersichtKarten({
  uebersicht,
  loading,
}: EigeneUebersichtProps) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 gap-3 my-4">
        <Card>
          <CardHeader>
            <Skeleton className="h-4 w-24" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-6 w-12 mb-1" />
            <Skeleton className="h-4 w-20" />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Skeleton className="h-4 w-24" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-6 w-12 mb-1" />
            <Skeleton className="h-4 w-20" />
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="my-4 space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm text-muted-foreground">
              Bestellungen
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xl font-bold">{uebersicht.anzahlBestellungen}</p>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {formatCents(uebersicht.bestellungenCents)} €
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm text-muted-foreground">
              Kassiert
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xl font-bold">{uebersicht.anzahlZahlungen}</p>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {formatCents(uebersicht.zahlungenCents)} €
            </p>
          </CardContent>
        </Card>
      </div>

      <EigeneArbeitSicht uebersicht={uebersicht} />
    </div>
  )
}

function EigeneArbeitSicht({ uebersicht }: { uebersicht: EigeneUebersicht }) {
  const navigate = useNavigate()

  if (uebersicht.alleErledigt) {
    return (
      <Card className="border-green-600/30 bg-green-600/5">
        <CardContent className="flex items-center gap-2 py-3 text-green-700">
          <CheckCircle2 className="size-5" />
          <span className="font-medium">Alles erledigt!</span>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">
          Deine offenen Tische
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {uebersicht.offeneTische.map((tisch) => (
          <button
            key={tisch.tischId}
            type="button"
            onClick={() => {
              void navigate(`/service/tische/${tisch.tischId.toString()}`)
            }}
            className="flex w-full items-center justify-between rounded-md border p-3 text-left transition-colors hover:bg-accent/50"
          >
            <span className="font-medium">{tisch.tischName}</span>
            <Badge variant="secondary">{tisch.anzahlOffen} offen</Badge>
          </button>
        ))}
      </CardContent>
    </Card>
  )
}
