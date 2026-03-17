import { z } from 'zod'

import type { BackendClient } from './Backend'

export const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])
export type Kategorie = z.infer<typeof KategorieSchema>

export const BonmodusSchema = z.enum(['pro_position', 'pro_bestellung'])
export type Bonmodus = z.infer<typeof BonmodusSchema>

export const DruckerKonfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: z.string(),
  bonmodus: BonmodusSchema,
})
export type DruckerKonfig = z.infer<typeof DruckerKonfigSchema>

export const UpdateDruckerConfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: z
    .string()
    .regex(/^(\d{1,3}\.){3}\d{1,3}$/, 'Ungültige IPv4-Adresse')
    .or(z.literal('')),
  bonmodus: BonmodusSchema,
})
export type UpdateDruckerConfig = z.infer<typeof UpdateDruckerConfigSchema>

export class DruckerBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getDruckerConfig(): Promise<DruckerKonfig[]> {
    const { drucker } = await this.backend.post(
      'admin/get-drucker-config',
      {},
      z.object({ drucker: z.array(DruckerKonfigSchema) }),
    )
    return drucker
  }

  public async updateDruckerConfig(config: UpdateDruckerConfig): Promise<void> {
    const body = UpdateDruckerConfigSchema.parse(config)
    await this.backend.post('admin/update-drucker-config', body)
  }
}
