import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { HealthBackend } from '@/lib/HealthBackend'

const healthBackend = new HealthBackend(BackendSingleton)

// Liefert die laufende Backend-Version (z. B. "v1.0.0") oder undefined,
// solange sie noch nicht geladen ist. Die Version ändert sich nur mit
// einem Update, daher kein Refetch innerhalb der Session.
export function useVersion(): string | undefined {
  const { data } = useQuery({
    queryKey: ['version'],
    queryFn: () => healthBackend.getVersion(),
    staleTime: Infinity,
  })
  return data
}
