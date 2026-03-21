import { Pen, Trash2 } from 'lucide-react'
import { useState } from 'react'

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
import { Switch } from '@/components/ui/switch'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatCents } from '@/lib/utils'

import { EditVariantDialog } from './EditVariantDialog'
import { type Variante, VarianteStatus } from './Product'
import type { ProductBackend } from './ProductBackend'

interface VariantItemProps {
  variant: Variante
  loading: boolean
  backend: Pick<ProductBackend, 'updateVariant'>
  onActivate: (variantId: number) => Promise<void>
  onDeactivate: (variantId: number) => Promise<void>
  onDelete: (variantId: number) => Promise<void>
  onUpdated: (variant: Variante) => void
}

export function VariantItem(props: VariantItemProps) {
  const [editOpen, setEditOpen] = useState(false)
  const isActive = props.variant.status === VarianteStatus.ACTIVE

  return (
    <>
      <div className="flex items-center gap-3 p-2 rounded-md bg-muted/50">
        <Tooltip>
          <TooltipTrigger asChild>
            <span>
              <Switch
                className="cursor-pointer"
                disabled={props.loading}
                checked={isActive}
                onCheckedChange={(checked) => {
                  if (checked) {
                    void props.onActivate(props.variant.id)
                  } else {
                    void props.onDeactivate(props.variant.id)
                  }
                }}
              />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            {isActive ? 'Variante ist aktiv' : 'Variante ist deaktiviert'}
          </TooltipContent>
        </Tooltip>

        <div className="flex-1 min-w-0">
          <span className="font-medium">{props.variant.name}</span>
        </div>

        <span className="text-muted-foreground text-sm whitespace-nowrap">
          {formatCents(props.variant.preisCents)} €
        </span>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon-sm"
              variant="ghost"
              className="rounded-full cursor-pointer shrink-0"
              aria-label="Variante bearbeiten"
              onClick={() => {
                setEditOpen(true)
              }}
            >
              <Pen className="h-4 w-4" />
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
                  <Trash2 className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
            </AlertDialogTrigger>
            <TooltipContent>Löschen</TooltipContent>
          </Tooltip>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Variante löschen?</AlertDialogTitle>
              <AlertDialogDescription>
                Die Variante &quot;{props.variant.name}&quot; wird
                unwiderruflich gelöscht.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
              <AlertDialogAction
                className="bg-destructive text-white hover:bg-destructive/90"
                onClick={() => void props.onDelete(props.variant.id)}
              >
                Löschen
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      <EditVariantDialog
        open={editOpen}
        variant={props.variant}
        backend={props.backend}
        updated={props.onUpdated}
        close={() => {
          setEditOpen(false)
        }}
      />
    </>
  )
}
