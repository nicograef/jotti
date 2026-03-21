import {
  ChevronDown,
  ChevronUp,
  Hamburger,
  Pen,
  Plus,
  Shell,
  Trash2,
  Wine,
} from 'lucide-react'
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
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import { NewVariantDialog } from './NewVariantDialog'
import {
  Kategorie,
  type Produkt,
  type Variante,
  VarianteStatus,
} from './Product'
import type { ProductBackend } from './ProductBackend'
import { VariantItem } from './VariantItem'

interface ProductItemProps {
  loading: boolean
  product: Produkt
  backend: Pick<
    ProductBackend,
    | 'activateVariant'
    | 'deactivateVariant'
    | 'createVariant'
    | 'updateVariant'
    | 'deleteVariant'
  >
  onEdit: (productId: number) => void
  onDelete: (productId: number) => Promise<void>
  onVariantCreated: (variant: Variante) => void
  onVariantUpdated: (variant: Variante) => void
  onVariantDeleted: (variantId: number) => void
  onVariantStatusChange: (variantId: number, status: VarianteStatus) => void
}

export function ProductItem(props: ProductItemProps) {
  const [expanded, setExpanded] = useState(false)
  const [variantLoading, setVariantLoading] = useState(false)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const handleDelete = async () => {
    setDeleteLoading(true)
    try {
      await props.onDelete(props.product.id)
    } catch (error) {
      console.error('Error deleting product:', error)
    }
    setDeleteLoading(false)
  }

  const activeVariantsCount = props.product.varianten.filter(
    (v) => v.status === VarianteStatus.ACTIVE,
  ).length

  const handleActivateVariant = async (variantId: number) => {
    setVariantLoading(true)
    try {
      await props.backend.activateVariant(variantId)
      props.onVariantStatusChange(variantId, VarianteStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating variant:', error)
    }
    setVariantLoading(false)
  }

  const handleDeactivateVariant = async (variantId: number) => {
    setVariantLoading(true)
    try {
      await props.backend.deactivateVariant(variantId)
      props.onVariantStatusChange(variantId, VarianteStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating variant:', error)
    }
    setVariantLoading(false)
  }

  const handleDeleteVariant = async (variantId: number) => {
    setVariantLoading(true)
    try {
      await props.backend.deleteVariant(props.product.id, variantId)
      props.onVariantDeleted(variantId)
    } catch (error) {
      console.error('Error deleting variant:', error)
    }
    setVariantLoading(false)
  }

  return (
    <Item variant="outline" className="flex-col items-stretch">
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex flex-col gap-3 shrink-0">
          <KategorieIcon category={props.product.kategorie} />
        </div>
        <ItemContent className="self-start flex-1">
          <ItemTitle>{props.product.name}</ItemTitle>
          <ItemDescription>
            {props.product.varianten.length} Variante
            {props.product.varianten.length !== 1 ? 'n' : ''} (
            {activeVariantsCount} aktiv)
          </ItemDescription>
          <ItemDescription>
            Erstellt am{' '}
            {new Date(props.product.createdAt).toLocaleDateString('de-DE')}
          </ItemDescription>
        </ItemContent>
        <ItemActions className="flex gap-2 w-full sm:w-auto justify-end">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="icon-sm"
                variant="outline"
                className="rounded-full cursor-pointer"
                aria-label="Produkt bearbeiten"
                onClick={() => {
                  props.onEdit(props.product.id)
                }}
              >
                <Pen />
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
                <AlertDialogTitle>Produkt löschen?</AlertDialogTitle>
                <AlertDialogDescription>
                  Das Produkt &quot;{props.product.name}&quot; und alle
                  zugehörigen Varianten werden unwiderruflich gelöscht.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Abbrechen</AlertDialogCancel>
                <AlertDialogAction
                  className="bg-destructive text-white hover:bg-destructive/90"
                  onClick={() => void handleDelete()}
                  disabled={deleteLoading}
                >
                  Löschen
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
          <Button
            size="sm"
            variant="ghost"
            className="cursor-pointer"
            onClick={() => {
              setExpanded(!expanded)
            }}
          >
            {expanded ? <ChevronUp /> : <ChevronDown />}
            {expanded ? 'Einklappen' : 'Varianten'}
          </Button>
        </ItemActions>
      </div>

      {expanded && (
        <div className="border-t mt-4 pt-4">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium text-muted-foreground">
              Produkt-Varianten
            </span>
            <NewVariantDialog
              productId={props.product.id}
              backend={props.backend}
              created={props.onVariantCreated}
            >
              <Button size="sm" variant="outline" className="cursor-pointer">
                <Plus className="h-4 w-4" /> Variante
              </Button>
            </NewVariantDialog>
          </div>
          {props.product.varianten.length === 0 ? (
            <p className="text-sm text-muted-foreground italic">
              Keine Varianten vorhanden
            </p>
          ) : (
            <div className="space-y-2">
              {props.product.varianten.map((variant) => (
                <VariantItem
                  key={variant.id}
                  variant={variant}
                  loading={props.loading || variantLoading}
                  backend={props.backend}
                  onActivate={handleActivateVariant}
                  onDeactivate={handleDeactivateVariant}
                  onDelete={handleDeleteVariant}
                  onUpdated={props.onVariantUpdated}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </Item>
  )
}

function KategorieIcon(props: { category: Kategorie }) {
  switch (props.category) {
    case Kategorie.ESSEN:
      return (
        <Tooltip>
          <TooltipTrigger>
            <Hamburger size={32} className="stroke-primary" />
          </TooltipTrigger>
          <TooltipContent>Essen</TooltipContent>
        </Tooltip>
      )
    case Kategorie.GETRAENK:
      return (
        <Tooltip>
          <TooltipTrigger>
            <Wine size={32} className="stroke-primary" />
          </TooltipTrigger>
          <TooltipContent>Getränk</TooltipContent>
        </Tooltip>
      )
    case Kategorie.SONSTIGES:
      return (
        <Tooltip>
          <TooltipTrigger>
            <Shell size={32} className="stroke-primary" />
          </TooltipTrigger>
          <TooltipContent>Sonstiges</TooltipContent>
        </Tooltip>
      )
  }
}
