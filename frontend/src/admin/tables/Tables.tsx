import { useEffect, useState } from 'react'

import { TablesTable } from '@/admin/tables/TablesTable'
import { Card } from '@/components/ui/card'
import { BackendSingleton } from '@/lib/backend'
import { type Table, TableBackend } from '@/lib/TableBackend'

import { EditTableDialog } from './EditTableDialog'
import { NewTableDialog } from './NewTableDialog'
import { TableCreatedDialog } from './TableCreatedDialog'

const initialTableCreatedState = {
  table: null as Table | null,
  open: false,
}

const initialEditTableState = {
  table: null as Table | null,
  open: false,
}

export function Tables() {
  const [loading, setLoading] = useState(true)
  const [tables, setTables] = useState<Table[]>([])
  const [tableCreatedState, setTableCreatedState] = useState(
    initialTableCreatedState,
  )
  const [editTableState, setEditTableState] = useState(initialEditTableState)

  useEffect(() => {
    async function fetchTables() {
      const response = await new TableBackend(BackendSingleton).getTables()
      setTables(response)
      setLoading(false)
    }
    void fetchTables()
  }, [])

  const updateTable = (table: Table) => {
    setTables((prevTables) =>
      prevTables.map((t) => (t.id === table.id ? table : t)),
    )
  }

  return (
    <>
      <NewTableDialog
        created={(table) => {
          setTables((prevTables) => [...prevTables, table])
          setTableCreatedState({ table, open: true })
        }}
      />
      <TableCreatedDialog
        {...tableCreatedState}
        close={() => {
          setTableCreatedState(initialTableCreatedState)
        }}
      />
      {editTableState.table && (
        <EditTableDialog
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
      <Card className="p-0">
        <TablesTable
          tables={tables}
          loading={loading}
          onClick={(table) => {
            setEditTableState({ table, open: true })
          }}
        />
      </Card>
    </>
  )
}
