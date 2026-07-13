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

// Substantiv (Singular/Plural) je Bon-Art für die Fehl-Bon-Meldungen. „Bon"
// bleibt dem operativen Arbeitsbon vorbehalten; der Gäste-Beleg ist der
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
  // Nur eine reine Arbeitsbon-Menge ist tatsächlich an einer Ausgabestation
  // (Küche/Theke) gelandet; nur dann trifft die Küchen-Formulierung zu.
  kuecheBetroffen: boolean
}

// beschreibeFehlBons leitet aus den Bon-Arten fehlgeschlagener Druckaufträge das
// passende Substantiv und die Frage ab, ob die Küchen-Formulierung passt. Eine
// gemischte Menge (oder eine unbekannte Bon-Art) fällt auf den neutralen
// Oberbegriff „Bon" ohne Küchen-Behauptung zurück.
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

// Bekannte Symptome roher Relay-Fehlertexte (Go-Fehlerketten mit IP, Status-Hex,
// „dial tcp"). Reihenfolge = Priorität. Übersetzt in eine knappe, für
// ehrenamtliche Helfer verständliche Meldung.
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

// formatDruckfehler übersetzt den rohen letzterFehler eines fehlgeschlagenen
// Druckauftrags in eine einheitliche, laienverständliche Meldung. Unbekannte
// Texte fallen auf eine neutrale Sammelmeldung zurück, nie auf den Rohtext.
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
