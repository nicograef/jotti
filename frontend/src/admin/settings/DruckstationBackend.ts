import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { DateStringSchema } from '@/lib/utils'

const KategorieSchema = z.enum([
  'essen',
  'getraenk',
  'sonstiges',
  'kassenbeleg',
  'abholbon',
])
export type Kategorie = z.infer<typeof KategorieSchema>

const BonmodusSchema = z.enum(['pro_position', 'pro_bestellung', 'pro_stueck'])
export type Bonmodus = z.infer<typeof BonmodusSchema>

// Pro Stück (je Einheit ein Bon) nur am Abholbon — spiegelt
// Kategorie.ErlaubtBonmodus im Backend. Ob eine Station überhaupt einen
// Bonmodus trägt, sagt hatBonmodus.
export function erlaubtBonmodus(
  kategorie: Kategorie,
  bonmodus: Bonmodus,
): boolean {
  return bonmodus !== 'pro_stueck' || kategorie === 'abholbon'
}

export const DruckstationConfigSchema = z
  .object({
    kategorie: KategorieSchema,
    druckerIp: z.ipv4('Ungültige IPv4-Adresse').or(z.literal('')),
    // leer nur für den Kassenbeleg, der keinen Bonmodus trägt
    bonmodus: BonmodusSchema.or(z.literal('')),
  })
  .superRefine((config, ctx) => {
    if (
      config.bonmodus === '' ||
      erlaubtBonmodus(config.kategorie, config.bonmodus)
    ) {
      return
    }
    ctx.addIssue({
      code: 'custom',
      path: ['bonmodus'],
      message: 'Pro Stück ist nur für den Abholbon zulässig',
    })
  })
export type DruckstationConfig = z.infer<typeof DruckstationConfigSchema>

const KATEGORIEN_MIT_BONMODUS: Kategorie[] = [
  'essen',
  'getraenk',
  'sonstiges',
  'abholbon',
]

export function hatBonmodus(kategorie: Kategorie): boolean {
  return KATEGORIEN_MIT_BONMODUS.includes(kategorie)
}

// Nach mehreren Fehlversuchen (rund 5 Minuten) aufgegeben.
export const FehlgeschlagenerDruckauftragSchema = z.object({
  id: z.number(),
  bonArt: z.string(),
  zielIp: z.string(),
  referenz: z.string(),
  versuche: z.number(),
  letzterFehler: z.string(),
  erstelltAm: DateStringSchema,
})
export type FehlgeschlagenerDruckauftrag = z.infer<
  typeof FehlgeschlagenerDruckauftragSchema
>

export function validateDruckerIp(druckerIp: string): string | null {
  if (druckerIp === '') {
    return null
  }
  return z.ipv4().safeParse(druckerIp).success ? null : 'Ungültige IPv4-Adresse'
}

// Referenz-Formate aus dem Backend (arbeitsbon_policy.go,
// kassenbeleg_command.go, station/application/command.go):
// "<technischer-event-name>:<eventId>", für Testbons "testdruck:<kategorie>".
const REFERENZ_PRAEFIX_LABEL: Record<string, string> = {
  'bestellung-aufgenommen': 'Bestellung',
  'zahlung-kassiert': 'Zahlung',
  'direktverkauf-getaetigt': 'Direktverkauf',
  'direktverkauf-storniert': 'Direktverkauf-Storno',
  'stornierung-erteilt': 'Stornierung',
}

// Anzeigename je Kategorie für Stationsköpfe, Referenz-Anzeige und
// Produktverwaltung; nur „Getränk" weicht (Singular) vom Produkt-Label ab.
export const STATION_KATEGORIE_LABEL: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
  kassenbeleg: 'Kassenbeleg',
  abholbon: 'Abholbon',
}

export function formatDruckauftragReferenz(referenz: string): string {
  const trennerIndex = referenz.indexOf(':')
  if (trennerIndex === -1) {
    return referenz
  }
  const praefix = referenz.slice(0, trennerIndex)
  const rest = referenz.slice(trennerIndex + 1)
  if (rest.length === 0) {
    return referenz
  }
  if (praefix === 'testdruck') {
    const kategorie = KategorieSchema.safeParse(rest)
    return `Testbon ${kategorie.success ? STATION_KATEGORIE_LABEL[kategorie.data] : rest}`
  }
  const label = REFERENZ_PRAEFIX_LABEL[praefix] ?? ''
  if (label === '') {
    return referenz
  }
  return `${label} Nr. ${rest}`
}

// „Bon" bleibt dem operativen Arbeitsbon vorbehalten; der Gäste-Beleg ist der
// „Kassenbeleg", der Prüf-Bon der „Testbon" (siehe docs/language.md).
const BON_ART_SUBSTANTIV: Record<string, { singular: string; plural: string }> =
  {
    arbeitsbon: { singular: 'Bon', plural: 'Bons' },
    kassenbeleg: { singular: 'Kassenbeleg', plural: 'Kassenbelege' },
    testbon: { singular: 'Testbon', plural: 'Testbons' },
  }

export interface FehlBonBeschreibung {
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

// Bekannte Symptome roher Relay-Fehlertexte (Go-Fehlerketten mit IP,
// Status-Hex, „dial tcp"). Die Reihenfolge ist die Priorität.
const DRUCKFEHLER_MELDUNGEN: { schluessel: string; meldung: string }[] = [
  { schluessel: 'papier', meldung: 'Papier leer' },
  { schluessel: 'abdeckung', meldung: 'Abdeckung offen' },
  { schluessel: 'nicht erreichbar', meldung: 'Drucker nicht erreichbar' },
  {
    schluessel: 'senden fehlgeschlagen',
    meldung: 'Übertragung fehlgeschlagen',
  },
]

const DRUCKFEHLER_FALLBACK = 'Druckfehler'

// Unbekannte Texte fallen auf eine Sammelmeldung zurück, nie auf den Rohtext.
export function formatDruckfehler(letzterFehler: string): string {
  const text = letzterFehler.toLowerCase()
  const treffer = DRUCKFEHLER_MELDUNGEN.find((eintrag) =>
    text.includes(eintrag.schluessel),
  )
  return treffer?.meldung ?? DRUCKFEHLER_FALLBACK
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
