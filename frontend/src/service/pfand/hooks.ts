import { useQuery } from '@tanstack/react-query'

import type { QueryResult } from '@/lib/queryClient'

import type { PfandrueckgabeHistorieEintrag } from './Pfand'
import { pfandBackend } from './PfandBackend'

// Die Seiten invalidieren diesen Schlüssel nach jedem erfolgreichen Vorgang.
export const PFANDRUECKGABE_HISTORIE_KEY = 'pfandrueckgabe-historie'

export function usePfandrueckgabeHistorie(tischId: number): QueryResult<
  PfandrueckgabeHistorieEintrag[]
> & {
  historie: PfandrueckgabeHistorieEintrag[]
} {
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
