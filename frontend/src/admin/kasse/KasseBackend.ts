import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Kassenbestand,
  KassenbestandSchema,
  type KassensitzungState,
  KassensitzungStateSchema,
} from './types'

export class KasseBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  async kassensitzungEroeffnen(
    datum: string,
    bezeichnung: string,
  ): Promise<number> {
    const { zNr } = await this.backend.post(
      'admin/kassensitzung-eroeffnen',
      { datum, bezeichnung },
      z.object({ zNr: z.number().int() }),
    )
    return zNr
  }

  async anfangsbestandSetzen(betragCents: number): Promise<void> {
    await this.backend.post('admin/anfangsbestand-setzen', { betragCents })
  }

  async kassenbewegungBuchen(
    art: string,
    betragCents: number,
    kommentar: string,
  ): Promise<void> {
    await this.backend.post('admin/kassenbewegung-buchen', {
      art,
      betragCents,
      kommentar,
    })
  }

  async kassensturzDurchfuehren(istBestandCents: number): Promise<void> {
    await this.backend.post('admin/kassensturz-durchfuehren', {
      istBestandCents,
    })
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
