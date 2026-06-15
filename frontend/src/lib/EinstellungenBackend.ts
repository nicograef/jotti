import { z } from 'zod'

import type { BackendClient } from './Backend'
import { DateStringSchema } from './utils'

export const BetreiberSchema = z.object({
  vereinsname: z.string(),
  strasse: z.string(),
  plz: z.string(),
  ort: z.string(),
  steuernummer: z.string().nullable(),
  ustId: z.string().nullable(),
})
export type Betreiber = z.infer<typeof BetreiberSchema>

const KassenidentitaetSchema = z.object({
  seriennummer: z.uuid(),
  angelegtAm: DateStringSchema,
})
export type Kassenidentitaet = z.infer<typeof KassenidentitaetSchema>

export const TSEKonfigurationSchema = z.object({
  apiKeyGesetzt: z.boolean(),
  apiSecretGesetzt: z.boolean(),
  tssId: z.string(),
  clientId: z.string(),
  istKonfiguriert: z.boolean(),
})
export type TSEKonfiguration = z.infer<typeof TSEKonfigurationSchema>

export const TSEVerbindungStatusSchema = z.object({
  umgebung: z.enum(['TEST', 'LIVE']),
  tssState: z.string().min(1, 'TSS-Status fehlt'),
  clientState: z.string().min(1, 'Client-Status fehlt'),
  clientSerialNumber: z.string().min(1, 'Client-Seriennummer fehlt'),
  seriennummerKorrekt: z.boolean(),
})
export type TSEVerbindungStatus = z.infer<typeof TSEVerbindungStatusSchema>

export const TSEClientBefundSchema = z.object({
  id: z.string(),
  serialNumber: z.string(),
  state: z.string(),
})
export type TSEClientBefund = z.infer<typeof TSEClientBefundSchema>

export const TSSBefundSchema = z.object({
  id: z.string(),
  state: z.string(),
  passenderClient: TSEClientBefundSchema.nullable(),
})
export type TSSBefund = z.infer<typeof TSSBefundSchema>

export const TSESetupBefundSchema = z.object({
  umgebung: z.enum(['TEST', 'LIVE']),
  vorhandeneTss: z.array(TSSBefundSchema),
})
export type TSESetupBefund = z.infer<typeof TSESetupBefundSchema>

const apiKeyField = z
  .string()
  .trim()
  .min(1, 'API-Key ist erforderlich')
  .max(500, 'API-Key darf höchstens 500 Zeichen lang sein')
const apiSecretField = z
  .string()
  .trim()
  .min(1, 'API-Secret ist erforderlich')
  .max(500, 'API-Secret darf höchstens 500 Zeichen lang sein')
const tssIdField = z
  .string()
  .trim()
  .min(1, 'TSS-ID ist erforderlich')
  .max(255, 'TSS-ID darf höchstens 255 Zeichen lang sein')

export const TSESetupZugangsdatenSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
})
export type TSESetupZugangsdaten = z.infer<typeof TSESetupZugangsdatenSchema>

export const TSEEinrichtenSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
  umgebung: z.enum(['TEST', 'LIVE']),
})
export type TSEEinrichten = z.infer<typeof TSEEinrichtenSchema>

// Übernahme einer vorhandenen TSS: tssId wählt die TSS aus dem Befund. pin trägt
// ab Zustand UNINITIALIZED die vom Admin verwahrte Admin-PIN; bei CREATED bleibt
// es leer (jotti bezieht PUK und PIN selbst).
export const TSEUebernehmenSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
  umgebung: z.enum(['TEST', 'LIVE']),
  tssId: tssIdField,
  pin: z
    .string()
    .trim()
    .max(50, 'Admin-PIN darf höchstens 50 Zeichen lang sein'),
})
export type TSEUebernehmen = z.infer<typeof TSEUebernehmenSchema>

// PUK und Admin-PIN kommen genau einmal mit der Antwort und werden nie
// gespeichert. Sie werden dem Admin einmalig zur externen Verwahrung gezeigt.
export const TSEEinrichtenErgebnisSchema = z.object({
  tssId: z.string(),
  clientId: z.string(),
  puk: z.string(),
  adminPin: z.string(),
  umgebung: z.enum(['TEST', 'LIVE']),
})
export type TSEEinrichtenErgebnis = z.infer<typeof TSEEinrichtenErgebnisSchema>

export const TSEStatusSchema = z.object({
  umgebung: z.string(),
  offeneNachsignierungen: z
    .number()
    .int()
    .min(0, 'Offene Nachsignierungen müssen >= 0 sein'),
  istKonfiguriert: z.boolean(),
})
export type TSEStatus = z.infer<typeof TSEStatusSchema>

