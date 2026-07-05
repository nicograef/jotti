import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

import { EroeffnenSection } from './EroeffnenSection'
import { GeldtransitSection } from './GeldtransitSection'
import { useOffeneKassensitzung } from './hooks'
import { KasseAbschliessenSection } from './KasseAbschliessenSection'
import { kassensitzungStatusLabel } from './Kassensitzung'

export { EroeffnenSection } from './EroeffnenSection'
export { KasseAbschliessenSection } from './KasseAbschliessenSection'

export function KassensitzungPage() {
  const { kassensitzung, isPending, refetch } = useOffeneKassensitzung()

  if (isPending) {
    return (
      <>
        <h1 className="text-2xl font-bold">Kassensitzung</h1>
        <p className="mt-4 text-muted-foreground">Laden…</p>
      </>
    )
  }

  return (
    <>
      <h1 className="text-2xl font-bold">Kassensitzung</h1>

      {kassensitzung ? (
        <div className="mt-4 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                Kassensitzung #{String(kassensitzung.zNr)}
                <Badge variant="secondary">
                  {kassensitzungStatusLabel(kassensitzung.status).symbol}{' '}
                  {kassensitzungStatusLabel(kassensitzung.status).text}
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <p>
                <span className="text-muted-foreground">Datum:</span>{' '}
                {kassensitzung.datum}
              </p>
              <p>
                <span className="text-muted-foreground">Bezeichnung:</span>{' '}
                {kassensitzung.bezeichnung}
              </p>
            </CardContent>
          </Card>

          <GeldtransitSection onSuccess={() => void refetch()} />
          <KasseAbschliessenSection
            kassensitzungNr={kassensitzung.zNr}
            onSuccess={() => void refetch()}
          />
        </div>
      ) : (
        <div className="mt-4 space-y-6">
          <p className="text-muted-foreground">Keine Kassensitzung geöffnet.</p>
          <EroeffnenSection onSuccess={() => void refetch()} />
        </div>
      )}
    </>
  )
}
