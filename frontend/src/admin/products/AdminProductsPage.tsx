import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { useAllProducts } from '@/lib/product/hooks'
import type { Product, ProductStatus } from '@/lib/product/Product'
import { ProductBackend } from '@/lib/product/ProductBackend'

import { EditProductDialog } from './EditProductDialog'
import { NewProductDialog } from './NewProductDialog'
import { Products } from './Products'

const initialEditState = {
  product: null as Product | null,
  open: false,
}

const productBackend = new ProductBackend(BackendSingleton)

export function AdminProductsPage() {
  const { loading, products, setProducts } = useAllProducts()
  const [editState, setEditState] = useState(initialEditState)

  const updateProduct = (product: Product) => {
    setProducts((prevProducts) =>
      prevProducts.map((u) => (u.id === product.id ? product : u)),
    )
  }

  const onStatusChange = (productId: number, status: ProductStatus) => {
    setProducts((prevProducts) =>
      prevProducts.map((u) => (u.id === productId ? { ...u, status } : u)),
    )
  }

  return (
    <>
      <NewProductDialog
        backend={productBackend}
        created={(product) => {
          setProducts((prevProducts) => [...prevProducts, product])
          toast.success(`Produkt "${product.name}" wurde angelegt.`)
        }}
      />
      {editState.product && (
        <EditProductDialog
          backend={productBackend}
          open={editState.open}
          product={editState.product}
          updated={(product) => {
            updateProduct(product)
          }}
          close={() => {
            setEditState(initialEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Produkte verwalten</h1>
      <Products
        loading={loading}
        backend={productBackend}
        products={products}
        onEdit={(productId) => {
          const productToEdit = products.find((u) => u.id === productId) ?? null
          setEditState({ product: productToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
