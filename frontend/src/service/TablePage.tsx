import { useParams } from 'react-router'

import { BackendSingleton } from '@/lib/Backend'
import { ProductBackend } from '@/product/ProductBackend'

import { ProductList } from './ProductList'

const productBackend = new ProductBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  return (
    <>
      <div>Table Page: {tableId}</div>
      <ProductList
        productBackend={productBackend}
        onSelect={(p) => {
          console.log(p)
        }}
      />
    </>
  )
}
