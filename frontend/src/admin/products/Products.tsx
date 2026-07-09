import { Package } from 'lucide-react'

import { EmptyState } from '@/components/common/EmptyState'
import { ItemGroup } from '@/components/ui/item'

import { adminListBottomClearance } from '../adminListLayout'
import { ProductItem } from './ProductItem'
import { type Produkt, type Variante, VarianteStatus } from './Produkt'
import { type ProduktBackend } from './ProduktBackend'

interface ProductsProps {
  loading: boolean
  backend: Pick<
    ProduktBackend,
    | 'aktiviereVariante'
    | 'deaktiviereVariante'
    | 'createVariante'
    | 'updateVariante'
    | 'deleteVariante'
  >
  products: Produkt[]
  onEdit: (produktId: number) => void
  onDelete: (produktId: number) => Promise<void>
  onVariantCreated: (produktId: number, variante: Variante) => void
  onVariantUpdated: (produktId: number, variante: Variante) => void
  onVariantDeleted: (produktId: number, varianteId: number) => void
  onVariantStatusChange: (
    produktId: number,
    varianteId: number,
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
    <ItemGroup
      className={`grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4 ${adminListBottomClearance}`}
    >
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
