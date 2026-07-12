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

// Fehlgeschlagener Druckauftrag: nach mehreren Fehlversuchen (rund 5 Minuten)
// aufgegeben. Wird auf der Druckstationen-Seite zur Verwaltung (erneut
// versuchen / verwerfen) angezeigt.
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

// validateDruckerIp prüft eine Drucker-IP für die Inline-Feldvalidierung.
// Leer ist erlaubt (kein Drucker); andernfalls muss es eine IPv4-Adresse sein.
// Gibt eine Fehlermeldung zurück oder null, wenn gültig.
export function validateDruckerIp(druckerIp: string): string | null {
  if (druckerIp === '') {
    return null
  }
  return z.ipv4().safeParse(druckerIp).success ? null : 'Ungültige IPv4-Adresse'
}

// Referenz-Formate aus dem Backend (unverändert, siehe arbeitsbon_policy.go,
// kassenbeleg_command.go und station/application/command.go):
// "<technischer-event-name>:<eventId>" bzw. für Testbons "testdruck:<kategorie>".
const REFERENZ_PRAEFIX_LABEL: Record<string, string> = {
  'bestellung-aufgenommen': 'Bestellung',
  'zahlung-kassiert': 'Zahlung',
  'direktverkauf-getaetigt': 'Direktverkauf',
  'direktverkauf-storniert': 'Direktverkauf-Storno',
  'stornierung-erteilt': 'Stornierung',
}

// Fachlicher Anzeigename je Kategorie, geteilt von der Referenz-Anzeige
// fehlgeschlagener Druckaufträge (unten) und den Stationsköpfen der
// Bondrucker-Seite. Testbons tragen als Referenz "testdruck:<kategorie>".
export const KATEGORIE_LABEL: Record<string, string> = {
  essen: 'Essen',
  getraenk: 'Getränk',
  sonstiges: 'Sonstiges',
  kassenbeleg: 'Kassenbeleg',
  abholbon: 'Abholbon',
}

// formatDruckauftragReferenz übersetzt die rohe Referenz eines fehlgeschlagenen
// Druckauftrags in einen fachlichen Text. Unbekannte Formate fallen auf den
// Rohwert zurück (z. B. bei künftigen, hier noch nicht gepflegten Event-Typen).
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
    return `Testbon ${KATEGORIE_LABEL[rest] ?? rest}`
  }
  const label = REFERENZ_PRAEFIX_LABEL[praefix] ?? ''
  if (label === '') {
    return referenz
  }
  return `${label} Nr. ${rest}`
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

  // Reiht einen Testbon (Stationsname + Zeitstempel) für die Kategorie in die
  // Druck-Warteschlange ein. Ohne konfigurierten Drucker antwortet das Backend
  // mit dem Fehlercode druckstation_nicht_konfiguriert.
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

  // Verwirft alle fehlgeschlagenen Druckaufträge und gibt die Anzahl der
  // verworfenen Aufträge zurück (für die Rückmeldung im Toast).
  public async druckauftraegeVerwerfen(): Promise<number> {
    const { verworfen } = await this.backend.post(
      'admin/druckauftraege-verwerfen',
      {},
      z.object({ verworfen: z.number() }),
    )
    return verworfen
  }
}
