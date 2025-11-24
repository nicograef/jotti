import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { type ProductPublic } from '@/product/Product'
import type { ProductBackend } from '@/product/ProductBackend'

interface ProductListProps {
  productBackend: Pick<ProductBackend, 'getActiveProducts'>
  onSelect: (product: ProductPublic) => void
}

export function ProductList(props: ProductListProps) {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<ProductPublic[]>([])

  useEffect(() => {
    async function fetchProducts() {
      setLoading(true)
      try {
        const products = await props.productBackend.getActiveProducts()
        setProducts(products)
      } catch (error) {
        console.error('Failed to fetch products:', error)
      }
      setLoading(false)
    }
    void fetchProducts()
  }, [props.productBackend])

  if (loading) {
    return <ProductListSkeleton />
  }

  return <ProductListComponent {...props} products={products} />
}

interface ProductListComponentProps {
  products: ProductPublic[]
  onSelect: (product: ProductPublic) => void
}

function ProductListComponent(props: ProductListComponentProps) {
  return (
    <>
      <h3>Produkt auswählen</h3>
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
    </>
  )
}

function ProductListSkeleton() {
  return (
    <>
      <h3>Produkt auswählen</h3>
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
    </>
  )
}
