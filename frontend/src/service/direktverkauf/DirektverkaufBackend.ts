import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import { BelegDruckenResponseSchema, type BelegStatus } from '../beleg'
import {
  type DirektverkaufHistorieEintrag,
  DirektverkaufHistorieEintragSchema,
  type DirektverkaufKassenbelegDrucken,
  DirektverkaufKassenbelegDruckenSchema,
  type DirektverkaufStornieren,
  DirektverkaufStornierenSchema,
  type DirektverkaufTaetigen,
  DirektverkaufTaetigenSchema,
} from './Direktverkauf'

export class DirektverkaufBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async direktverkaufTaetigen(
    verkauf: DirektverkaufTaetigen,
  ): Promise<void> {
    const body = DirektverkaufTaetigenSchema.parse(verkauf)
    await this.backend.post('service/direktverkauf-taetigen', body)
  }

  public async getDirektverkaufHistorie(): Promise<
    DirektverkaufHistorieEintrag[]
  > {
    const { historie } = await this.backend.post(
      'service/get-direktverkauf-historie',
      {},
      z.object({ historie: DirektverkaufHistorieEintragSchema.array() }),
    )
    return historie
  }

  public async direktverkaufStornieren(
    storno: DirektverkaufStornieren,
  ): Promise<void> {
    const body = DirektverkaufStornierenSchema.parse(storno)
    await this.backend.post('serviceleitung/direktverkauf-stornieren', body)
  }

  public async kassenbelegDrucken(
    cmd: DirektverkaufKassenbelegDrucken,
  ): Promise<BelegStatus> {
    const body = DirektverkaufKassenbelegDruckenSchema.parse(cmd)
    const { status } = await this.backend.post(
      'service/beleg-drucken',
      body,
      BelegDruckenResponseSchema,
    )
    return status
  }
}
