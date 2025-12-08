import { z } from 'zod'

import { type Order, OrderSchema, PlaceOrderSchema } from './Order'

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
  ): Promise<string> {
    const body = PlaceOrderSchema.parse(placeOrder)
    const { id } = await this.backend.post(
      'place-order',
      body,
      z.object({ id: OrderSchema.shape.id }),
    )
    return id
  }

  public async getOrders(tableId: number): Promise<Order[]> {
    const body = OrderSchema.pick({ tableId: true }).parse({ tableId })
    const { orders } = await this.backend.post(
      'get-orders',
      body,
      z.object({ orders: z.array(OrderSchema) }),
    )
    return orders
  }
}
