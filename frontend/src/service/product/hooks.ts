import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Produkt } from './Produkt'
import { ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

/** Custom hook to fetch active products from backend. */
export function useAktiveProdukte() {
  const { data: produkte, ...rest } = useFetch(
    () => produktBackend.getAktiveProdukte(),
    [] as Produkt[],
  )
  return { ...rest, produkte }
}
