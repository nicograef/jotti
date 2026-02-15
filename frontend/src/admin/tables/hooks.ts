import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Table } from './Table'
import { TableBackend } from './TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

/** Custom hook to fetch all tables from backend. */
export function useAllTables() {
  const { data: tables, setData: setTables, ...rest } = useFetch(
    () => tableBackend.getAllTables(),
    [] as Table[],
  )
  return { ...rest, tables, setTables }
}
