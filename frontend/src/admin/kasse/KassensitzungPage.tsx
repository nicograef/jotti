import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

import { AdminPageHeader } from '../components/AdminPageHeader'
import { EroeffnenSection } from './EroeffnenSection'
import { GeldtransitSection } from './GeldtransitSection'
import { useOffeneKassensitzung } from './hooks'
import { KasseAbschliessenSection } from './KasseAbschliessenSection'
import { kassensitzungStatusLabel } from './Kassensitzung'

export { EroeffnenSection } from './EroeffnenSection'
export { KasseAbschliessenSection } from './KasseAbschliessenSection'

export function KassensitzungPage() {
  const { kassensitzung, isPending, isError, refetch } =
    useOffeneKassensitzung()

  const header = (
    <AdminPageHeader
      titel="Kassentag"
      unterzeile="Ein Kassentag läuft von der Eröffnung bis zum Tagesabschluss (Z-Bon)."
    />
  )

  if (isPending) {
    return (
      <>
        {header}
        <p className="mt-4 text-muted-foreground">Laden…</p>
      </>
    )
  }

  // Expliziter Fehlerzustand statt des Leer-Defaults („Keine Kassensitzung
  // geöffnet.") — sonst wirkt die Kasse bei Netzabbruch fälschlich geschlossen.
  if (isError) {
    return (
      <>
        {header}
        <LadefehlerAlert
          titel="Kassendaten konnten nicht geladen werden"
          onErneutVersuchen={() => void refetch()}
          className="mt-4"
        />
      </>
    )
  }

  return (
    <>
      {header}

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
