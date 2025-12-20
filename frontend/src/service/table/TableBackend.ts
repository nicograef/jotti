import { z } from 'zod'

import { type Table, TableSchema } from './Table'

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class TableBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  public async getActiveTables(): Promise<Table[]> {
    const { tables } = await this.backend.post(
      'service/get-active-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async getTable(id: number): Promise<Table> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    const { table } = await this.backend.post(
      'service/get-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }
}
