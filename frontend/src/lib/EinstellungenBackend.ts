import { z } from 'zod'

import type { BackendClient } from './Backend'

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
  angelegtAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  }),
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
})
export type TSEVerbindungStatus = z.infer<typeof TSEVerbindungStatusSchema>

export const TSEStatusSchema = z.object({
  umgebung: z.string(),
  offeneNachsignierungen: z
    .number()
    .int()
    .min(0, 'Offene Nachsignierungen müssen >= 0 sein'),
  istKonfiguriert: z.boolean(),
})
export type TSEStatus = z.infer<typeof TSEStatusSchema>

export const TSEKonfigurationSpeichernSchema = z.object({
  apiKey: z
    .string()
    .trim()
    .min(1, 'API-Key ist erforderlich')
    .max(500, 'API-Key darf höchstens 500 Zeichen lang sein'),
  apiSecret: z
    .string()
    .trim()
    .min(1, 'API-Secret ist erforderlich')
    .max(500, 'API-Secret darf höchstens 500 Zeichen lang sein'),
  tssId: z
    .string()
    .trim()
    .min(1, 'TSS-ID ist erforderlich')
    .max(255, 'TSS-ID darf höchstens 255 Zeichen lang sein'),
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

  public async saveBetreiber(b: Betreiber): Promise<void> {
    const body = BetreiberSchema.parse(b)
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

  public async getTSEStatus(): Promise<TSEStatus> {
    return this.backend.post('admin/get-tse-status', {}, TSEStatusSchema)
  }
}
