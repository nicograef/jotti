import { z } from 'zod'

import type { BackendClient } from './Backend'

const KategorieSchema = z.enum([
  'essen',
  'getraenk',
  'sonstiges',
  'kassenbeleg',
  'abholbon',
])
export type Kategorie = z.infer<typeof KategorieSchema>

const BonmodusSchema = z.enum(['pro_position', 'pro_bestellung'])
export type Bonmodus = z.infer<typeof BonmodusSchema>

export const DruckstationConfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: z.ipv4('Ungültige IPv4-Adresse').or(z.literal('')),
  // leer für kassenbeleg/abholbon (diese Stationen tragen keinen Bonmodus)
  bonmodus: BonmodusSchema.or(z.literal('')),
})
export type DruckstationConfig = z.infer<typeof DruckstationConfigSchema>

// Stationen mit Bonmodus: die drei Produktkategorien sowie der Abholbon werden
// wahlweise pro Position oder pro Bestellung gedruckt. Nur der Kassenbeleg nicht.
const KATEGORIEN_MIT_BONMODUS: Kategorie[] = [
  'essen',
  'getraenk',
  'sonstiges',
  'abholbon',
]

export function hatBonmodus(kategorie: Kategorie): boolean {
  return KATEGORIEN_MIT_BONMODUS.includes(kategorie)
}

// validateDruckerIp prüft eine Drucker-IP für die Inline-Feldvalidierung.
// Leer ist erlaubt (kein Drucker); andernfalls muss es eine IPv4-Adresse sein.
// Gibt eine Fehlermeldung zurück oder null, wenn gültig.
export function validateDruckerIp(druckerIp: string): string | null {
  if (druckerIp === '') {
    return null
  }
  return z.ipv4().safeParse(druckerIp).success ? null : 'Ungültige IPv4-Adresse'
}

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
    await this.backend.post('admin/update-druckstationen', config)
  }
}
