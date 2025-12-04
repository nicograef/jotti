import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Order } from './Order'
import { OrderBackend } from './OrderBackend'

const orderBackend = new OrderBackend(BackendSingleton)

/** Custom hook to fetch orders for a specific table from backend. */
export function useOrders(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    async function fetchOrders() {
      setLoading(true)

      try {
        const orders = await orderBackend.getOrders(tableId)
        setOrders(orders)
      } catch (error) {
        console.error('Failed to fetch orders:', error)
      }

      setLoading(false)
    }

    void fetchOrders()
  }, [tableId])

  return { loading, orders }
}
