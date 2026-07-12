import { toast } from 'sonner'

import { useMengen } from '@/hooks/use-mengen'
import { formatPositionName } from '@/lib/utils'

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import type { Produkt } from '../../product/Produkt'
import { calculateTotalPrice } from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'
import type { ReceiptPosition } from '../table/Receipt'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  products: Produkt[]
  productsLoading: boolean
  onVerkauft?: () => void
}

interface SelectedItem {
  produktId: number
  varianteId: number
  name: string
  einzelpreisCents: number
  menge: number
}

function selectItems(
  products: Produkt[],
  mengen: Record<number, number>,
): SelectedItem[] {
  return products.flatMap((product) =>
    product.varianten
      .filter((variant) => (mengen[variant.id] || 0) > 0)
      .map((variant) => ({
        produktId: product.id,
        varianteId: variant.id,
        name: formatPositionName(product.name, variant.name),
        einzelpreisCents: variant.preisCents,
        menge: mengen[variant.id],
      })),
  )
}

export function Direktverkauf({
  backend,
  products,
  productsLoading,
  onVerkauft,
}: DirektverkaufProps) {
  const { mengen, add, remove, reset } = useMengen<number>()

  const items = selectItems(products, mengen)
  const total = calculateTotalPrice(items)
  const anzahl = items.reduce((sum, item) => sum + item.menge, 0)
  const receiptItems: ReceiptPosition[] = items.map((item) => ({
    name: item.name,
    einzelpreisCents: item.einzelpreisCents,
    menge: item.menge,
  }))

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  return (
    <>
      <DirektverkaufDrawer
        backend={backend}
        receiptItems={receiptItems}
        positionen={items.map((item) => ({
          produktId: item.produktId,
          varianteId: item.varianteId,
          menge: item.menge,
        }))}
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
