import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandBackend } from '../pfand/PfandBackend'
import type { PfandVorgang } from '../pfand/usePfandVorgang'
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
  // Von DirektverkaufPage gehoben, damit Pfand und rueckgabeId einen Blick in die Historie überstehen.
  pfandVorgang: PfandVorgang
  productsLoading: boolean
  onErfolg: (nachricht: string) => void
}

export function DirektverkaufTab({
  backend,
  pfandBackend,
  products,
  pfandProdukte,
  pfandVorgang,
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
            pfandVorgang={pfandVorgang}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={total}
            verkaufAbgeschlossen={verkaufAbgeschlossen}
            verkaufVerworfen={reset}
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
        pfandVorgang={pfandVorgang}
        receiptItems={receiptItems}
        positionen={inputItems}
        anzahl={anzahl}
        totalCents={total}
        verkaufAbgeschlossen={verkaufAbgeschlossen}
        verkaufVerworfen={reset}
        pfandAusgezahlt={pfandAusgezahlt}
      />
      {productList}
    </>
  )
}
