import { Button } from '@/components/ui/button'
import { useVersionsGuard } from '@/hooks/use-versions-guard'
import { seiteNeuLaden } from '@/lib/reload'

/**
 * Hinweis zum Versions-Handshake — sichtbar nur, solange ein Vorgang offen ist
 * (`wartet`) oder ein Reload wirkungslos geblieben ist (`gebremst`).
 *
 * Bewusst kein modaler Dialog: Radix' `AlertDialog` fängt zwingend den Fokus
 * (`modal: true` fest verdrahtet) und legt die Seite darunter still — der
 * Hinweis wartet aber genau auf die Bedienung des laufenden Vorgangs.
 *
 * Er steht im Fluss über dem Seitenlayout statt als fixierte Leiste: Oben wie
 * unten läge er über Bedienelementen (ServiceDock, Fußleiste der Tischauswahl,
 * fixierte Kopfleisten). Der Text nennt keine Richtung — nach einem Rollback
 * meldet der Handshake auch eine ältere Serverversion.
 */
export function VersionsHinweis() {
  const versionsZustand = useVersionsGuard()

  if (versionsZustand === 'aus' || versionsZustand === 'laedt') return null

  return (
    <div
      role="alert"
      // `relative z-[60] pointer-events-auto` hebt den Hinweis über ein offenes
      // Radix-Modal (Overlay `fixed inset-0 z-50`, dazu
      // `body { pointer-events: none }`); sonst wäre „Jetzt neu laden" — der
      // einzige Ausweg aus `gebremst` — nicht zu treffen. Alle drei sind nötig:
      // `z-index` wirkt nur positioniert, die Zeigersperre hebt allein
      // `pointer-events-auto` auf.
      className="relative z-[60] pointer-events-auto flex flex-col items-center justify-center gap-2 bg-primary px-4 py-2 text-center text-sm text-primary-foreground print:hidden sm:flex-row"
    >
      {versionsZustand === 'wartet' ? (
        <p>
          Der Server läuft mit einer anderen Version als diese Seite. Bitte den
          laufenden Vorgang abschließen oder verwerfen — danach lädt sich die
          Seite von selbst neu.
        </p>
      ) : (
        <>
          <p>
            Der Server läuft mit einer anderen Version als diese Seite. Das
            automatische Neuladen hat nicht geklappt — bitte von Hand neu laden.
          </p>
          <Button variant="secondary" size="sm" onClick={seiteNeuLaden}>
            Jetzt neu laden
          </Button>
        </>
      )}
    </div>
  )
}