// Nachsignier-Auftrag: bei TSE-Ausfall vorgemerkter Vorgang. Die Liste dient
// zugleich als TSE-Ausfalldokumentation (AEAO zu § 146a, 1.14.1):
// erstelltAm = Beginn, erledigtAm = Ende, letzterFehler = Grund.
export const TSENachsignierAuftragSchema = z.object({
  id: z.number().int(),
  txId: z.string(),
  processType: z.string(),
  status: z.enum(['offen', 'erledigt', 'fehlgeschlagen', 'verworfen']),
  versuche: z.number().int(),
  letzterFehler: z.string(),
  erstelltAm: DateStringSchema,
  erledigtAm: DateStringSchema.nullable(),
})
export type TSENachsignierAuftrag = z.infer<typeof TSENachsignierAuftragSchema>

export const TSEKonfigurationSpeichernSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
  tssId: tssIdField,
  clientId: z
    .string()
    .trim()
    .min(1, 'Client-ID ist erforderlich')
    .max(255, 'Client-ID darf höchstens 255 Zeichen lang sein'),
})
export type TSEKonfigurationSpeichern = z.infer<
  typeof TSEKonfigurationSpeichernSchema
>

export class EinstellungenBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getKassenidentitaet(): Promise<Kassenidentitaet> {
    return this.backend.post(
      'admin/get-kassenidentitaet',
      {},
      KassenidentitaetSchema,
    )
  }

  public async getBetreiber(): Promise<Betreiber> {
    return this.backend.post('admin/get-betreiber', {}, BetreiberSchema)
  }

  public async saveBetreiber(betreiber: Betreiber): Promise<void> {
    const body = BetreiberSchema.parse(betreiber)
    await this.backend.post('admin/update-betreiber', body)
  }

  public async getTSEKonfiguration(): Promise<TSEKonfiguration> {
    return this.backend.post(
      'admin/get-tse-konfiguration',
      {},
      TSEKonfigurationSchema,
    )
  }

  public async saveTSEKonfiguration(
    config: TSEKonfigurationSpeichern,
  ): Promise<void> {
    const body = TSEKonfigurationSpeichernSchema.parse(config)
    await this.backend.post('admin/update-tse-konfiguration', body)
  }

  public async clearTSEKonfiguration(): Promise<void> {
    await this.backend.post('admin/update-tse-konfiguration', {
      apiKey: '',
      apiSecret: '',
      tssId: '',
      clientId: '',
    })
  }

  public async testTSEVerbindung(): Promise<TSEVerbindungStatus> {
    return this.backend.post(
      'admin/test-tse-verbindung',
      {},
      TSEVerbindungStatusSchema,
    )
  }

  public async pruefeTSESetup(
    zugangsdaten: TSESetupZugangsdaten,
  ): Promise<TSESetupBefund> {
    const body = TSESetupZugangsdatenSchema.parse(zugangsdaten)
    return this.backend.post(
      'admin/tse-setup-pruefen',
      body,
      TSESetupBefundSchema,
    )
  }

  public async richteTSEEin(
    eingabe: TSEEinrichten,
  ): Promise<TSEEinrichtenErgebnis> {
    const body = TSEEinrichtenSchema.parse(eingabe)
    return this.backend.post(
      'admin/tse-einrichten',
      body,
      TSEEinrichtenErgebnisSchema,
    )
  }

  public async uebernimmTSE(
    eingabe: TSEUebernehmen,
  ): Promise<TSEEinrichtenErgebnis> {
    const body = TSEUebernehmenSchema.parse(eingabe)
    return this.backend.post(
      'admin/tse-uebernehmen',
      body,
      TSEEinrichtenErgebnisSchema,
    )
  }

  public async getTSEStatus(): Promise<TSEStatus> {
    return this.backend.post('admin/get-tse-status', {}, TSEStatusSchema)
  }

  public async getTSENachsignierAuftraege(): Promise<TSENachsignierAuftrag[]> {
    const { auftraege } = await this.backend.post(
      'admin/get-tse-nachsignier-auftraege',
      {},
      z.object({
        auftraege: z.array(TSENachsignierAuftragSchema),
      }),
    )
    return auftraege
  }

  public async tseNachsignierAuftragZuruecksetzen(id: number): Promise<void> {
    await this.backend.post('admin/tse-nachsignier-auftrag-zuruecksetzen', {
      id,
    })
  }

  public async tseNachsignierAuftragVerwerfen(id: number): Promise<void> {
    await this.backend.post('admin/tse-nachsignier-auftrag-verwerfen', { id })
  }
}
