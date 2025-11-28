import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { useAllTables } from '@/lib/table/hooks'
import type { Table, TableStatus } from '@/lib/table/Table'
import { TableBackend } from '@/lib/table/TableBackend'

import { EditTableDialog } from './EditTableDialog'
import { NewTableDialog } from './NewTableDialog'
import { Tables } from './Tables'

const initialEditState = {
  table: null as Table | null,
  open: false,
}

const tableBackend = new TableBackend(BackendSingleton)

export function AdminTablesPage() {
  const { loading, tables, setTables } = useAllTables()
  const [editState, setEditState] = useState(initialEditState)

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
      {editState.table && (
        <EditTableDialog
          backend={tableBackend}
          open={editState.open}
          table={editState.table}
          updated={(table) => {
            updateTable(table)
          }}
          close={() => {
            setEditState(initialEditState)
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
          setEditState({ table: tableToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
