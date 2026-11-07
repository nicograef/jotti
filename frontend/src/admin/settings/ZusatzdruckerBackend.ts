import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { createNameSchema } from '@/lib/nameSchema'

import { BonmodusSchema, DruckerIpSchema } from './DruckstationBackend'

export const ZusatzdruckerNameSchema = createNameSchema(100)

// Pro Stück und Kein Bon gehören dem Abholbon.
const ZusatzdruckerBonmodusSchema = BonmodusSchema.extract([
  'pro_position',
  'pro_bestellung',
])
export const ZUSATZDRUCKER_BONMODI = ZusatzdruckerBonmodusSchema.options

export const ZusatzdruckerSchema = z.object({
  id: z.number().int().min(1),
  name: ZusatzdruckerNameSchema,
  druckerIp: DruckerIpSchema,
  bonmodus: ZusatzdruckerBonmodusSchema,
})
export type Zusatzdrucker = z.infer<typeof ZusatzdruckerSchema>

export class ZusatzdruckerBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getZusatzdrucker(): Promise<Zusatzdrucker[]> {
    const { zusatzdrucker } = await this.backend.post(
      'admin/get-zusatzdrucker',
      {},
      z.object({ zusatzdrucker: z.array(ZusatzdruckerSchema) }),
    )
    return zusatzdrucker
  }

  // Ein neuer Zusatzdrucker startet ohne Drucker-IP und mit „Pro Position“.
  public async anlegen(name: string): Promise<number> {
    const { id } = await this.backend.post(
      'admin/create-zusatzdrucker',
      { name: ZusatzdruckerNameSchema.parse(name) },
      z.object({ id: z.number() }),
    )
    return id
  }

  public async aendern(zusatzdrucker: Zusatzdrucker): Promise<void> {
    const { id, name, druckerIp, bonmodus } =
      ZusatzdruckerSchema.parse(zusatzdrucker)
    await this.backend.post('admin/update-zusatzdrucker', {
      id,
      name,
      druckerIp,
      bonmodus,
    })
  }

  // Die zugeordneten Produkte drucken danach wieder am Drucker ihrer Kategorie.
  public async loeschen(id: number): Promise<void> {
    await this.backend.post('admin/delete-zusatzdrucker', { id })
  }

  // Ohne Drucker-IP antwortet das Backend mit zusatzdrucker_nicht_konfiguriert.
  public async testbonDrucken(id: number): Promise<void> {
    await this.backend.post('admin/zusatzdrucker-testbon-drucken', { id })
  }
}
