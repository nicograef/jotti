import { Minus, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { type ProductPublic } from '@/lib/product/Product'

interface ProductListComponentProps {
  products: ProductPublic[]
  productAmounts: Record<number, number>
  onAdd: (productId: number) => void
  onRemove: (productId: number) => void
}

export function ProductList(props: ProductListComponentProps) {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.products.map((product) => (
        <Item key={product.id} variant="outline">
          <ItemContent>
            <ItemTitle>{product.name}</ItemTitle>
            {product.description ? (
              <ItemDescription>
                <span className="font-bold">
                  {(product.netPriceCents / 100).toFixed(2)}&nbsp;€
                </span>
                &nbsp; &ndash; &nbsp;
                {product.description}
              </ItemDescription>
            ) : (
              <ItemDescription>
                <span className="font-bold">
                  {(product.netPriceCents / 100).toFixed(2)}&nbsp;€
                </span>
              </ItemDescription>
            )}
          </ItemContent>
          <ItemActions>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full"
              aria-label="Produkt entfernen"
              onClick={() => {
                props.onRemove(product.id)
              }}
            >
              <Minus />
            </Button>
            <span className="text-lg mx-1">
              {props.productAmounts[product.id] || 0}
            </span>
            <Button
              size="icon-sm"
              variant="outline"
              className="rounded-full"
              aria-label="Produkt hinzufügen"
              onClick={() => {
                props.onAdd(product.id)
              }}
            >
              <Plus />
            </Button>
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  )
}

export function ProductListSkeleton() {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {Array.from({ length: 6 }).map((_, index) => (
        <Item key={`skeleton-${index.toString()}`} variant="outline">
          <ItemContent>
            <Skeleton className="h-4 w-24" />
          </ItemContent>
          <ItemActions>
            <Plus />
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  )
}
