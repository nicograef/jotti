import { Pencil, Trash2 } from 'lucide-react'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
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
  onDelete: (tischId: number) => Promise<void>
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
          Erstellt am{' '}
          {new Date(props.tisch.createdAt).toLocaleDateString('de-DE')}
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
        <AlertDialog>
          <Tooltip>
            <AlertDialogTrigger asChild>
              <TooltipTrigger asChild>
                <Button
                  size="icon-sm"
                  variant="outline"
                  className="rounded-full cursor-pointer text-destructive"
                >
                  <Trash2 />
                </Button>
              </TooltipTrigger>
            </AlertDialogTrigger>
            <TooltipContent>Löschen</TooltipContent>
          </Tooltip>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Tisch löschen?</AlertDialogTitle>
              <AlertDialogDescription>
                Der Tisch &ldquo;{props.tisch.name}&rdquo; wird unwiderruflich
                gelöscht.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
              <AlertDialogAction
                className="bg-destructive text-white hover:bg-destructive/90"
                onClick={() => void props.onDelete(props.tisch.id)}
              >
                Löschen
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </ItemActions>
    </Item>
  )
}
