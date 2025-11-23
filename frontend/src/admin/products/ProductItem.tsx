import { Tooltip } from '@radix-ui/react-tooltip'
import { Hamburger, Pen, Shell, Wine } from 'lucide-react'

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
import { type Product, ProductCategory, ProductStatus } from '@/product/Product'

interface ProductItemProps {
  loading: boolean
  product: Product
  onEdit: (productId: number) => void
  onActivate: (productId: number) => Promise<void>
  onDeactivate: (productId: number) => Promise<void>
}

export function ProductItem(props: ProductItemProps) {
  const isActive = props.product.status === ProductStatus.ACTIVE
  return (
    <Item variant="outline">
      <ItemMedia className="flex flex-col gap-3">
        <Tooltip>
          <TooltipTrigger asChild>
            <span>
              <Switch
                className="cursor-pointer"
                disabled={props.loading}
                checked={isActive}
                onCheckedChange={(checked) => {
                  if (checked) {
                    void props.onActivate(props.product.id)
                  } else {
                    void props.onDeactivate(props.product.id)
                  }
                }}
              />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            {isActive ? 'Produkt ist aktiv' : 'Produkt ist deaktiviert'}
          </TooltipContent>
        </Tooltip>
        <ProductCategoryIcon category={props.product.category} />
      </ItemMedia>
      <ItemContent className="self-start">
        <ItemTitle>
          {props.product.name}
          <Tooltip>
            <TooltipTrigger>
              <span className="ml-1 text-muted-foreground text-sm font-normal">
                {props.product.netPrice.toFixed(2)} €
              </span>
            </TooltipTrigger>
            <TooltipContent>Netto-Preis</TooltipContent>
          </Tooltip>
        </ItemTitle>
        <ItemDescription>{props.product.description}</ItemDescription>
        <ItemDescription>
          Erstellt am {new Date(props.product.createdAt).toLocaleDateString()}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
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
      </ItemActions>
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
