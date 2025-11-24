import { ChevronRightIcon, Lamp } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router'

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
  tableBackend: Pick<TableBackend, 'getActiveTables'>
}

export function TableList(props: TableListProps) {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<TablePublic[]>([])

  useEffect(() => {
    async function fetchTables() {
      setLoading(true)
      try {
        const tables = await props.tableBackend.getActiveTables()
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
}

function TableListComponent(props: TableListComponentProps) {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.tables.map((table) => (
        <Item key={table.id} variant="outline" asChild>
          <Link to={`/service/tables/${table.id.toString()}`}>
            <ItemContent>
              <ItemTitle className="text-lg">
                <Lamp /> {table.name}
              </ItemTitle>
            </ItemContent>
            <ItemActions>
              <ChevronRightIcon />
            </ItemActions>
          </Link>
        </Item>
      ))}
    </ItemGroup>
  )
}

function TableListSkeleton() {
  return (
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
  )
}
