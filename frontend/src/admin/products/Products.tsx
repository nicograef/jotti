import { Package } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'

import { type Product, type Variant, VariantStatus } from './Product'
import { type ProductBackend } from './ProductBackend'
import { ProductItem } from './ProductItem'

interface ProductsProps {
  loading: boolean
  backend: Pick<
    ProductBackend,
    'activateVariant' | 'deactivateVariant' | 'createVariant' | 'updateVariant'
  >
  products: Product[]
  onEdit: (productId: number) => void
  onVariantCreated: (productId: number, variant: Variant) => void
  onVariantUpdated: (productId: number, variant: Variant) => void
  onVariantStatusChange: (
    productId: number,
    variantId: number,
    status: VariantStatus,
  ) => void
}

export function Products(props: ProductsProps) {
  if (props.products.length === 0 && !props.loading) {
    return (
      <EmptyState
        icon={Package}
        title="Keine Produkte vorhanden"
        description="Erstelle ein neues Produkt und füge Varianten mit Preisen hinzu."
      />
    )
  }

  return (
    <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.products.map((product) => (
        <ProductItem
          key={product.id}
          loading={props.loading}
          product={product}
          backend={props.backend}
          onEdit={props.onEdit}
          onVariantCreated={(variant) => {
            props.onVariantCreated(product.id, variant)
          }}
          onVariantUpdated={(variant) => {
            props.onVariantUpdated(product.id, variant)
          }}
          onVariantStatusChange={(variantId, status) => {
            props.onVariantStatusChange(product.id, variantId, status)
          }}
        />
      ))}
    </ItemGroup>
  )
}
