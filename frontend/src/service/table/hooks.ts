import { useCallback, useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Cancelation } from './Cancelation'
import type { Delivery } from './Delivery'
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

/** Custom hook to fetch the history for a specific table from backend. */
export function useTableHistory(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [history, setHistory] = useState<
    (Order | Payment | Cancelation | Delivery)[]
  >([])

  useEffect(() => {
    async function fetchHistory() {
      setLoading(true)

      try {
        const history = await tableBackend.getTableHistory(tableId)
        setHistory(history)
      } catch (error) {
        console.error('Failed to fetch history:', error)
      }

      setLoading(false)
    }

    void fetchHistory()
  }, [tableId])

  return { loading, history }
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

export function useTableUndeliveredProducts(tableId: number) {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<Order['products']>([])

  const fetchUndeliveredProducts = useCallback(async () => {
    setLoading(true)

    try {
      const products = await tableBackend.getTableUndeliveredProducts(tableId)
      setProducts(products)
    } catch (error) {
      console.error('Failed to fetch undelivered products:', error)
    }

    setLoading(false)
  }, [tableId])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchUndeliveredProducts()
  }, [fetchUndeliveredProducts])

  return { loading, products, reload: fetchUndeliveredProducts }
}
