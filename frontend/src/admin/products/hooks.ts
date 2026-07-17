import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { Produkt } from './Produkt'
import { ProduktBackend } from './ProduktBackend'

export const produktBackend = new ProduktBackend(BackendSingleton)

export const ALLE_PRODUKTE_KEY = 'alle-produkte'

export function useAllProdukte() {
  const { data: produkte = [] as Produkt[], isPending } = useQuery({
    queryKey: [ALLE_PRODUKTE_KEY],
    queryFn: () => produktBackend.getAllProdukte(),
  })
  return { produkte, isPending }
}
