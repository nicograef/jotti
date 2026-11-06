import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'

import { type AktiveProdukte, ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

interface AktiveProdukteResult {
  produkte: Produkt[]
  pfandProdukte: Produkt[]
  isPending: boolean
  isError: boolean
  refetch: UseQueryResult<AktiveProdukte>['refetch']
}

const leer: AktiveProdukte = { produkte: [], pfandProdukte: [] }

export function useAktiveProdukte(): AktiveProdukteResult {
  const {
    data = leer,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['aktive-produkte'],
    queryFn: () => produktBackend.getAktiveProdukte(),
  })
  return {
    produkte: data.produkte,
    pfandProdukte: data.pfandProdukte,
    isPending,
    isError,
    refetch,
  }
}
