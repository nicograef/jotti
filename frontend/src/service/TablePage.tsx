import { useParams } from 'react-router'

import { useActiveProducts } from '@/lib/product/hooks'
import { useTable } from '@/lib/table/hooks'

import { ProductList, ProductListSkeleton } from './ProductList'

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: productsLoading, products } = useActiveProducts()
  const { loading, table } = useTable(Number(tableId))

  if (loading) {
    return <div>Lade Tisch...</div>
  }

  return (
    <>
      <div>Table Page: {table?.name}</div>
      {productsLoading ? (
        <ProductListSkeleton />
      ) : (
        <ProductList
          products={products}
          onSelect={(p) => {
            console.log(p)
          }}
        />
      )}
    </>
  )
}
