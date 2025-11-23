import { ChevronRightIcon } from 'lucide-react'
import { useEffect, useState } from 'react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { type TablePublic } from '@/table/Table'
import type { TableBackend } from '@/table/TableBackend'

interface TableListProps {
  tableBackend: Pick<TableBackend, 'getTables'>
  onSelect: (table: TablePublic) => void
}

export function TableList(props: TableListProps) {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<TablePublic[]>([])

  useEffect(() => {
    async function fetchTables() {
      setLoading(true)
      try {
        const tables = await props.tableBackend.getTables()
        setTables(tables)
      } catch (error) {
        console.error('Failed to fetch tables:', error)
      }
      setLoading(false)
    }
    void fetchTables()
  }, [props.tableBackend])

  if (loading) {
    return <TableListSkeleton />
  }

  return <TableListComponent {...props} tables={tables} />
}

interface TableListComponentProps {
  tables: TablePublic[]
  onSelect: (table: TablePublic) => void
}

function TableListComponent(props: TableListComponentProps) {
  return (
    <>
      <h3>Tisch auswählen</h3>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.tables.map((table) => (
          <Item
            key={table.id}
            variant="outline"
            onClick={() => {
              props.onSelect(table)
            }}
          >
            <ItemContent>
              <ItemTitle>{table.name}</ItemTitle>
            </ItemContent>
            <ItemActions>
              <ChevronRightIcon />
            </ItemActions>
          </Item>
        ))}
      </ItemGroup>
    </>
  )
}

function TableListSkeleton() {
  return (
    <>
      <h3>Tisch auswählen</h3>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {Array.from({ length: 6 }).map((_, index) => (
          <Item key={`skeleton-${index.toString()}`} variant="outline">
            <ItemContent>
              <Skeleton className="h-4 w-24" />
            </ItemContent>
            <ItemActions>
              <ChevronRightIcon />
            </ItemActions>
          </Item>
        ))}
      </ItemGroup>
    </>
  )
}
