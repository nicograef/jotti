import { ChevronRightIcon } from 'lucide-react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'

import { type Table } from '../TableBackend'

interface TableListProps {
  tables: Table[]
  onSelect: (tableId: number) => void
}

export function TableList(props: TableListProps) {
  return (
    <>
      <h3>Tisch auswählen</h3>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.tables.map((table) => (
          <Item
            key={table.id}
            variant="outline"
            onClick={() => {
              props.onSelect(table.id)
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
