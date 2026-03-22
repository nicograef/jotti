import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Produkt } from './Produkt'
import { ProduktBackend } from './ProduktBackend'

const produktBackend = new ProduktBackend(BackendSingleton)

/** Custom hook to fetch all products from backend. */
export function useAllProdukte() {
  const {
    data: produkte,
    setData: setProdukte,
    ...rest
  } = useFetch(() => produktBackend.getAllProdukte(), [] as Produkt[])
  return { ...rest, produkte, setProdukte }
}
