import { z } from 'zod'

import {
  type Produkt,
  ProduktIdSchema,
  ProduktSchema,
  VarianteIdSchema,
  VarianteSchema,
} from './Product'

export const CreateProductSchema = ProduktSchema.pick({
  name: true,
  category: true,
})

export const UpdateProductSchema = ProduktSchema.pick({
  id: true,
  name: true,
  category: true,
})

export const CreateVarianteSchema = z.object({
  productId: ProduktIdSchema,
  name: VarianteSchema.shape.name,
  preisCents: VarianteSchema.shape.preisCents,
})

export const UpdateVarianteSchema = VarianteSchema.pick({
  id: true,
  name: true,
  preisCents: true,
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
      z.object({ id: ProduktIdSchema }),
    )
    return id
  }

  public async updateProduct(
    updatedProduct: z.infer<typeof UpdateProductSchema>,
  ): Promise<void> {
    const body = UpdateProductSchema.parse(updatedProduct)
    await this.backend.post('admin/update-produkt', body)
  }

  public async getAllProducts(): Promise<Produkt[]> {
    const { produkte } = await this.backend.post(
      'admin/get-all-produkte',
      {},
      z.object({ produkte: z.array(ProduktSchema) }),
    )
    return produkte
  }

  // Variant methods

  public async createVariant(
    newVariant: z.infer<typeof CreateVarianteSchema>,
  ): Promise<number> {
    const body = CreateVarianteSchema.parse(newVariant)
    const { id } = await this.backend.post(
      'admin/create-variante',
      body,
      z.object({ id: VarianteIdSchema }),
    )
    return id
  }

  public async updateVariant(
    updatedVariant: z.infer<typeof UpdateVarianteSchema>,
  ): Promise<void> {
    const body = UpdateVarianteSchema.parse(updatedVariant)
    await this.backend.post('admin/update-variante', body)
  }

  public async activateVariant(id: number): Promise<void> {
    const body = { id: VarianteIdSchema.parse(id) }
    await this.backend.post('admin/activate-variante', body)
  }

  public async deactivateVariant(id: number): Promise<void> {
    const body = { id: VarianteIdSchema.parse(id) }
    await this.backend.post('admin/deactivate-variante', body)
  }
}
