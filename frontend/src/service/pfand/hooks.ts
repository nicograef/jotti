import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import type { PfandrueckgabeHistorieEintrag } from './Pfand'
import { pfandBackend } from './PfandBackend'

// Die Seiten invalidieren diesen Schlüssel nach jedem erfolgreichen Vorgang.
export const PFANDRUECKGABE_HISTORIE_KEY = 'pfandrueckgabe-historie'

interface PfandrueckgabeHistorieResult {
  historie: PfandrueckgabeHistorieEintrag[]
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<PfandrueckgabeHistorieEintrag[]>['refetch']
}

export function usePfandrueckgabeHistorie(
  tischId: number,
): PfandrueckgabeHistorieResult {
  const {
    data: historie = [],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: [PFANDRUECKGABE_HISTORIE_KEY, tischId],
    queryFn: () => pfandBackend.getPfandrueckgabeHistorie(tischId),
  })
  return { historie, isPending, isError, refetch }
}
