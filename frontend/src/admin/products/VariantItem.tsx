import { Tooltip } from '@radix-ui/react-tooltip'
import { Pen } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { formatCents } from '@/lib/utils'

import { EditVariantDialog } from './EditVariantDialog'
import { type Variant, VariantStatus } from './Product'
import type { ProductBackend } from './ProductBackend'

interface VariantItemProps {
  variant: Variant
  loading: boolean
  backend: Pick<ProductBackend, 'updateVariant'>
  onActivate: (variantId: number) => Promise<void>
  onDeactivate: (variantId: number) => Promise<void>
  onUpdated: (variant: Variant) => void
}

export function VariantItem(props: VariantItemProps) {
  const [editOpen, setEditOpen] = useState(false)
  const isActive = props.variant.status === VariantStatus.ACTIVE

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
          {formatCents(props.variant.priceCents)} €
        </span>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon-sm"
              variant="ghost"
              className="rounded-full cursor-pointer shrink-0"
              aria-label="Edit Variant"
              onClick={() => {
                setEditOpen(true)
              }}
            >
              <Pen className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Bearbeiten</TooltipContent>
        </Tooltip>
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
