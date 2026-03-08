import { z } from 'zod'

import {
  type Product,
  ProductIdSchema,
  ProductSchema,
  VariantIdSchema,
  VariantSchema,
} from './Product'

export const CreateProductSchema = ProductSchema.pick({
  name: true,
  category: true,
})

export const UpdateProductSchema = ProductSchema.pick({
  id: true,
  name: true,
  category: true,
})

export const CreateVariantSchema = z.object({
  productId: ProductIdSchema,
  name: VariantSchema.shape.name,
  priceCents: VariantSchema.shape.priceCents,
})

export const UpdateVariantSchema = VariantSchema.pick({
  id: true,
  name: true,
  priceCents: true,
})

import type { BackendClient } from '@/lib/Backend'

export class ProductBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  // Product methods

  public async createProduct(
    newProduct: z.infer<typeof CreateProductSchema>,
  ): Promise<number> {
    const body = CreateProductSchema.parse(newProduct)
    const { id } = await this.backend.post(
      'admin/create-produkt',
      body,
      z.object({ id: ProductIdSchema }),
    )
    return id
  }

  public async updateProduct(
    updatedProduct: z.infer<typeof UpdateProductSchema>,
  ): Promise<void> {
    const body = UpdateProductSchema.parse(updatedProduct)
    await this.backend.post('admin/update-produkt', body)
  }

  public async getAllProducts(): Promise<Product[]> {
    const { produkte } = await this.backend.post(
      'admin/get-all-produkte',
      {},
      z.object({ produkte: z.array(ProductSchema) }),
    )
    return produkte
  }

  // Variant methods

  public async createVariant(
    newVariant: z.infer<typeof CreateVariantSchema>,
  ): Promise<number> {
    const body = CreateVariantSchema.parse(newVariant)
    const { id } = await this.backend.post(
      'admin/create-variante',
      body,
      z.object({ id: VariantIdSchema }),
    )
    return id
  }

  public async updateVariant(
    updatedVariant: z.infer<typeof UpdateVariantSchema>,
  ): Promise<void> {
    const body = UpdateVariantSchema.parse(updatedVariant)
    await this.backend.post('admin/update-variante', body)
  }

  public async activateVariant(id: number): Promise<void> {
    const body = { id: VariantIdSchema.parse(id) }
    await this.backend.post('admin/activate-variante', body)
  }

  public async deactivateVariant(id: number): Promise<void> {
    const body = { id: VariantIdSchema.parse(id) }
    await this.backend.post('admin/deactivate-variante', body)
  }
}
