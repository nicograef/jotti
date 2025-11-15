import { z } from 'zod'

const TableIdSchema = z.number().int().min(1)
const TableNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(30, { message: 'Der Name ist zu lang.' })
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const TableSchema = z.object({
  id: TableIdSchema.int().positive(),
  name: TableNameSchema,
  locked: z.boolean(),
  createdAt: DateStringSchema,
})
export type Table = z.infer<typeof TableSchema>

export const CreateTableRequestSchema = z.object({
  name: TableNameSchema,
})
export type CreateTableRequest = z.infer<typeof CreateTableRequestSchema>

const UpdateTableRequestSchema = z.object({
  id: TableIdSchema,
  name: TableNameSchema,
})
export type UpdateTableRequest = z.infer<typeof UpdateTableRequestSchema>

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class TableBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  public async getTables(): Promise<Table[]> {
    const { tables } = await this.backend.post(
      'admin/get-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }

  public async createTable(newTable: CreateTableRequest): Promise<Table> {
    const body = CreateTableRequestSchema.parse(newTable)
    const { table } = await this.backend.post(
      'admin/create-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }

  public async updateTable(updatedTable: UpdateTableRequest): Promise<Table> {
    const body = UpdateTableRequestSchema.parse(updatedTable)
    const { table } = await this.backend.post(
      'admin/update-table',
      body,
      z.object({ table: TableSchema }),
    )
    return table
  }
}
