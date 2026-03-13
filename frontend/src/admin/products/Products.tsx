import { Package } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'

import { type Produkt, type Variante, VarianteStatus } from './Product'
import { type ProductBackend } from './ProductBackend'
import { ProductItem } from './ProductItem'

interface ProductsProps {
  loading: boolean
  backend: Pick<
    ProductBackend,
    | 'activateVariant'
    | 'deactivateVariant'
    | 'createVariant'
    | 'updateVariant'
    | 'deleteVariant'
  >
  products: Produkt[]
  onEdit: (productId: number) => void
  onDelete: (productId: number) => Promise<void>
  onVariantCreated: (productId: number, variant: Variante) => void
  onVariantUpdated: (productId: number, variant: Variante) => void
  onVariantDeleted: (productId: number, variantId: number) => void
  onVariantStatusChange: (
    productId: number,
    variantId: number,
    status: VarianteStatus,
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
          onDelete={props.onDelete}
          onVariantCreated={(variant) => {
            props.onVariantCreated(product.id, variant)
          }}
          onVariantUpdated={(variant) => {
            props.onVariantUpdated(product.id, variant)
          }}
          onVariantDeleted={(variantId) => {
            props.onVariantDeleted(product.id, variantId)
          }}
          onVariantStatusChange={(variantId, status) => {
            props.onVariantStatusChange(product.id, variantId, status)
          }}
        />
      ))}
    </ItemGroup>
  )
}
