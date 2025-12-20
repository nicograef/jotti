import { useCallback, useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Order } from './Order'
import type { Payment } from './Payment'
import type { Table } from './Table'
import { TableBackend } from './TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

/** Custom hook to fetch a single table from backend. */
export function useTable(id: number) {
  const [loading, setLoading] = useState(false)
  const [table, setTable] = useState<Table | null>(null)

  useEffect(() => {
    async function fetchTable() {
      setLoading(true)
      try {
        const table = await tableBackend.getTable(id)
        setTable(table)
      } catch (error) {
        console.error('Failed to fetch tables:', error)
      }
      setLoading(false)
    }

    void fetchTable()
  }, [id])

  return { loading, table }
}

/** Custom hook to fetch active tables from backend. */
export function useActiveTables() {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<Table[]>([])

  useEffect(() => {
    async function fetchTables() {
      setLoading(true)
      try {
        const tables = await tableBackend.getActiveTables()
        setTables(tables)
      } catch (error) {
        console.error('Failed to fetch tables:', error)
      }
      setLoading(false)
    }

    void fetchTables()
  }, [])

  return { loading, tables }
}

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

/** Custom hook to fetch payments for a specific table from backend. */
export function useTablePayments(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [payments, setPayments] = useState<Payment[]>([])

  useEffect(() => {
    async function fetchPayments() {
      setLoading(true)

      try {
        const payments = await tableBackend.getTablePayments(tableId)
        setPayments(payments)
      } catch (error) {
        console.error('Failed to fetch payments:', error)
      }

      setLoading(false)
    }

    void fetchPayments()
  }, [tableId])

  return { loading, payments }
}

export function useTableBalance(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [balanceCents, setBalanceCents] = useState<number>(0)

  const fetchBalance = useCallback(async () => {
    setLoading(true)

    try {
      const balance = await tableBackend.getTableBalance(tableId)
      setBalanceCents(balance)
    } catch (error) {
      console.error('Failed to fetch table balance:', error)
    }

    setLoading(false)
  }, [tableId])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchBalance()
  }, [fetchBalance])

  return { loading, balanceCents, reload: fetchBalance }
}

export function useTableUnpaidProducts(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<Order['products']>([])

  const fetchUnpaidProducts = useCallback(async () => {
    setLoading(true)

    try {
      const products = await tableBackend.getTableUnpaidProducts(tableId)
      setProducts(products)
    } catch (error) {
      console.error('Failed to fetch unpaid products:', error)
    }

    setLoading(false)
  }, [tableId])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchUnpaidProducts()
  }, [fetchUnpaidProducts])

  return { loading, products, reload: fetchUnpaidProducts }
}
