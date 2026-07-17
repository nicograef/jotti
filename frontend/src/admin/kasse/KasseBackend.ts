import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  BetragCentsSchema,
  BezeichnungSchema,
  type GeldtransitBuchung,
  GeldtransitBuchungSchema,
  type GeldtransitRichtung,
  GeldtransitRichtungSchema,
  type Kassenbestand,
  KassenbestandSchema,
  KassensitzungSchema,
  KommentarSchema,
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
  kommentar: KommentarSchema,
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
// wie viele Signaturen noch ausstehen.
export const SignaturenAusstehendDetailsSchema = z.object({
  anzahl: z.number().int(),
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

  async getGeldtransitListe(
    kassensitzungNr: number,
  ): Promise<GeldtransitBuchung[]> {
    return this.backend.post(
      'admin/get-geldtransit-liste',
      { kassensitzungNr },
      z.array(GeldtransitBuchungSchema),
    )
  }
}
