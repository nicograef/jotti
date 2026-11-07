import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'

import { type TSSBefund, type Umgebung } from './TSEBackend'

export function UmgebungAnzeige({ umgebung }: { umgebung: Umgebung }) {
  if (umgebung === 'LIVE') {
    return (
      <Alert variant="destructive">
        <AlertTitle className="flex items-center gap-2">
          Umgebung: <Badge variant="destructive">LIVE</Badge>
        </AlertTitle>
        <AlertDescription>
          Dies ist die echte Produktivumgebung. Hier angelegte TSE verursachen
          laufende Kosten und lassen sich nicht löschen.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert>
      <AlertTitle className="flex items-center gap-2">
        Umgebung: <Badge variant="secondary">TEST</Badge>
      </AlertTitle>
      <AlertDescription>
        Test-Umgebung zum Ausprobieren. Hier signierte Belege sind steuerlich
        nicht gültig.
      </AlertDescription>
    </Alert>
  )
}

// Der rohe Zustand bleibt daneben als Kennung für den Support sichtbar.
function tssZustandKlartext(state: string): string {
  switch (state) {
    case 'CREATED':
      return 'neu angelegt'
    case 'UNINITIALIZED':
      return 'angelegt, noch nicht einsatzbereit'
    case 'INITIALIZED':
      return 'einsatzbereit'
    case 'DISABLED':
      return 'stillgelegt'
    default:
      return state
  }
}

export function TSSListe({ tssListe }: { tssListe: TSSBefund[] }) {
  if (tssListe.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        In diesem Konto wurde noch keine TSE gefunden.
      </p>
    )
  }

  return (
    <div className="grid gap-3">
      <p className="text-sm font-medium">Gefundene TSE</p>
      {tssListe.map((tss) => (
        <div key={tss.id} className="grid gap-1 rounded-md border p-3 text-sm">
          <div className="flex items-center justify-between gap-2">
            <Badge variant="outline">{tssZustandKlartext(tss.state)}</Badge>
            <span className="font-mono text-xs text-muted-foreground">
              {tss.state}
            </span>
          </div>
          <p className="text-muted-foreground">
            {tss.passenderClient
              ? 'Diese Kasse ist hier bereits angemeldet.'
              : 'Diese Kasse ist hier noch nicht angemeldet.'}
          </p>
          <p className="text-muted-foreground text-xs break-all">
            Technische Kennung: {tss.id}
          </p>
        </div>
      ))}
    </div>
  )
}
