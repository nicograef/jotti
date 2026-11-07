import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { KategorieSchema as ProduktKategorieSchema } from '@/lib/produktSchemas'
import { DateStringSchema } from '@/lib/utils'

// Die Produkt-Kategorien plus die beiden Stationen ohne Produkte.
const KategorieSchema = z.enum([
  ...ProduktKategorieSchema.options,
  'kassenbeleg',
  'abholbon',
])
export type Kategorie = z.infer<typeof KategorieSchema>

const BonmodusSchema = z.enum([
  'pro_position',
  'pro_bestellung',
  'pro_stueck',
  'kein_bon',
])
export type Bonmodus = z.infer<typeof BonmodusSchema>

// Leer heißt: kein Drucker an dieser Station.
const DruckerIpSchema = z.ipv4('Ungültige IPv4-Adresse').or(z.literal(''))

// bonmodusOptionen sind die Bonmodi, die das Backend für diese Station annimmt;
// leer am Kassenbeleg, der keinen Bonmodus trägt.
const DruckstationConfigSchema = z.object({
  kategorie: KategorieSchema,
  druckerIp: DruckerIpSchema,
  bonmodus: BonmodusSchema.or(z.literal('')),
  bonmodusOptionen: z.array(BonmodusSchema),
})
export type DruckstationConfig = z.infer<typeof DruckstationConfigSchema>

// Nach mehreren Fehlversuchen (rund 5 Minuten) aufgegeben.
const FehlgeschlagenerDruckauftragSchema = z.object({
  id: z.number(),
  bonArt: z.string(),
  zielIp: z.string(),
  referenz: z.string(),
  referenzAnzeige: z.string(),
  versuche: z.number(),
  fehlerMeldung: z.string(),
  erstelltAm: DateStringSchema,
})
export type FehlgeschlagenerDruckauftrag = z.infer<
  typeof FehlgeschlagenerDruckauftragSchema
>

export function validateDruckerIp(druckerIp: string): string | null {
  const { error } = DruckerIpSchema.safeParse(druckerIp)
  return error ? (error.issues[0]?.message ?? null) : null
}

// Anzeigename je Kategorie für Stationsköpfe und Produktverwaltung; nur „Getränk" weicht (Singular) vom Produkt-Label ab.
export const STATION_KATEGORIE_LABEL: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
  kassenbeleg: 'Kassenbeleg',
  abholbon: 'Abholbon',
}

// „Bon" bleibt dem operativen Arbeitsbon vorbehalten; der Gäste-Beleg ist der
// „Kassenbeleg", der Prüf-Bon der „Testbon" (siehe docs/language.md).
const BON_ART_SUBSTANTIV: Record<string, { singular: string; plural: string }> =
  {
    arbeitsbon: { singular: 'Bon', plural: 'Bons' },
    kassenbeleg: { singular: 'Kassenbeleg', plural: 'Kassenbelege' },
    testbon: { singular: 'Testbon', plural: 'Testbons' },
  }

interface FehlBonBeschreibung {
  singular: string
  plural: string
  // Nur eine reine Arbeitsbon-Menge landet an einer Ausgabestation
  // (Küche/Theke); nur dann trifft die Küchen-Formulierung zu.
  kuecheBetroffen: boolean
}

export function beschreibeFehlBons(bonArten: string[]): FehlBonBeschreibung {
  const eindeutigeArten = new Set(bonArten)
  const art = bonArten[0]
  if (
    eindeutigeArten.size === 1 &&
    Object.prototype.hasOwnProperty.call(BON_ART_SUBSTANTIV, art)
  ) {
    return { ...BON_ART_SUBSTANTIV[art], kuecheBetroffen: art === 'arbeitsbon' }
  }
  return { singular: 'Bon', plural: 'Bons', kuecheBetroffen: false }
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

  public async updateDruckstation({
    kategorie,
    druckerIp,
    bonmodus,
  }: DruckstationConfig): Promise<void> {
    await this.backend.post('admin/update-druckstationen', {
      kategorie,
      druckerIp,
      bonmodus,
    })
  }

  // Ohne konfigurierten Drucker antwortet das Backend mit dem Fehlercode
  // druckstation_nicht_konfiguriert.
  public async testbonDrucken(kategorie: Kategorie): Promise<void> {
    await this.backend.post('admin/testbon-drucken', { kategorie })
  }

  public async getFehlgeschlageneDruckauftraege(): Promise<
    FehlgeschlagenerDruckauftrag[]
  > {
    const { druckauftraege } = await this.backend.post(
      'admin/get-fehlgeschlagene-druckauftraege',
      {},
      z.object({
        druckauftraege: z.array(FehlgeschlagenerDruckauftragSchema),
      }),
    )
    return druckauftraege
  }

  public async druckauftragErneutVersuchen(id: number): Promise<void> {
    await this.backend.post('admin/druckauftrag-erneut-versuchen', { id })
  }

  public async druckauftragVerwerfen(id: number): Promise<void> {
    await this.backend.post('admin/druckauftrag-verwerfen', { id })
  }

  public async druckauftraegeVerwerfen(): Promise<number> {
    const { verworfen } = await this.backend.post(
      'admin/druckauftraege-verwerfen',
      {},
      z.object({ verworfen: z.number() }),
    )
    return verworfen
  }
}
