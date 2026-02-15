import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import type { Cancelation } from './Cancelation'
import type { Delivery } from './Delivery'
import type { LineItem, Order } from './Order'
import type { Payment } from './Payment'
import type { Table } from './Table'
import { TableBackend } from './TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

/** Custom hook to fetch a single table from backend. */
export function useTable(id: number) {
  const { data: table, ...rest } = useFetch(
    () => tableBackend.getTable(id),
    null as Table | null,
    [id],
  )
  return { ...rest, table }
}

/** Custom hook to fetch active tables from backend. */
export function useActiveTables() {
  const { data: tables, ...rest } = useFetch(
    () => tableBackend.getActiveTables(),
    [] as Table[],
  )
  return { ...rest, tables }
}

/** Custom hook to fetch the history for a specific table from backend. */
export function useTableHistory(tableId: number) {
  const { data: history, ...rest } = useFetch(
    () => tableBackend.getTableHistory(tableId),
    [] as (Order | Payment | Cancelation | Delivery)[],
    [tableId],
  )
  return { ...rest, history }
}

export function useTableBalance(tableId: number) {
  const { data: balanceCents, ...rest } = useFetch(
    () => tableBackend.getTableBalance(tableId),
    0,
    [tableId],
  )
  return { ...rest, balanceCents }
}

export function useTableUnpaidVariants(tableId: number) {
  const { data: variants, ...rest } = useFetch(
    () => tableBackend.getTableUnpaidVariants(tableId),
    [] as LineItem[],
    [tableId],
  )
  return { ...rest, variants }
}

export function useTableUndeliveredVariants(tableId: number) {
  const { data: variants, ...rest } = useFetch(
    () => tableBackend.getTableUndeliveredVariants(tableId),
    [] as LineItem[],
    [tableId],
  )
  return { ...rest, variants }
}
