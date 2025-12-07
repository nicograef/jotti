import { z } from 'zod'

import {
  type Product,
  ProductIdSchema,
  type ProductPublic,
  ProductPublicSchema,
  ProductSchema,
} from './Product'

export const CreateProductSchema = ProductSchema.pick({
  name: true,
  description: true,
  netPriceCents: true,
  category: true,
})

export const UpdateProductSchema = ProductSchema.pick({
  id: true,
  name: true,
  description: true,
  netPriceCents: true,
  category: true,
})

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

  public async createProduct(
    newProduct: z.infer<typeof CreateProductSchema>,
  ): Promise<number> {
    const body = CreateProductSchema.parse(newProduct)
    const { id } = await this.backend.post(
      'create-product',
      body,
      z.object({ id: ProductIdSchema }),
    )
    return id
  }

  public async updateProduct(
    updatedProduct: z.infer<typeof UpdateProductSchema>,
  ): Promise<void> {
    const body = UpdateProductSchema.parse(updatedProduct)
    await this.backend.post('update-product', body)
  }

  public async getAllProducts(): Promise<Product[]> {
    const { products } = await this.backend.post(
      'get-all-products',
      {},
      z.object({ products: z.array(ProductSchema) }),
    )
    return products
  }

  public async getActiveProducts(): Promise<ProductPublic[]> {
    const { products } = await this.backend.post(
      'get-active-products',
      {},
      z.object({ products: z.array(ProductPublicSchema) }),
    )
    return products
  }

  public async activateProduct(id: number): Promise<void> {
    const body = ProductSchema.pick({ id: true }).parse({ id })
    await this.backend.post('activate-product', body)
  }

  public async deactivateProduct(id: number): Promise<void> {
    const body = ProductSchema.pick({ id: true }).parse({ id })
    await this.backend.post('deactivate-product', body)
  }
}
