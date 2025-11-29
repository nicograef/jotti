import { z } from 'zod'

import {
  type Product,
  type ProductPublic,
  ProductPublicSchema,
  ProductSchema,
} from './Product'

export const CreateProductRequestSchema = ProductSchema.pick({
  name: true,
  description: true,
  netPriceCents: true,
  category: true,
})

export const UpdateProductRequestSchema = ProductSchema.pick({
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
    newProduct: z.infer<typeof CreateProductRequestSchema>,
  ): Promise<{ product: Product }> {
    const body = CreateProductRequestSchema.parse(newProduct)
    const { product } = await this.backend.post(
      'create-product',
      body,
      z.object({ product: ProductSchema }),
    )
    return { product }
  }

  public async updateProduct(
    updatedProduct: z.infer<typeof UpdateProductRequestSchema>,
  ): Promise<Product> {
    const body = UpdateProductRequestSchema.parse(updatedProduct)
    const { product } = await this.backend.post(
      'update-product',
      body,
      z.object({ product: ProductSchema }),
    )
    return product
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
