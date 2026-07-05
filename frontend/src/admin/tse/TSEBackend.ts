import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'
import { DateStringSchema } from '@/lib/utils'

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

// Eine TSE ist signierfähig, wenn die TSS initialisiert, der Client registriert
// und seine Seriennummer die Kassen-Seriennummer ist. Eine einzige Definition
// für Wizard-Abschluss und manuellen Verbindungstest.
export function verbindungIstSigniertfaehig(
  status: TSEVerbindungStatus,
): boolean {
  return (
    status.tssState.toUpperCase() === 'INITIALIZED' &&
    status.clientState.toUpperCase() === 'REGISTERED' &&
    status.seriennummerKorrekt
  )
}

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

// neuAnlegenTrotzVorhandener erzwingt in TEST bewusst eine zweite, frische TSE
// trotz vorhandener TSS (F2). Optional und nur in TEST wirksam; LIVE bleibt im
// Backend hart gesperrt.
export const TSEEinrichtenSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
  umgebung: z.enum(['TEST', 'LIVE']),
  neuAnlegenTrotzVorhandener: z.boolean().optional(),
})
export type TSEEinrichten = z.infer<typeof TSEEinrichtenSchema>

// Übernahme einer vorhandenen TSS: tssId wählt die TSS aus dem Befund. pin trägt
// ab Zustand UNINITIALIZED die vom Admin verwahrte Admin-PIN; bei CREATED bleibt
// es leer (jotti bezieht PUK und PIN selbst). puk ist nur für den PIN-Reset
// gesetzt: ist die PIN verloren oder gesperrt, setzt jotti mit dem PUK eine
// frische PIN und übernimmt damit weiter.
export const TSEUebernehmenSchema = z.object({
  apiKey: apiKeyField,
  apiSecret: apiSecretField,
  umgebung: z.enum(['TEST', 'LIVE']),
  tssId: tssIdField,
  pin: z
    .string()
    .trim()
    .max(50, 'Admin-PIN darf höchstens 50 Zeichen lang sein'),
  puk: z
    .string()
    .trim()
    .max(100, 'Admin-PUK darf höchstens 100 Zeichen lang sein')
    .optional(),
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
  istKonfiguriert: z.boolean(),
})
export type TSEStatus = z.infer<typeof TSEStatusSchema>

// Zustand der Signatur-Queue für das Admin-Monitoring: Rückstand (offene
// Aufträge, Alter des ältesten) und Leistung über ein gleitendes
// 15-Minuten-Fenster (Signaturen/Minute, Signierdauer p95). Fehlgeschlagene
// Aufträge und letzterFehler sind sitzungsbezogen (nur die aktive
// Kassensitzung); mit dem Kassenabschluss verschwindet die Warnung.
export const TSESignaturQueueSchema = z.object({
  offeneAuftraege: z.number().int(),
  fehlgeschlageneAuftraege: z.number().int(),
  letzterFehler: z.string(),
  rueckstandSekunden: z.number().int(),
  signaturenProMinute: z.number(),
  signierdauerP95Sekunden: z.number(),
})
export type TSESignaturQueue = z.infer<typeof TSESignaturQueueSchema>

// Störungszeitraum aus dem Störungsprotokoll (TSE-Ausfalldokumentation):
// ein Zeitraum mit Beginn, Ende (null solange aktiv) und Grund-Art.
export const TSEStoerungSchema = z.object({
  id: z.number().int(),
  beginn: DateStringSchema,
  ende: DateStringSchema.nullable(),
  grundArt: z.enum(['tse_fehler', 'rueckstand', 'keine_konfiguration']),
  fehlertext: z.string(),
})
export type TSEStoerung = z.infer<typeof TSEStoerungSchema>

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

export class TSEBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
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

  public async checkTSESetup(
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

  public async getTSESignaturQueue(): Promise<TSESignaturQueue> {
    return this.backend.post(
      'admin/get-tse-signatur-queue',
      {},
      TSESignaturQueueSchema,
    )
  }

  public async getTSEStoerungen(): Promise<TSEStoerung[]> {
    const { stoerungen } = await this.backend.post(
      'admin/get-tse-stoerungen',
      {},
      z.object({
        stoerungen: z.array(TSEStoerungSchema),
      }),
    )
    return stoerungen
  }
}
