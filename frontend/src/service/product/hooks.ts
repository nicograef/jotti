import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Produkt } from './Product'
import { ProductBackend } from './ProductBackend'

const productBackend = new ProductBackend(BackendSingleton)

/** Custom hook to fetch active products from backend. */
export function useActiveProducts() {
  const { data: products, ...rest } = useFetch(
    () => productBackend.getActiveProducts(),
    [] as Produkt[],
  )
  return { ...rest, products }
}
