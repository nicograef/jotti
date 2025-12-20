import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Product } from './Product'
import { ProductBackend } from './ProductBackend'

const productBackend = new ProductBackend(BackendSingleton)

/** Custom hook to fetch active products from backend. */
export function useActiveProducts() {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<Product[]>([])

  useEffect(() => {
    async function fetchProducts() {
      setLoading(true)

      try {
        const products = await productBackend.getActiveProducts()
        setProducts(products)
      } catch (error) {
        console.error('Failed to fetch products:', error)
      }

      setLoading(false)
    }

    void fetchProducts()
  }, [])

  return { loading, products }
}
