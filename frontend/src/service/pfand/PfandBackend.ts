import { type BackendClient, BackendSingleton } from '@/lib/Backend'

import { type PfandZurueckgeben, PfandZurueckgebenSchema } from './Pfand'

export class PfandBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async pfandZurueckgeben(rueckgabe: PfandZurueckgeben): Promise<void> {
    const body = PfandZurueckgebenSchema.parse(rueckgabe)
    await this.backend.post('service/pfand-zurueckgeben', body)
  }
}

export const pfandBackend = new PfandBackend(BackendSingleton)
