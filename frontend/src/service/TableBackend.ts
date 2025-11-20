import { z } from 'zod'

const TableIdSchema = z.number().int().min(1)
const TableNameSchema = z.string().min(3)
const TableSchema = z.object({
  id: TableIdSchema,
  name: TableNameSchema,
})
export type Table = z.infer<typeof TableSchema>

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

  public async getTables(): Promise<Table[]> {
    const { tables } = await this.backend.post(
      'service/get-tables',
      {},
      z.object({ tables: z.array(TableSchema) }),
    )
    return tables
  }
}
