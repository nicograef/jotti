import { z } from 'zod'

import { type Table, TableSchema } from './Table'

export const CreateTableRequestSchema = TableSchema.pick({
  name: true,
})

export const UpdateTableRequestSchema = TableSchema.pick({
  id: true,
  name: true,
})

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

  public async getAllTables(): Promise<Table[]> {
    const { tables } = await this.backend.post(
      'admin/get-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async getTables(): Promise<Table[]> {
    const { tables } = await this.backend.post(
      'service/get-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async createTable(
    newTable: z.infer<typeof CreateTableRequestSchema>,
  ): Promise<Table> {
    const body = CreateTableRequestSchema.parse(newTable)
    const { table } = await this.backend.post(
      'admin/create-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async updateTable(
    updatedTable: z.infer<typeof UpdateTableRequestSchema>,
  ): Promise<Table> {
    const body = UpdateTableRequestSchema.parse(updatedTable)
    const { table } = await this.backend.post(
      'admin/update-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async activateTable(id: number): Promise<void> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/activate-table', body)
  }

  public async deactivateTable(id: number): Promise<void> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    await this.backend.post('admin/deactivate-table', body)
  }
}
