import type { BackendClient } from '@/lib/Backend'

import {
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
}
