import { z } from 'zod'

import type { BackendClient } from './Backend'

const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])

const BonmodusSchema = z.enum(['pro_position', 'pro_bestellung'])
export type Bonmodus = z.infer<typeof BonmodusSchema>

export const DruckstationConfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: z.ipv4('Ungültige IPv4-Adresse').or(z.literal('')),
  bonmodus: BonmodusSchema,
})
export type DruckstationConfig = z.infer<typeof DruckstationConfigSchema>

export class DruckstationBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getDruckstationen(): Promise<DruckstationConfig[]> {
    const { druckstationen } = await this.backend.post(
      'admin/get-druckstationen',
      {},
      z.object({ druckstationen: z.array(DruckstationConfigSchema) }),
    )
    return druckstationen
  }

  public async updateDruckstation(config: DruckstationConfig): Promise<void> {
    const body = DruckstationConfigSchema.parse(config)
    await this.backend.post('admin/update-druckstationen', body)
  }
}
