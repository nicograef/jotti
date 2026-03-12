import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Bestellung,
  BestellungAufgebenSchema,
  BestellungSchema,
  type Position,
  PositionSchema,
} from './Bestellung'
import {
  type Lieferung,
  LieferungSchema,
  ProdukteLiefernSchema,
} from './Lieferung'
import {
  ProdukteStornierenSchema,
  type Stornierung,
  StornierungSchema,
} from './Stornierung'
import { type Tisch, TischIdSchema, TischSchema } from './Tisch'
import {
  type Zahlung,
  ZahlungRegistrierenSchema,
  ZahlungSchema,
} from './Zahlung'

export class TischBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getAktiveTische(): Promise<Tisch[]> {
    const { tische } = await this.backend.post(
      'service/get-aktive-tische',
      {},
      z.object({ tische: z.array(TischSchema) }),
    )
    return tische
  }

  public async getTisch(id: number): Promise<Tisch> {
    const body = TischSchema.pick({ id: true }).parse({ id })
    const { tisch } = await this.backend.post(
      'service/get-tisch',
      body,
      z.object({ tisch: TischSchema }),
    )
    return tisch
  }

  public async bestellungAufgeben(
    bestellung: z.infer<typeof BestellungAufgebenSchema>,
  ): Promise<void> {
    const body = BestellungAufgebenSchema.parse(bestellung)
    await this.backend.post('service/bestellung-aufgeben', body)
  }

  public async zahlungRegistrieren(
    zahlung: z.infer<typeof ZahlungRegistrierenSchema>,
  ): Promise<void> {
    const body = ZahlungRegistrierenSchema.parse(zahlung)
    await this.backend.post('service/zahlung-registrieren', body)
  }

  public async produkteStornieren(
    stornierung: z.infer<typeof ProdukteStornierenSchema>,
  ): Promise<void> {
    const body = ProdukteStornierenSchema.parse(stornierung)
    await this.backend.post('serviceleitung/produkte-stornieren', body)
  }

  public async produkteLiefern(
    lieferung: z.infer<typeof ProdukteLiefernSchema>,
  ): Promise<void> {
    const body = ProdukteLiefernSchema.parse(lieferung)
    await this.backend.post('service/produkte-liefern', body)
  }

  public async getTischHistorie(
    tischId: number,
  ): Promise<(Bestellung | Zahlung | Stornierung | Lieferung)[]> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    const { historie } = await this.backend.post(
      'service/get-tisch-historie',
      body,
      z.object({
        historie: z.array(
          z.union([
            BestellungSchema,
            ZahlungSchema,
            StornierungSchema,
            LieferungSchema,
          ]),
        ),
      }),
    )
    return historie
  }

  public async getTischSaldo(tischId: number): Promise<number> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    const { saldoCents } = await this.backend.post(
      'service/get-tisch-saldo',
      body,
      z.object({ saldoCents: z.number().int() }),
    )
    return saldoCents
  }

  public async getTischUnbezahlt(tischId: number): Promise<Position[]> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    const { positionen } = await this.backend.post(
      'service/get-tisch-unbezahlt',
      body,
      z.object({ positionen: z.array(PositionSchema) }),
    )
    return positionen
  }

  public async getTischUngeliefert(tischId: number): Promise<Position[]> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    const { positionen } = await this.backend.post(
      'service/get-tisch-ungeliefert',
      body,
      z.object({ positionen: z.array(PositionSchema) }),
    )
    return positionen
  }
}
