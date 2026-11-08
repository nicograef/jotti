import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { createNameSchema } from '@/lib/nameSchema'
import { KategorieSchema, ProduktIdSchema } from '@/lib/produktSchemas'

import { BonmodusSchema, DruckerIpSchema } from './DruckstationBackend'

export const ZusatzdruckerNameSchema = createNameSchema(100)

// Pro Stück und Kein Bon gehören dem Abholbon.
const ZusatzdruckerBonmodusSchema = BonmodusSchema.extract([
  'pro_position',
  'pro_bestellung',
])
export const ZUSATZDRUCKER_BONMODI = ZusatzdruckerBonmodusSchema.options

const ZusatzdruckerIdSchema = z.number().int().min(1)

const ZugeordnetesProduktSchema = z.object({
  id: ProduktIdSchema,
  name: z.string(),
})

export const ZusatzdruckerSchema = z.object({
  id: ZusatzdruckerIdSchema,
  name: ZusatzdruckerNameSchema,
  druckerIp: DruckerIpSchema,
  bonmodus: ZusatzdruckerBonmodusSchema,
  produkte: z.array(ZugeordnetesProduktSchema),
})
export type Zusatzdrucker = z.infer<typeof ZusatzdruckerSchema>

// Alle nicht gelöschten Produkte; zusatzdruckerName ist leer ohne Zuordnung.
const ProduktAuswahlSchema = z.object({
  id: ProduktIdSchema,
  name: z.string(),
  kategorie: KategorieSchema,
  zusatzdruckerName: z.string(),
})
export type ProduktAuswahl = z.infer<typeof ProduktAuswahlSchema>

const ZusatzdruckerUebersichtSchema = z.object({
  zusatzdrucker: z.array(ZusatzdruckerSchema),
  produktAuswahl: z.array(ProduktAuswahlSchema),
})
export type ZusatzdruckerUebersicht = z.infer<
  typeof ZusatzdruckerUebersichtSchema
>

export class ZusatzdruckerBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getZusatzdrucker(): Promise<ZusatzdruckerUebersicht> {
    return this.backend.post(
      'admin/get-zusatzdrucker',
      {},
      ZusatzdruckerUebersichtSchema,
    )
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

  // Eine bestehende Zuordnung des Produkts wandert zu diesem Drucker.
  public async produktZuordnen(
    produktId: number,
    zusatzdruckerId: number,
  ): Promise<void> {
    await this.backend.post('admin/assign-produkt-zusatzdrucker', {
      produktId: ProduktIdSchema.parse(produktId),
      zusatzdruckerId: ZusatzdruckerIdSchema.parse(zusatzdruckerId),
    })
  }

  public async zuordnungEntfernen(produktId: number): Promise<void> {
    await this.backend.post('admin/unassign-produkt-zusatzdrucker', {
      produktId: ProduktIdSchema.parse(produktId),
    })
  }
}
