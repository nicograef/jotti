import { Tooltip } from '@radix-ui/react-tooltip'
import {
  ChevronDown,
  ChevronUp,
  Hamburger,
  Pen,
  Plus,
  Shell,
  Wine,
} from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

import { NewVariantDialog } from './NewVariantDialog'
import {
  type Product,
  ProductCategory,
  type Variant,
  VariantStatus,
} from './Product'
import type { ProductBackend } from './ProductBackend'
import { VariantItem } from './VariantItem'

interface ProductItemProps {
  loading: boolean
  product: Product
  backend: Pick<
    ProductBackend,
    'activateVariant' | 'deactivateVariant' | 'createVariant' | 'updateVariant'
  >
  onEdit: (productId: number) => void
  onVariantCreated: (variant: Variant) => void
  onVariantUpdated: (variant: Variant) => void
  onVariantStatusChange: (variantId: number, status: VariantStatus) => void
}

export function ProductItem(props: ProductItemProps) {
  const [expanded, setExpanded] = useState(false)
  const [variantLoading, setVariantLoading] = useState(false)

  const activeVariantsCount = props.product.variants.filter(
    (v) => v.status === VariantStatus.ACTIVE,
  ).length

  const handleActivateVariant = async (variantId: number) => {
    setVariantLoading(true)
    try {
      await props.backend.activateVariant(variantId)
      props.onVariantStatusChange(variantId, VariantStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating variant:', error)
    }
    setVariantLoading(false)
  }

  const handleDeactivateVariant = async (variantId: number) => {
    setVariantLoading(true)
    try {
      await props.backend.deactivateVariant(variantId)
      props.onVariantStatusChange(variantId, VariantStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating variant:', error)
    }
    setVariantLoading(false)
  }

  return (
    <Item variant="outline" className="flex-col items-stretch">
      <div className="flex items-center gap-4">
        <div className="flex flex-col gap-3 shrink-0">
          <ProductCategoryIcon category={props.product.category} />
        </div>
        <ItemContent className="self-start flex-1">
          <ItemTitle>{props.product.name}</ItemTitle>
          <ItemDescription>
            {props.product.variants.length} Variante
            {props.product.variants.length !== 1 ? 'n' : ''} (
            {activeVariantsCount} aktiv)
          </ItemDescription>
          <ItemDescription>
            Erstellt am {new Date(props.product.createdAt).toLocaleDateString()}
          </ItemDescription>
        </ItemContent>
        <ItemActions className="flex gap-2">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="icon-sm"
                variant="outline"
                className="rounded-full cursor-pointer"
                aria-label="Edit Product"
                onClick={() => {
                  props.onEdit(props.product.id)
                }}
              >
                <Pen />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Bearbeiten</TooltipContent>
          </Tooltip>
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
          {props.product.variants.length === 0 ? (
            <p className="text-sm text-muted-foreground italic">
              Keine Varianten vorhanden
            </p>
          ) : (
            <div className="space-y-2">
              {props.product.variants.map((variant) => (
                <VariantItem
                  key={variant.id}
                  variant={variant}
                  loading={props.loading || variantLoading}
                  backend={props.backend}
                  onActivate={handleActivateVariant}
                  onDeactivate={handleDeactivateVariant}
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

function ProductCategoryIcon(props: { category: ProductCategory }) {
  switch (props.category) {
    case ProductCategory.FOOD:
      return (
        <Tooltip>
          <TooltipTrigger>
            <Hamburger size={32} className="stroke-primary" />
          </TooltipTrigger>
          <TooltipContent>Essen</TooltipContent>
        </Tooltip>
      )
    case ProductCategory.BEVERAGE:
      return (
        <Tooltip>
          <TooltipTrigger>
            <Wine size={32} className="stroke-primary" />
          </TooltipTrigger>
          <TooltipContent>Getränk</TooltipContent>
        </Tooltip>
      )
    case ProductCategory.OTHER:
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
