import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { HealthBackend } from '@/lib/HealthBackend'
import { OHNE_FEHLER_TOAST } from '@/lib/queryClient'

const healthBackend = new HealthBackend(BackendSingleton)

// Gleichauf mit der einzigen anderen Polling-Stelle des Repos
// (src/admin/reporting/hooks.ts). /health pingt bei jedem Aufruf die Datenbank
// und wird mit dreißig Helfer-Handys der meistgerufene Endpunkt des Systems —
// dreißig Sekunden sind für Postgres belanglos und für einen Versionswechsel
// schnell genug.
export const VERSIONSABFRAGE_INTERVALL_MS = 30_000

/**
 * Liefert die laufende Backend-Version (z. B. "v1.0.0") oder undefined,
 * solange sie noch nicht geladen ist.
 *
 * Die Abfrage läuft regelmäßig weiter und wiederholt sich beim Zurückkehren in
 * den Vordergrund, damit ein Serverwechsel ohne Neuladen der Seite auffällt.
 * Das Nachholen im Vordergrund hängt allein daran, dass hier kein `staleTime`
 * gesetzt ist — react-query holt nur eine veraltete Abfrage nach.
 * Ein Fehlschlag bleibt bewusst stumm (kein globaler Fehler-Toast): Die Abfrage
 * läuft im Hintergrund, und niemand kann auf sie reagieren.
 */
export function useVersion(): string | undefined {
  const { data } = useQuery({
    queryKey: ['version'],
    queryFn: () => healthBackend.getVersion(),
    refetchInterval: VERSIONSABFRAGE_INTERVALL_MS,
    meta: OHNE_FEHLER_TOAST,
  })
  return data
}
