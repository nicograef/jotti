import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { DirektverkaufBackend } from './DirektverkaufBackend'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export const DIREKTVERKAUF_HISTORIE_KEY = 'direktverkauf-historie'
// `isLoadingError` statt `isError`: Nur ein gescheitertes Erstladen rechtfertigt
// einen Fehlerzustand statt der Historie. Scheitert ein Hintergrund-Refetch,
// bleibt die zwischengespeicherte Historie stehen; die Meldung trägt der
// zentrale Fehler-Toast aus queryClient.ts.
export function useDirektverkaufHistorie() {
  const {
    data: historie = [],
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [DIREKTVERKAUF_HISTORIE_KEY],
    queryFn: () => direktverkaufBackend.getDirektverkaufHistorie(),
  })
  return { historie, isPending, isLoadingError, refetch }
}
