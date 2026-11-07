import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { QueryResult } from '@/lib/queryClient'

import type { PfandrueckgabeHistorieEintrag } from './Pfand'
import { PfandBackend } from './PfandBackend'

export const pfandBackend = new PfandBackend(BackendSingleton)

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
