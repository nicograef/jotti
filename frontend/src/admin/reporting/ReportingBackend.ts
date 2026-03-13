import type { BackendClient } from '@/lib/Backend'

import {
  type Dashboard,
  DashboardSchema,
  type Tagesabrechnung,
  TagesabrechnungSchema,
} from './types'

export class ReportingBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getDashboard(): Promise<Dashboard> {
    return this.backend.post('admin/get-dashboard', {}, DashboardSchema)
  }

  public async getTagesabrechnung(
    von: string,
    bis: string,
  ): Promise<Tagesabrechnung> {
    return this.backend.post(
      'admin/get-tagesabrechnung',
      { von, bis },
      TagesabrechnungSchema,
    )
  }
}
