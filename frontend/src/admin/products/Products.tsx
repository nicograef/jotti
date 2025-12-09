import { useState } from 'react'

import { ItemGroup } from '@/components/ui/item'

import { type Product, ProductStatus } from './Product'
import { type ProductBackend } from './ProductBackend'
import { ProductItem } from './ProductItem'

interface ProductsProps {
  loading: boolean
  backend: Pick<ProductBackend, 'activateProduct' | 'deactivateProduct'>
  products: Product[]
  onEdit: (productId: number) => void
  onStatusChange: (productId: number, status: ProductStatus) => void
}

export function Products(props: ProductsProps) {
  const [loading, setLoading] = useState(props.loading)

  const activateProduct = async (productId: number) => {
    setLoading(true)
    try {
      await props.backend.activateProduct(productId)
      props.onStatusChange(productId, ProductStatus.ACTIVE)
    } catch (error) {
      console.error('Error activating user:', error)
    }
    setLoading(false)
  }

  const deactivateProduct = async (productId: number) => {
    setLoading(true)
    try {
      await props.backend.deactivateProduct(productId)
      props.onStatusChange(productId, ProductStatus.INACTIVE)
    } catch (error) {
      console.error('Error deactivating user:', error)
    }
    setLoading(false)
  }

  return (
    <>
      <ItemGroup className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {props.products.map((product) => (
          <ProductItem
            key={product.id}
            loading={loading || props.loading}
            product={product}
            onActivate={activateProduct}
            onDeactivate={deactivateProduct}
            onEdit={props.onEdit}
          />
        ))}
      </ItemGroup>
    </>
  )
}
