import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Ausgabe,
  AusgabeBestaetigenSchema,
  AusgabeSchema,
} from './Ausgabe'
import {
  type Auszahlung,
  AuszahlungLeistenSchema,
  AuszahlungSchema,
} from './Auszahlung'
import {
  type Bestellung,
  BestellungAufnehmenSchema,
  BestellungSchema,
} from './Bestellung'
import {
  type Stornierung,
  StornierungErteilenSchema,
  StornierungSchema,
} from './Stornierung'
import {
  type Tisch,
  TischIdSchema,
  TischSchema,
  type TischState,
  TischStateSchema,
} from './Tisch'
import { type Zahlung, ZahlungKassierenSchema, ZahlungSchema } from './Zahlung'

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

  public async bestellungAufnehmen(
    bestellung: z.infer<typeof BestellungAufnehmenSchema>,
  ): Promise<void> {
    const body = BestellungAufnehmenSchema.parse(bestellung)
    await this.backend.post('service/bestellung-aufnehmen', body)
  }

  public async zahlungKassieren(
    zahlung: z.infer<typeof ZahlungKassierenSchema>,
  ): Promise<void> {
    const body = ZahlungKassierenSchema.parse(zahlung)
    await this.backend.post('service/zahlung-kassieren', body)
  }

  public async stornierungErteilen(
    stornierung: z.infer<typeof StornierungErteilenSchema>,
  ): Promise<void> {
    const body = StornierungErteilenSchema.parse(stornierung)
    await this.backend.post('serviceleitung/stornierung-erteilen', body)
  }

  public async auszahlungLeisten(
    cmd: z.infer<typeof AuszahlungLeistenSchema>,
  ): Promise<void> {
    const body = AuszahlungLeistenSchema.parse(cmd)
    await this.backend.post('serviceleitung/auszahlung-leisten', body)
  }

  public async ausgabeBestaetigen(
    ausgabe: z.infer<typeof AusgabeBestaetigenSchema>,
  ): Promise<void> {
    const body = AusgabeBestaetigenSchema.parse(ausgabe)
    await this.backend.post('service/ausgabe-bestaetigen', body)
  }

  public async getTischHistorie(
    tischId: number,
  ): Promise<(Bestellung | Zahlung | Stornierung | Ausgabe | Auszahlung)[]> {
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
            AusgabeSchema,
            AuszahlungSchema,
          ]),
        ),
      }),
    )
    return historie
  }

  public async getTischState(tischId: number): Promise<TischState> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    return this.backend.post('service/get-tisch-state', body, TischStateSchema)
  }
}
