import type { BackendClient } from '@/lib/Backend'

import { type Reporting, ReportingSchema } from './types'

export class ReportingBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getReporting(von: string, bis: string): Promise<Reporting> {
    return this.backend.post(
      'admin/get-reporting',
      { von, bis },
      ReportingSchema,
    )
  }
}
