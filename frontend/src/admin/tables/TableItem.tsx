import { Tooltip } from '@radix-ui/react-tooltip'
import { Pencil } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from '@/components/ui/item'
import { Switch } from '@/components/ui/switch'
import { TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { type Table, TableStatus } from '@/table/Table'

interface TableItemProps {
  loading: boolean
  table: Table
  onEdit: (tableId: number) => void
  onActivate: (tableId: number) => Promise<void>
  onDeactivate: (tableId: number) => Promise<void>
}

export function TableItem(props: TableItemProps) {
  const isActive = props.table.status === TableStatus.ACTIVE

  return (
    <Item variant="outline">
      <ItemMedia>
        <Tooltip>
          <TooltipTrigger asChild>
            <span>
              <Switch
                className="cursor-pointer"
                disabled={props.loading}
                checked={isActive}
                onCheckedChange={(checked) => {
                  if (checked) {
                    void props.onActivate(props.table.id)
                  } else {
                    void props.onDeactivate(props.table.id)
                  }
                }}
              />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            {isActive ? 'Tisch ist aktiv' : 'Tisch ist deaktiviert'}
          </TooltipContent>
        </Tooltip>
      </ItemMedia>
      <ItemContent>
        <ItemTitle>{props.table.name}</ItemTitle>
        <ItemDescription>
          Erstellt am {new Date(props.table.createdAt).toLocaleDateString()}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full cursor-pointer"
              aria-label="Edit Table"
              onClick={() => {
                props.onEdit(props.table.id)
              }}
            >
              <Pencil />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Bearbeiten</TooltipContent>
        </Tooltip>
      </ItemActions>
    </Item>
  )
}
