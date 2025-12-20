import { z } from 'zod'

import { type Product, ProductSchema } from './Product'

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class ProductBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
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
