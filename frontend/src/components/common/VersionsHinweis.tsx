import { Button } from '@/components/ui/button'
import { useVersionsGuard } from '@/hooks/use-versions-guard'
import { Seite } from '@/lib/reload'

/**
 * Versions-Handshake notice, shown only in `wartet` or `gebremst` (docs/handbuch.md §6.8); non-modal because Radix'
 * `AlertDialog` always traps focus while the notice waits for the running Vorgang. It sits in the flow above the layout,
 * since a fixed bar would cover controls (ServiceDock, Tischauswahl footer, fixed headers).
 */
export function VersionsHinweis() {
  const versionsZustand = useVersionsGuard()

  if (versionsZustand === 'aus' || versionsZustand === 'laedt') return null

  return (
    <div
      role="alert"
      // `relative z-[60] pointer-events-auto` lifts the notice above an open Radix modal (`fixed inset-0 z-50` overlay,
      // `body { pointer-events: none }`), else „Jetzt neu laden", the only way out of `gebremst`, cannot be hit.
      // All three are needed: `z-index` works only when positioned, and only `pointer-events-auto` lifts the pointer lock.
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
          <Button variant="secondary" size="sm" onClick={Seite.neuLaden}>
            Jetzt neu laden
          </Button>
        </>
      )}
    </div>
  )
}
