import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Product } from './Product'
import { ProductBackend } from './ProductBackend'

const productBackend = new ProductBackend(BackendSingleton)

/** Custom hook to fetch all products from backend. */
export function useAllProducts() {
  const {
    data: products,
    setData: setProducts,
    ...rest
  } = useFetch(() => productBackend.getAllProducts(), [] as Product[])
  return { ...rest, products, setProducts }
}
