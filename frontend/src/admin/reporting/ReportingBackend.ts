import type { BackendClient } from '@/lib/Backend'

import { type Reporting, ReportingSchema } from './types'

export class ReportingBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getReporting(kassensitzungNr: number): Promise<Reporting> {
    return this.backend.post(
      'admin/get-abrechnung',
      { kassensitzungNr },
      ReportingSchema,
    )
  }
}
