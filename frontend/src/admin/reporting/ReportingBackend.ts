import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Kassensitzung,
  KassensitzungSchema,
  type ReportingData,
  ReportingDataSchema,
} from './types'

export class ReportingBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getReporting(kassensitzungNr: number): Promise<ReportingData> {
    return this.backend.post(
      'admin/get-abrechnung',
      { kassensitzungNr },
      ReportingDataSchema,
    )
  }

  public async getAllKassensitzungen(): Promise<Kassensitzung[]> {
    const response = await this.backend.post(
      'admin/get-all-kassensitzungen',
      {},
      z.object({ kassensitzungen: z.array(KassensitzungSchema) }),
    )
    return response.kassensitzungen
  }
}
