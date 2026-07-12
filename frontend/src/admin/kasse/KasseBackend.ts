import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  BetragCentsSchema,
  BezeichnungSchema,
  type GeldtransitRichtung,
  GeldtransitRichtungSchema,
  type Kassenbestand,
  KassenbestandSchema,
  KassensitzungSchema,
} from './Kassensitzung'

export const KassensitzungEroeffnenSchema = z.object({
  bezeichnung: BezeichnungSchema,
  betragCents: BetragCentsSchema,
})

// Die offene Kassensitzung liefert zusätzlich den Eröffnungszeitpunkt
// (eroeffnetAm, RFC-3339), den die abgeschlossenen Sitzungen im Reporting nicht
// mitgeben — daher eine eigene Erweiterung der kanonischen KassensitzungSchema.
export const OffeneKassensitzungSchema = KassensitzungSchema.extend({
  eroeffnetAm: z.string(),
})
export type OffeneKassensitzung = z.infer<typeof OffeneKassensitzungSchema>

export const GeldtransitBuchenSchema = z.object({
  geldtransitId: z.uuid(),
  richtung: GeldtransitRichtungSchema,
  betragCents: z
    .number()
    .int()
    .min(1, { message: 'Betrag muss mindestens 1 Cent sein.' }),
  kommentar: z
    .string()
    .min(3, { message: 'Kommentar muss mindestens 3 Zeichen lang sein.' })
    .max(200, { message: 'Kommentar darf maximal 200 Zeichen lang sein.' }),
})

export const KasseAbschliessenSchema = z.object({
  istBestandCents: BetragCentsSchema,
})

// KassenabschlussErgebnis weist die beim Abschluss verbliebenen Ausfall-Reste
// aus: Vorgänge, die die TSE noch nachsigniert, und Vorgänge ohne Signatur
// mangels TSE-Konfiguration.
export const KassenabschlussErgebnisSchema = z.object({
  ausfallResteAnzahl: z.number().int(),
  ohneKonfigurationAnzahl: z.number().int(),
})
export type KassenabschlussErgebnis = z.infer<
  typeof KassenabschlussErgebnisSchema
>

// SignaturenAusstehendDetails sind die 409-Details des Kassenabschluss-Gates:
// wie viele Signaturen noch ausstehen und wie alt der älteste offene Auftrag ist.
export const SignaturenAusstehendDetailsSchema = z.object({
  anzahl: z.number().int(),
  alterSekunden: z.number().int(),
})

export class KasseBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  async kassensitzungEroeffnen(
    bezeichnung: string,
    betragCents: number,
  ): Promise<number> {
    const body = KassensitzungEroeffnenSchema.parse({
      bezeichnung,
      betragCents,
    })
    const { zNr } = await this.backend.post(
      'admin/kassensitzung-eroeffnen',
      body,
      z.object({ zNr: z.number().int() }),
    )
    return zNr
  }

  async geldtransitBuchen(
    geldtransitId: string,
    richtung: GeldtransitRichtung,
    betragCents: number,
    kommentar: string,
  ): Promise<void> {
    const body = GeldtransitBuchenSchema.parse({
      geldtransitId,
      richtung,
      betragCents,
      kommentar,
    })
    await this.backend.post('admin/geldtransit-buchen', body)
  }

  async kasseAbschliessen(
    istBestandCents: number,
  ): Promise<KassenabschlussErgebnis> {
    const body = KasseAbschliessenSchema.parse({ istBestandCents })
    return this.backend.post(
      'admin/kasse-abschliessen',
      body,
      KassenabschlussErgebnisSchema,
    )
  }

  async getOffeneKassensitzung(): Promise<OffeneKassensitzung | null> {
    const data = await this.backend.post(
      'admin/get-offene-kassensitzung',
      {},
      OffeneKassensitzungSchema.nullable(),
    )
    return data
  }

  async getKassenbestand(kassensitzungNr: number): Promise<Kassenbestand> {
    return this.backend.post(
      'admin/get-kassenbestand',
      { kassensitzungNr },
      KassenbestandSchema,
    )
  }
}
