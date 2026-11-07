import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { belegAnfordern, type BelegStatus } from '../beleg'
import { TischIdSchema } from '../schemas'
import {
  type BestellungAufnehmen,
  BestellungAufnehmenSchema,
} from './Bestellung'
import {
  type StornierungErteilen,
  StornierungErteilenSchema,
} from './Stornierung'
import {
  type AktiverTischMitFavorit,
  AktiverTischMitFavoritSchema,
  type EigeneUebersicht,
  EigeneUebersichtSchema,
  type Tisch,
  type TischHistorieEintrag,
  TischHistorieEintragSchema,
  TischSchema,
  type TischSession,
  TischSessionSchema,
} from './Tisch'
import { type BestellungUmbuchen, BestellungUmbuchenSchema } from './Umbuchung'
import { type ZahlungKassieren, ZahlungKassierenSchema } from './Zahlung'

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
    bestellung: BestellungAufnehmen,
  ): Promise<void> {
    const body = BestellungAufnehmenSchema.parse(bestellung)
    await this.backend.post('service/bestellung-aufnehmen', body)
  }

  public async zahlungKassieren(zahlung: ZahlungKassieren): Promise<void> {
    const body = ZahlungKassierenSchema.parse(zahlung)
    await this.backend.post('service/zahlung-kassieren', body)
  }

  public async belegDrucken(
    tischId: number,
    zahlungId: string,
  ): Promise<BelegStatus> {
    const body = z
      .object({ tischId: TischIdSchema, zahlungId: z.uuid() })
      .parse({ tischId, zahlungId })
    return belegAnfordern(this.backend, body)
  }

  public async stornobelegDrucken(
    tischId: number,
    stornierungId: string,
  ): Promise<BelegStatus> {
    const body = z
      .object({ tischId: TischIdSchema, stornierungId: z.uuid() })
      .parse({ tischId, stornierungId })
    return belegAnfordern(this.backend, body)
  }

  public async stornierungErteilen(
    stornierung: StornierungErteilen,
  ): Promise<void> {
    const body = StornierungErteilenSchema.parse(stornierung)
    await this.backend.post('serviceleitung/stornierung-erteilen', body)
  }

  public async bestellungUmbuchen(
    umbuchung: BestellungUmbuchen,
  ): Promise<void> {
    const body = BestellungUmbuchenSchema.parse(umbuchung)
    await this.backend.post('service/bestellung-umbuchen', body)
  }

  public async getTischHistorie(
    tischId: number,
  ): Promise<TischHistorieEintrag[]> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    const { historie } = await this.backend.post(
      'service/get-tisch-historie',
      body,
      z.object({ historie: z.array(TischHistorieEintragSchema) }),
    )
    return historie
  }

  public async getTischState(tischId: number): Promise<TischSession> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    return this.backend.post(
      'service/get-tisch-state',
      body,
      TischSessionSchema,
    )
  }

  public async favoritHinzufuegen(tischId: number): Promise<void> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    await this.backend.post('service/favorit-hinzufuegen', body)
  }

  public async favoritEntfernen(tischId: number): Promise<void> {
    const body = z.object({ tischId: TischIdSchema }).parse({ tischId })
    await this.backend.post('service/favorit-entfernen', body)
  }

  public async getAktiveTischeMitFavoriten(): Promise<
    AktiverTischMitFavorit[]
  > {
    const { tische } = await this.backend.post(
      'service/get-aktive-tische-mit-favoriten',
      {},
      z.object({ tische: z.array(AktiverTischMitFavoritSchema) }),
    )
    return tische
  }

  public async getMeineTischeState(): Promise<TischSession[]> {
    const { tische } = await this.backend.post(
      'service/get-meine-tische-state',
      {},
      z.object({ tische: z.array(TischSessionSchema) }),
    )
    return tische
  }

  public async getEigeneUebersicht(): Promise<EigeneUebersicht> {
    return this.backend.post(
      'service/get-eigene-uebersicht',
      {},
      EigeneUebersichtSchema,
    )
  }
}
