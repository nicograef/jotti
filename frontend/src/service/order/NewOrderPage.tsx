import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { Spinner } from '@/components/ui/spinner'
import { BackendSingleton } from '@/lib/Backend'
import type { Table } from '@/table/Table'
import { TableBackend } from '@/table/TableBackend'

import { TableList } from './TableList'

const tableBackend = new TableBackend(BackendSingleton)

export function NewOrderPage() {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<Table[]>([])

  useEffect(() => {
    async function fetchTables() {
      setLoading(true)
      try {
        const tables = await tableBackend.getTables()
        setTables(tables)
      } catch (error) {
        console.error('Failed to fetch tables:', error)
      }
      setLoading(false)
    }
    void fetchTables()
  }, [])

  return (
    <>
      <h1 className="text-2xl font-bold">Neue Bestellung</h1>
      <br />
      {loading && <Spinner />}
      <TableList
        tables={tables}
        onSelect={(tableId) => {
          toast.success(`Tisch mit ID ${tableId.toString()} ausgewählt.`)
        }}
      />
    </>
  )
}
