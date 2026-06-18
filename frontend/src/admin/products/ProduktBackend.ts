import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Produkt,
  ProduktIdSchema,
  ProduktSchema,
  VarianteIdSchema,
  VarianteSchema,
} from './Produkt'

export const CreateProduktSchema = ProduktSchema.pick({
  name: true,
  kategorie: true,
  steuersatz: true,
})

export const UpdateProduktSchema = ProduktSchema.pick({
  id: true,
  name: true,
  kategorie: true,
  steuersatz: true,
})

export const CreateVarianteSchema = z.object({
  produktId: ProduktIdSchema,
  name: VarianteSchema.shape.name,
  preisCents: VarianteSchema.shape.preisCents,
})

export const UpdateVarianteSchema = VarianteSchema.pick({
  id: true,
  name: true,
  preisCents: true,
})

export class ProduktBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async createProdukt(
    newProdukt: z.infer<typeof CreateProduktSchema>,
  ): Promise<number> {
    const body = CreateProduktSchema.parse(newProdukt)
    const { id } = await this.backend.post(
      'admin/create-produkt',
      body,
      z.object({ id: ProduktIdSchema }),
    )
    return id
  }

  public async updateProdukt(
    updatedProdukt: z.infer<typeof UpdateProduktSchema>,
  ): Promise<void> {
    const body = UpdateProduktSchema.parse(updatedProdukt)
    await this.backend.post('admin/update-produkt', body)
  }

  public async getAllProdukte(): Promise<Produkt[]> {
    const { produkte } = await this.backend.post(
      'admin/get-all-produkte',
      {},
      z.object({ produkte: z.array(ProduktSchema) }),
    )
    return produkte
  }

  public async createVariante(
    newVariante: z.infer<typeof CreateVarianteSchema>,
  ): Promise<number> {
    const body = CreateVarianteSchema.parse(newVariante)
    const { id } = await this.backend.post(
      'admin/create-variante',
      body,
      z.object({ id: VarianteIdSchema }),
    )
    return id
  }

  public async updateVariante(
    updatedVariante: z.infer<typeof UpdateVarianteSchema>,
  ): Promise<void> {
    const body = UpdateVarianteSchema.parse(updatedVariante)
    await this.backend.post('admin/update-variante', body)
  }

  public async aktiviereVariante(id: number): Promise<void> {
    const body = { id: VarianteIdSchema.parse(id) }
    await this.backend.post('admin/activate-variante', body)
  }

  public async deaktiviereVariante(id: number): Promise<void> {
    const body = { id: VarianteIdSchema.parse(id) }
    await this.backend.post('admin/deactivate-variante', body)
  }

  public async deleteProdukt(id: number): Promise<void> {
    const body = { id: ProduktIdSchema.parse(id) }
    await this.backend.post('admin/delete-produkt', body)
  }

  public async deleteVariante(produktId: number, id: number): Promise<void> {
    const body = {
      produktId: ProduktIdSchema.parse(produktId),
      id: VarianteIdSchema.parse(id),
    }
    await this.backend.post('admin/delete-variante', body)
  }
}
