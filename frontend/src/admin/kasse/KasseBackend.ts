import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  BetragCentsSchema,
  BezeichnungSchema,
  type GeldtransitRichtung,
  GeldtransitRichtungSchema,
  type Kassenbestand,
  KassenbestandSchema,
  type KassensitzungState,
  KassensitzungStateSchema,
} from './Kassensitzung'

export const KassensitzungEroeffnenSchema = z.object({
  bezeichnung: BezeichnungSchema,
  betragCents: BetragCentsSchema,
})

export const GeldtransitBuchenSchema = z.object({
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

export const KassensturzDurchfuehrenSchema = z.object({
  istBestandCents: BetragCentsSchema,
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
    richtung: GeldtransitRichtung,
    betragCents: number,
    kommentar: string,
  ): Promise<void> {
    const body = GeldtransitBuchenSchema.parse({
      richtung,
      betragCents,
      kommentar,
    })
    await this.backend.post('admin/geldtransit-buchen', body)
  }

  async kassensturzDurchfuehren(istBestandCents: number): Promise<void> {
    const body = KassensturzDurchfuehrenSchema.parse({ istBestandCents })
    await this.backend.post('admin/kassensturz-durchfuehren', body)
  }

  async tagesabschlussErstellen(): Promise<void> {
    await this.backend.post('admin/tagesabschluss-erstellen', {})
  }

  async getOffeneKassensitzung(): Promise<KassensitzungState | null> {
    const data = await this.backend.post(
      'admin/get-offene-kassensitzung',
      {},
      KassensitzungStateSchema.nullable(),
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
