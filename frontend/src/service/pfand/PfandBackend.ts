import { z } from 'zod'

import { type BackendClient, BackendSingleton } from '@/lib/Backend'

import { BelegDruckenResponseSchema, type BelegStatus } from '../beleg'
import {
  type PfandbelegDrucken,
  PfandbelegDruckenSchema,
  type PfandrueckgabeHistorieEintrag,
  PfandrueckgabeHistorieEintragSchema,
  type PfandrueckgabeStornieren,
  PfandrueckgabeStornierenSchema,
  type PfandZurueckgeben,
  PfandZurueckgebenSchema,
} from './Pfand'

export class PfandBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async pfandZurueckgeben(rueckgabe: PfandZurueckgeben): Promise<void> {
    const body = PfandZurueckgebenSchema.parse(rueckgabe)
    await this.backend.post('service/pfand-zurueckgeben', body)
  }

  // tischId 0 liefert die Rückgaben des Direktverkaufs.
  public async getPfandrueckgabeHistorie(
    tischId: number,
  ): Promise<PfandrueckgabeHistorieEintrag[]> {
    const { historie } = await this.backend.post(
      'service/get-pfandrueckgabe-historie',
      { tischId },
      z.object({ historie: PfandrueckgabeHistorieEintragSchema.array() }),
    )
    return historie
  }

  public async pfandrueckgabeStornieren(
    storno: PfandrueckgabeStornieren,
  ): Promise<void> {
    const body = PfandrueckgabeStornierenSchema.parse(storno)
    await this.backend.post('serviceleitung/pfandrueckgabe-stornieren', body)
  }

  public async pfandbelegDrucken(cmd: PfandbelegDrucken): Promise<BelegStatus> {
    const body = PfandbelegDruckenSchema.parse(cmd)
    const { status } = await this.backend.post(
      'service/beleg-drucken',
      body,
      BelegDruckenResponseSchema,
    )
    return status
  }
}

export const pfandBackend = new PfandBackend(BackendSingleton)
