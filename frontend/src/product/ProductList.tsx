import { ChevronRightIcon } from 'lucide-react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'

import { type Product } from './Product'

interface ProductListProps {
  products: Product[]
  onSelect: (productId: number) => void
}

export function ProductList(props: ProductListProps) {
  return (
    <>
      <h3>Tisch auswählen</h3>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.products.map((product) => (
          <Item
            key={product.id}
            variant="outline"
            onClick={() => {
              props.onSelect(product.id)
            }}
          >
            <ItemContent>
              <ItemTitle>{product.name}</ItemTitle>
            </ItemContent>
            <ItemActions>
              <ChevronRightIcon />
            </ItemActions>
          </Item>
        ))}
      </ItemGroup>
    </>
  )
}
