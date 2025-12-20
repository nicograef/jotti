import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Order } from './Order'
import { TableBackend } from './TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

/** Custom hook to fetch orders for a specific table from backend. */
export function useTableOrders(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [orders, setOrders] = useState<Order[]>([])

  useEffect(() => {
    async function fetchOrders() {
      setLoading(true)

      try {
        const orders = await tableBackend.getTableOrders(tableId)
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

export function useTableBalance(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [balanceCents, setBalanceCents] = useState<number>(0)

  useEffect(() => {
    async function fetchBalance() {
      setLoading(true)

      try {
        const balance = await tableBackend.getTableBalance(tableId)
        setBalanceCents(balance)
      } catch (error) {
        console.error('Failed to fetch table balance:', error)
      }

      setLoading(false)
    }

    void fetchBalance()
  }, [tableId])

  return { loading, balanceCents }
}

export function useTableUnpaidProducts(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<Order['products']>([])

  useEffect(() => {
    async function fetchUnpaidProducts() {
      setLoading(true)

      try {
        const products = await tableBackend.getTableUnpaidProducts(tableId)
        setProducts(products)
      } catch (error) {
        console.error('Failed to fetch unpaid products:', error)
      }

      setLoading(false)
    }

    void fetchUnpaidProducts()
  }, [tableId])

  return { loading, products }
}
