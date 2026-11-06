import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandBackend } from '../pfand/PfandBackend'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { calculateTotalPrice, toBestellungData } from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'
import type { DirektverkaufBackend } from './DirektverkaufBackend'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufTabProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  products: Produkt[]
  pfandProdukte: Produkt[]
  productsLoading: boolean
  onErfolg: (nachricht: string) => void
}

export function DirektverkaufTab({
  backend,
  pfandBackend,
  products,
  pfandProdukte,
  productsLoading,
  onErfolg,
}: DirektverkaufTabProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = useMengen<number>()

  const { receiptItems, inputItems } = toBestellungData(products, mengen)
  const total = calculateTotalPrice(receiptItems)
  const anzahl = inputItems.reduce((sum, item) => sum + item.menge, 0)

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  const verkaufAbgeschlossen = () => {
    reset()
    onErfolg('Verkauf abgeschlossen.')
  }

  const pfandAusgezahlt = () => {
    onErfolg('Pfand ausgezahlt.')
  }

  const productList = (
    <ProductList
      products={products}
      variantMengen={mengen}
      onAdd={add}
      onRemove={remove}
    />
  )

  // Ab lg feste Abschluss-Spalte: der Abschluss-Inhalt mountet genau einmal.
  if (!isMobile) {
    return (
      <ServiceSplitLayout
        auswahl={productList}
        abschluss={
          <DirektverkaufAbschluss
            variant="spalte"
            backend={backend}
            pfandBackend={pfandBackend}
            pfandProdukte={pfandProdukte}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={total}
            verkaufAbgeschlossen={verkaufAbgeschlossen}
            pfandAusgezahlt={pfandAusgezahlt}
          />
        }
      />
    )
  }

  return (
    <>
      <DirektverkaufDrawer
        backend={backend}
        pfandBackend={pfandBackend}
        pfandProdukte={pfandProdukte}
        receiptItems={receiptItems}
        positionen={inputItems}
        anzahl={anzahl}
        totalCents={total}
        verkaufAbgeschlossen={verkaufAbgeschlossen}
        pfandAusgezahlt={pfandAusgezahlt}
      />
      {productList}
    </>
  )
}
