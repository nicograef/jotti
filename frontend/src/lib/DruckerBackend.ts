import { z } from 'zod'

import type { BackendClient } from './Backend'

const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])

const BonmodusSchema = z.enum(['pro_position', 'pro_bestellung'])
export type Bonmodus = z.infer<typeof BonmodusSchema>

export const DruckerConfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: z
    .string()
    .regex(/^(\d{1,3}\.){3}\d{1,3}$/, 'Ungültige IPv4-Adresse')
    .or(z.literal('')),
  bonmodus: BonmodusSchema,
})
export type DruckerConfig = z.infer<typeof DruckerConfigSchema>

export class DruckerBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getDruckerConfig(): Promise<DruckerConfig[]> {
    const { drucker } = await this.backend.post(
      'admin/get-drucker-konfiguration',
      {},
      z.object({ drucker: z.array(DruckerConfigSchema) }),
    )
    return drucker
  }

  public async updateDruckerConfig(config: DruckerConfig): Promise<void> {
    const body = DruckerConfigSchema.parse(config)
    await this.backend.post('admin/update-drucker-konfiguration', body)
  }
}
