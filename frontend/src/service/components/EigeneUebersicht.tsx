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
    <div className="grid grid-cols-2 gap-3 my-4">
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
  )
}
