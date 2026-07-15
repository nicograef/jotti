import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import type { Produkt } from '../../product/Produkt'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { calculateTotalPrice, toBestellungData } from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  products: Produkt[]
  productsLoading: boolean
  // Meldet den abgeschlossenen Verkauf samt Bestätigungstext an die Seite, die
  // den Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg?: (nachricht: string) => void
}

export function Direktverkauf({
  backend,
  products,
  productsLoading,
  onErfolg,
}: DirektverkaufProps) {
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
    onErfolg?.('Verkauf abgeschlossen.')
  }

  const productList = (
    <ProductList
      products={products}
      variantMengen={mengen}
      onAdd={add}
      onRemove={remove}
    />
  )

  // Ab lg: feste Abschluss-Spalte rechts, Produkte links. Der extrahierte
  // Abschluss-Inhalt mountet genau einmal (isMobile entscheidet den Zweig).
  if (!isMobile) {
    return (
      <ServiceSplitLayout
        auswahl={productList}
        abschluss={
          <DirektverkaufAbschluss
            variant="spalte"
            backend={backend}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={total}
            verkaufAbgeschlossen={verkaufAbgeschlossen}
          />
        }
      />
    )
  }

  // Unter lg: unverändert Dock-Aktionsbutton plus Bottom-Sheet-Drawer.
  return (
    <>
      <DirektverkaufDrawer
        backend={backend}
        receiptItems={receiptItems}
        positionen={inputItems}
        anzahl={anzahl}
        totalCents={total}
        verkaufAbgeschlossen={verkaufAbgeschlossen}
      />
      {productList}
    </>
  )
}
