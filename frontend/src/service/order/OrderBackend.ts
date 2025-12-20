import { z } from 'zod'

import {
  type Order,
  type OrderProduct,
  OrderProductSchema,
  OrderSchema,
  PlaceOrderSchema,
  RegisterPaymentSchema,
} from './Order'

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class OrderBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  public async placeOrder(
    placeOrder: z.infer<typeof PlaceOrderSchema>,
  ): Promise<void> {
    const body = PlaceOrderSchema.parse(placeOrder)
    await this.backend.post('service/place-order', body)
  }

  // eslint-disable-next-line @typescript-eslint/require-await
  public async registerPayment(
    registerPayment: z.infer<typeof RegisterPaymentSchema>,
  ): Promise<void> {
    console.log('registerPayment', registerPayment)
  }

  public async getTableOrders(tableId: number): Promise<Order[]> {
    const body = OrderSchema.pick({ tableId: true }).parse({ tableId })
    const { orders } = await this.backend.post(
      'service/get-table-orders',
      body,
      z.object({ orders: z.array(OrderSchema) }),
    )
    return orders
  }

  public async getTableBalance(tableId: number): Promise<number> {
    const body = OrderSchema.pick({ tableId: true }).parse({ tableId })
    const { balanceCents } = await this.backend.post(
      'service/get-table-balance',
      body,
      z.object({ balanceCents: z.number().int() }),
    )
    return balanceCents
  }

  public async getTableUnpaidProducts(
    tableId: number,
  ): Promise<OrderProduct[]> {
    const body = OrderSchema.pick({ tableId: true }).parse({ tableId })
    const { products } = await this.backend.post(
      'service/get-table-unpaid-products',
      body,
      z.object({ products: z.array(OrderProductSchema) }),
    )
    return products
  }
}
