import { z } from 'zod'

import { type Table, TableIdSchema, TableSchema } from './Table'

export const CreateTableSchema = TableSchema.pick({
  name: true,
})

export const UpdateTableSchema = TableSchema.pick({
  id: true,
  name: true,
})

import type { BackendClient } from '@/lib/Backend'

export class TableBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
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
