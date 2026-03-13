import { useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditProductDialog } from './EditProductDialog'
import { useAllProducts } from './hooks'
import { NewProductDialog } from './NewProductDialog'
import type { Produkt, Variante } from './Product'
import { ProductBackend } from './ProductBackend'
import { Products } from './Products'

const initialProductEditState = {
  product: null as Produkt | null,
  open: false,
}

const productBackend = new ProductBackend(BackendSingleton)

export function AdminProductsPage() {
  const { loading, products, setProducts } = useAllProducts()
  const [productEditState, setProductEditState] = useState(
    initialProductEditState,
  )

  const updateProduct = (product: Produkt) => {
    setProducts((prevProducts) =>
      prevProducts.map((p) => (p.id === product.id ? product : p)),
    )
  }

  const updateVariantInProduct = (
    productId: number,
    updater: (varianten: Variante[]) => Variante[],
  ) => {
    setProducts((prevProducts) =>
      prevProducts.map((p) =>
        p.id === productId ? { ...p, varianten: updater(p.varianten) } : p,
      ),
    )
  }

  const onVariantCreated = (productId: number, variant: Variante) => {
    updateVariantInProduct(productId, (varianten) => [...varianten, variant])
    toast.success(`Variante "${variant.name}" wurde angelegt.`)
  }

  const onVariantUpdated = (productId: number, variant: Variante) => {
    updateVariantInProduct(productId, (varianten) =>
      varianten.map((v) => (v.id === variant.id ? variant : v)),
    )
  }

  const onVariantStatusChange = (
    productId: number,
    variantId: number,
    status: 'active' | 'inactive',
  ) => {
    updateVariantInProduct(productId, (varianten) =>
      varianten.map((v) => (v.id === variantId ? { ...v, status } : v)),
    )
  }

  const onProductDelete = async (productId: number) => {
    await productBackend.deleteProduct(productId)
    setProducts((prev) => prev.filter((p) => p.id !== productId))
    toast.success('Produkt wurde gelöscht.')
  }

  const onVariantDeleted = (productId: number, variantId: number) => {
    updateVariantInProduct(productId, (varianten) =>
      varianten.filter((v) => v.id !== variantId),
    )
    toast.success('Variante wurde gelöscht.')
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
      {productEditState.product && (
        <EditProductDialog
          backend={productBackend}
          open={productEditState.open}
          product={productEditState.product}
          updated={(product) => {
            updateProduct(product)
          }}
          close={() => {
            setProductEditState(initialProductEditState)
          }}
        />
      )}
      <h1 className="text-2xl font-bold">Produkte verwalten</h1>
      <Products
        loading={loading}
        backend={productBackend}
        products={products}
        onEdit={(productId) => {
          const productToEdit = products.find((p) => p.id === productId) ?? null
          setProductEditState({ product: productToEdit, open: true })
        }}
        onDelete={onProductDelete}
        onVariantCreated={onVariantCreated}
        onVariantUpdated={onVariantUpdated}
        onVariantDeleted={onVariantDeleted}
        onVariantStatusChange={onVariantStatusChange}
      />
    </>
  )
}
