import { Plus } from 'lucide-react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { type ProductPublic } from '@/lib/product/Product'

interface ProductListComponentProps {
  products: ProductPublic[]
  onSelect: (product: ProductPublic) => void
}

export function ProductList(props: ProductListComponentProps) {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.products.map((product) => (
        <Item
          key={product.id}
          variant="outline"
          onClick={() => {
            props.onSelect(product)
          }}
        >
          <ItemContent>
            <ItemTitle>{product.name}</ItemTitle>
          </ItemContent>
          <ItemActions>
            <Plus />
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
