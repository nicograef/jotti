import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import { STAMMDATEN_AKTUALITAET_MS } from '@/lib/queryClient'

import { ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

export const AKTIVE_PRODUKTE_KEY = 'aktive-produkte'
// `isLoadingError` statt `isError`: Nur ein gescheitertes Erstladen rechtfertigt
// einen Fehlerzustand statt der Produktliste. Scheitert ein Hintergrund-Refetch,
// bleiben die zwischengespeicherten Produkte stehen; die Meldung trägt der
// zentrale Fehler-Toast aus queryClient.ts.
export function useAktiveProdukte() {
  const {
    data = [],
    isPending,
    isLoadingError,
    refetch,
  } = useQuery({
    queryKey: [AKTIVE_PRODUKTE_KEY],
    queryFn: () => produktBackend.getAktiveProdukte(),
    staleTime: STAMMDATEN_AKTUALITAET_MS,
  })
  return { produkte: data, isPending, isLoadingError, refetch }
}
