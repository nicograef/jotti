import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditTableDialog } from './EditTableDialog'
import { NewTableDialog } from './NewTableDialog'
import { type Table, TableBackend, TableStatus } from './TableBackend'
import { Tables } from './Tables'

const initialEditTableState = {
  table: null as Table | null,
  open: false,
}

const tableBackend = new TableBackend(BackendSingleton)

export function AdminTablesPage() {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<Table[]>([])
  const [editTableState, setEditTableState] = useState(initialEditTableState)

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

  const updateTable = (table: Table) => {
    setTables((prevTables) =>
      prevTables.map((t) => (t.id === table.id ? table : t)),
    )
  }

  const onStatusChange = (tableId: number, status: TableStatus) => {
    setTables((prevTables) =>
      prevTables.map((t) => (t.id === tableId ? { ...t, status } : t)),
    )
  }

  return (
    <>
      <NewTableDialog
        backend={tableBackend}
        created={(table) => {
          setTables((prevTables) => [...prevTables, table])
          toast.success(`Tisch "${table.name}" wurde angelegt.`)
        }}
      />
      {editTableState.table && (
        <EditTableDialog
          backend={tableBackend}
          open={editTableState.open}
          table={editTableState.table}
          updated={(table) => {
            updateTable(table)
          }}
          close={() => {
            setEditTableState(initialEditTableState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Tische verwalten</h1>
      <Tables
        loading={loading}
        backend={tableBackend}
        tables={tables}
        onEdit={(tableId) => {
          const tableToEdit = tables.find((t) => t.id === tableId) ?? null
          setEditTableState({ table: tableToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
