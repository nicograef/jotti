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
  ): Promise<Order> {
    const body = PlaceOrderSchema.parse(placeOrder)
    const { order } = await this.backend.post(
      'place-order',
      body,
      z.object({ order: OrderSchema }),
    )
    return order
  }
}
