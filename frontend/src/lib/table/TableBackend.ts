import { z } from 'zod'

import {
  type Table,
  type TablePublic,
  TablePublicSchema,
  TableSchema,
} from './Table'

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
      'get-all-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async getActiveTables(): Promise<TablePublic[]> {
    const { tables } = await this.backend.post(
      'get-active-tables',
      {},
      z.object({ tables: z.array(TablePublicSchema) }),
    )
    return tables
  }

  public async getTable(id: number): Promise<Table> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    const { table } = await this.backend.post(
      'get-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async createTable(
    newTable: z.infer<typeof CreateTableSchema>,
  ): Promise<Table> {
    const body = CreateTableSchema.parse(newTable)
    const { table } = await this.backend.post(
      'create-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async updateTable(
    updatedTable: z.infer<typeof UpdateTableSchema>,
  ): Promise<Table> {
    const body = UpdateTableSchema.parse(updatedTable)
    const { table } = await this.backend.post(
      'update-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async activateTable(id: number): Promise<void> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    await this.backend.post('activate-table', body)
  }

  public async deactivateTable(id: number): Promise<void> {
    const body = TableSchema.pick({ id: true }).parse({ id })
    await this.backend.post('deactivate-table', body)
  }
}
