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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import { type Tisch, TischStatus } from './Tisch'

interface TischItemProps {
  loading: boolean
  tisch: Tisch
  onEdit: (tischId: number) => void
  onActivate: (tischId: number) => Promise<void>
  onDeactivate: (tischId: number) => Promise<void>
}

export function TischItem(props: TischItemProps) {
  const isActive = props.tisch.status === TischStatus.ACTIVE

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
                    void props.onActivate(props.tisch.id)
                  } else {
                    void props.onDeactivate(props.tisch.id)
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
        <ItemTitle>{props.tisch.name}</ItemTitle>
        <ItemDescription>
          Erstellt am {new Date(props.tisch.createdAt).toLocaleDateString()}
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
                props.onEdit(props.tisch.id)
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
