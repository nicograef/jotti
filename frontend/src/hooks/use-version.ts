import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { HealthBackend } from '@/lib/HealthBackend'
import { OHNE_FEHLER_TOAST } from '@/lib/queryClient'

const healthBackend = new HealthBackend(BackendSingleton)

// Gleichauf mit der einzigen anderen Polling-Stelle (src/admin/reporting/hooks.ts).
// /health pingt bei jedem Aufruf die Datenbank und wird mit dreißig
// Helfer-Handys der meistgerufene Endpunkt des Systems — dreißig Sekunden sind
// für Postgres belanglos und für einen Versionswechsel schnell genug.
export const VERSIONSABFRAGE_INTERVALL_MS = 30_000

/**
 * Laufende Backend-Version (z. B. "v1.0.0"), undefined bis zur ersten Antwort.
 *
 * Kein `staleTime`: Nur daran hängt das Nachholen beim Zurückkehren in den
 * Vordergrund — react-query holt allein eine veraltete Abfrage nach.
 * Ein Fehlschlag bleibt stumm (kein globaler Fehler-Toast): Die Abfrage läuft
 * im Hintergrund, und niemand kann auf sie reagieren.
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
