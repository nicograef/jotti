import { z } from 'zod'

import type { BackendClient } from '@/lib/Backend'

import {
  type Cancelation,
  CancelationSchema,
  CancelVariantsSchema,
} from './Cancelation'
import {
  DeliverVariantsSchema,
  type Delivery,
  DeliverySchema,
} from './Delivery'
import {
  type LineItem,
  LineItemSchema,
  type Order,
  OrderSchema,
  PlaceOrderSchema,
} from './Order'
import { type Payment, PaymentSchema, RegisterPaymentSchema } from './Payment'
import { type Table, TableIdSchema, TableSchema } from './Table'

export class TableBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
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

  public async placeTableOrder(
    placeOrder: z.infer<typeof PlaceOrderSchema>,
  ): Promise<void> {
    const body = PlaceOrderSchema.parse(placeOrder)
    await this.backend.post('service/place-table-order', body)
  }

  public async registerTablePayment(
    registerPayment: z.infer<typeof RegisterPaymentSchema>,
  ): Promise<void> {
    const body = RegisterPaymentSchema.parse(registerPayment)
    await this.backend.post('service/register-table-payment', body)
  }

  public async cancelTableVariants(
    cancelVariants: z.infer<typeof CancelVariantsSchema>,
  ): Promise<void> {
    const body = CancelVariantsSchema.parse(cancelVariants)
    await this.backend.post('service/cancel-table-variants', body)
  }

  public async deliverTableVariants(
    deliverVariants: z.infer<typeof DeliverVariantsSchema>,
  ): Promise<void> {
    const body = DeliverVariantsSchema.parse(deliverVariants)
    await this.backend.post('service/deliver-table-variants', body)
  }

  public async getTableHistory(
    tableId: number,
  ): Promise<(Order | Payment | Cancelation | Delivery)[]> {
    const body = z.object({ tableId: TableIdSchema }).parse({ tableId })
    const { history } = await this.backend.post(
      'service/get-table-history',
      body,
      z.object({
        history: z.array(
          z.union([
            OrderSchema,
            PaymentSchema,
            CancelationSchema,
            DeliverySchema,
          ]),
        ),
      }),
    )
    return history
  }

  public async getTableBalance(tableId: number): Promise<number> {
    const body = z.object({ tableId: TableIdSchema }).parse({ tableId })
    const { balanceCents } = await this.backend.post(
      'service/get-table-balance',
      body,
      z.object({ balanceCents: z.number().int() }),
    )
    return balanceCents
  }

  public async getTableUnpaidVariants(tableId: number): Promise<LineItem[]> {
    const body = z.object({ tableId: TableIdSchema }).parse({ tableId })
    const { variants } = await this.backend.post(
      'service/get-table-unpaid-variants',
      body,
      z.object({ variants: z.array(LineItemSchema) }),
    )
    return variants
  }

  public async getTableUndeliveredVariants(
    tableId: number,
  ): Promise<LineItem[]> {
    const body = z.object({ tableId: TableIdSchema }).parse({ tableId })
    const { variants } = await this.backend.post(
      'service/get-table-undelivered-variants',
      body,
      z.object({ variants: z.array(LineItemSchema) }),
    )
    return variants
  }
}
