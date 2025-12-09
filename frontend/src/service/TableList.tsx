import { ChevronRightIcon, Lamp } from 'lucide-react'
import { Link } from 'react-router'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'

import { type Table } from './table/Table'

interface TableListComponentProps {
  tables: Table[]
}

export function TableList(props: TableListComponentProps) {
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

export function TableListSkeleton() {
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
