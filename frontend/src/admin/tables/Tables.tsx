import { useState } from 'react'

import { ItemGroup } from '@/components/ui/item'

import { type Table, TableStatus } from './Table'
import type { TableBackend } from './TableBackend'
import { TableItem } from './TableItem'

interface TablesProps {
  loading: boolean
  backend: Pick<TableBackend, 'activateTable' | 'deactivateTable'>
  tables: Table[]
  onEdit: (tableId: number) => void
  onStatusChange: (tableId: number, status: TableStatus) => void
}

export function Tables(props: TablesProps) {
  const [loading, setLoading] = useState(props.loading)

  const activateTable = async (tableId: number) => {
    setLoading(true)
    try {
      await props.backend.activateTable(tableId)
      props.onStatusChange(tableId, TableStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating table:', error)
    }
    setLoading(false)
  }

  const deactivateTable = async (tableId: number) => {
    setLoading(true)
    try {
      await props.backend.deactivateTable(tableId)
      props.onStatusChange(tableId, TableStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating table:', error)
    }
    setLoading(false)
  }

  return (
    <>
      <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.tables.map((table) => (
          <TableItem
            key={table.id}
            loading={loading || props.loading}
            table={table}
            onActivate={activateTable}
            onDeactivate={deactivateTable}
            onEdit={props.onEdit}
          />
        ))}
      </ItemGroup>
    </>
  )
}
