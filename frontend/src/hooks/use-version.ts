import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { HealthBackend } from '@/lib/HealthBackend'
import { OHNE_FEHLER_TOAST } from '@/lib/queryClient'

const healthBackend = new HealthBackend(BackendSingleton)

// Same interval as the only other polling site (src/admin/reporting/hooks.ts). /health pings the DB and is
// the most-called endpoint with thirty phones; 30 s is trivial for Postgres and fast enough for a version change.
export const VERSIONSABFRAGE_INTERVALL_MS = 30_000

/**
 * Running backend version (e.g. "v1.0.0"), undefined until the first response; no `staleTime`, so react-query refetches on return to the foreground.
 * A failure raises no global error toast, since nobody can act on a background query.
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
