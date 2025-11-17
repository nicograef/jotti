import { Tooltip } from '@radix-ui/react-tooltip'
import { Star, UserPen } from 'lucide-react'

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
import { type User, UserRole, UserStatus } from '@/lib/UserBackend'

interface UserItemProps {
  loading: boolean
  user: User
  onEdit: (userId: number) => void
  onActivate: (userId: number) => Promise<void>
  onDeactivate: (userId: number) => Promise<void>
}

export function UserItem(props: UserItemProps) {
  const isActive = props.user.status === UserStatus.ACTIVE

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
                    void props.onActivate(props.user.id)
                  } else {
                    void props.onDeactivate(props.user.id)
                  }
                }}
              />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            {isActive
              ? 'Benutzer ist aktiv'
              : 'Benutzer ist derzeit deaktiviert'}
          </TooltipContent>
        </Tooltip>
      </ItemMedia>
      <ItemContent>
        <ItemTitle>
          {props.user.role === UserRole.ADMIN && (
            <Tooltip>
              <TooltipTrigger>
                <Star size={16} className="stroke-none fill-primary" />
              </TooltipTrigger>
              <TooltipContent>Administrator</TooltipContent>
            </Tooltip>
          )}
          {props.user.name}
          <Tooltip>
            <TooltipTrigger>
              <span className="ml-1 text-muted-foreground text-sm font-normal">
                {props.user.username}
              </span>
            </TooltipTrigger>
            <TooltipContent>Benutzername</TooltipContent>
          </Tooltip>
        </ItemTitle>
        <ItemDescription>
          Erstellt am {new Date(props.user.createdAt).toLocaleDateString()}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full cursor-pointer"
              aria-label="Edit User"
              onClick={() => {
                props.onEdit(props.user.id)
              }}
            >
              <UserPen />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Bearbeiten</TooltipContent>
        </Tooltip>
      </ItemActions>
    </Item>
  )
}
