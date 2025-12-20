import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

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
