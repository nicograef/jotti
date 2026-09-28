import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'

import { ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

interface AktiveProdukteResult {
  produkte: Produkt[]
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<Produkt[]>['refetch']
}

export function useAktiveProdukte(): AktiveProdukteResult {
  const {
    data = [],
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['aktive-produkte'],
    queryFn: () => produktBackend.getAktiveProdukte(),
  })
  return { produkte: data, isPending, isError, refetch }
}
