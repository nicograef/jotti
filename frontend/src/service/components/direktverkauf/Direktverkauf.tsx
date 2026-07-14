import { toast } from 'sonner'

import { useMengen } from '@/hooks/use-mengen'

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import type { Produkt } from '../../product/Produkt'
import { calculateTotalPrice, toBestellungData } from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  products: Produkt[]
  productsLoading: boolean
  onVerkauft?: () => void
}

export function Direktverkauf({
  backend,
  products,
  productsLoading,
  onVerkauft,
}: DirektverkaufProps) {
  const { mengen, add, remove, reset } = useMengen<number>()

  const { receiptItems, inputItems } = toBestellungData(products, mengen)
  const total = calculateTotalPrice(receiptItems)
  const anzahl = inputItems.reduce((sum, item) => sum + item.menge, 0)

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <DirektverkaufDrawer
        backend={backend}
        receiptItems={receiptItems}
        positionen={inputItems}
        anzahl={anzahl}
        totalCents={total}
        verkaufAbgeschlossen={() => {
          reset()
          toast.success('Verkauf abgeschlossen.')
          onVerkauft?.()
        }}
      />
      <ProductList
        products={products}
        variantMengen={mengen}
        onAdd={add}
        onRemove={remove}
      />
    </>
  )
}
