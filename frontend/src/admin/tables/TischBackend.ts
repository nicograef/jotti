import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { type Tisch, TischIdSchema, TischSchema } from './Tisch'

export const CreateTischSchema = TischSchema.pick({
  name: true,
})

export const UpdateTischSchema = TischSchema.pick({
  id: true,
  name: true,
})

export class TischBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getAllTische(): Promise<Tisch[]> {
    const { tische } = await this.backend.post(
      'admin/get-all-tische',
      {},
      z.object({ tische: z.array(TischSchema) }),
    )
    return tische
  }

  public async createTisch(
    newTisch: z.infer<typeof CreateTischSchema>,
  ): Promise<number> {
    const body = CreateTischSchema.parse(newTisch)
    const { id } = await this.backend.post(
      'admin/create-tisch',
      body,
      z.object({ id: TischIdSchema }),
    )
    return id
  }

  public async updateTisch(
    updatedTisch: z.infer<typeof UpdateTischSchema>,
  ): Promise<void> {
    const body = UpdateTischSchema.parse(updatedTisch)
    await this.backend.post('admin/update-tisch', body)
  }

  public async activateTisch(id: number): Promise<void> {
    const body = TischSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/activate-tisch', body)
  }

  public async deactivateTisch(id: number): Promise<void> {
    const body = TischSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/deactivate-tisch', body)
  }

  public async deleteTisch(id: number): Promise<void> {
    const body = TischSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/delete-tisch', body)
  }
}
