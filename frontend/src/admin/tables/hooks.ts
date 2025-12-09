import { useEffect, useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'

import type { Table } from './Table'
import { TableBackend } from './TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

/** Custom hook to fetch all tables from backend. */
export function useAllTables() {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<Table[]>([])

  useEffect(() => {
    async function fetchTables() {
      setLoading(true)

      try {
        const response = await tableBackend.getAllTables()
        setTables(response)
      } catch (error) {
        console.error('Failed to fetch tables:', error)
      }

      setLoading(false)
    }

    void fetchTables()
  }, [])

  return { loading, tables, setTables }
}
