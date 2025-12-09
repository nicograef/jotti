import { z } from 'zod'

import { type Table, TableIdSchema, TableSchema } from './Table'

export const CreateTableSchema = TableSchema.pick({
  name: true,
})

export const UpdateTableSchema = TableSchema.pick({
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
      'admin/get-all-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async createTable(
    newTable: z.infer<typeof CreateTableSchema>,
  ): Promise<number> {
    const body = CreateTableSchema.parse(newTable)
    const { id } = await this.backend.post(
      'admin/create-table',
      body,
      z.object({ id: TableIdSchema }),
    )
    return id
  }

  public async updateTable(
    updatedTable: z.infer<typeof UpdateTableSchema>,
  ): Promise<void> {
    const body = UpdateTableSchema.parse(updatedTable)
    await this.backend.post('admin/update-table', body)
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
