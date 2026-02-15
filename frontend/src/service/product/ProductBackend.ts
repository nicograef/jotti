import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { type Product, ProductSchema } from './Product'

export class ProductBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getActiveProducts(): Promise<Product[]> {
    const { products } = await this.backend.post(
      'service/get-active-products',
      {},
      z.object({ products: z.array(ProductSchema) }),
    )
    return products
  }
}
