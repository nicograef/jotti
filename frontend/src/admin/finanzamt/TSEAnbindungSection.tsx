import { NavLink } from 'react-router'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import { useTSEStatus } from '../settings/hooks'

export function TSEAnbindungSection() {
  const { tseStatus, isPending, error } = useTSEStatus()

  return (
    <Card>
      <CardHeader>
        <CardTitle>TSE-Anbindung</CardTitle>
        <CardDescription>
          Die technische Sicherheitseinrichtung (TSE) signiert jeden
          Kassenvorgang. Die Zugangsdaten werden auf einer eigenen Seite
          eingerichtet.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        {isPending && (
          <p className="text-muted-foreground text-sm">Lade TSE-Status…</p>
        )}
        {error && (
          <p className="text-destructive text-sm">
            Fehler beim Laden des TSE-Status.
          </p>
        )}
        {tseStatus && (
          <dl className="grid gap-2 text-sm">
            <div className="flex justify-between gap-4">
              <dt className="text-muted-foreground">Konfiguriert</dt>
              <dd className="font-medium">
                {tseStatus.istKonfiguriert ? 'Ja' : 'Nein'}
              </dd>
            </div>
            <div className="flex justify-between gap-4">
              <dt className="text-muted-foreground">Umgebung</dt>
              <dd className="font-medium">{tseStatus.umgebung || '—'}</dd>
            </div>
          </dl>
        )}
        <Button asChild variant="outline" className="w-fit">
          <NavLink to="/admin/tse-einrichtung">Einrichten oder ändern</NavLink>
        </Button>
      </CardContent>
    </Card>
  )
}
