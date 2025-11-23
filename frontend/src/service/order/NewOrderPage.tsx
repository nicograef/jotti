import { useState } from 'react'

import { BackendSingleton } from '@/lib/Backend'
import type { TablePublic } from '@/table/Table'
import { TableBackend } from '@/table/TableBackend'

import { TableList } from './TableList'

const tableBackend = new TableBackend(BackendSingleton)

export function NewOrderPage() {
  const [selectedTable, setSelectedTable] = useState<TablePublic | null>(null)

  return (
    <>
      <h1 className="text-2xl font-bold">Neue Bestellung</h1>
      <br />
      {selectedTable && <p>Ausgewählter Tisch: {selectedTable.name}</p>}
      <TableList
        tableBackend={tableBackend}
        onSelect={(table) => {
          setSelectedTable(table)
        }}
      />
    </>
  )
}
