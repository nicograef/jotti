import { useEffect, useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { EditProductDialog } from './EditProductDialog'
import { NewProductDialog } from './NewProductDialog'
import { type Product, ProductBackend, ProductStatus } from './ProductBackend'
import { Products } from './Products'

const initialEditProductState = {
  product: null as Product | null,
  open: false,
}

const productBackend = new ProductBackend(BackendSingleton)

export function AdminProductsPage() {
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<Product[]>([])
  const [editProductState, setEditProductState] = useState(
    initialEditProductState,
  )

  useEffect(() => {
    async function fetchProducts() {
      setLoading(true)
      try {
        const response = await productBackend.getProducts()
        setProducts(response)
      } catch (error) {
        console.error('Failed to fetch products:', error)
      }
      setLoading(false)
    }
    void fetchProducts()
  }, [])

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
      {editProductState.product && (
        <EditProductDialog
          backend={productBackend}
          open={editProductState.open}
          product={editProductState.product}
          updated={(product) => {
            updateProduct(product)
          }}
          close={() => {
            setEditProductState(initialEditProductState)
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
          setEditProductState({ product: productToEdit, open: true })
        }}
        onStatusChange={onStatusChange}
      />
    </>
  )
}
