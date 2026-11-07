import { useIsMobile } from '@/hooks/use-mobile'
import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { BestellungAbschluss } from './BestellungAbschluss'
import { BestellungDrawer } from './BestellungDrawer'
import {
  calculateAnzahl,
  calculateTotalPrice,
  toBestellungData,
} from './drawerUtils'
import { ProductList, ProductListSkeleton } from './ProductList'
import type { Tisch } from './Tisch'

interface BestellungTabProps {
  tisch: Tisch
  products: Produkt[]
  productsLoading: boolean
  // Bestell-Korb (Variante-ID → Menge) samt bestellungId, von TablePage gehoben.
  vorgang: Vorgang<number>
  onErfolg: (nachricht: string) => void
}

export function BestellungTab({
  tisch,
  products,
  productsLoading,
  vorgang,
  onErfolg,
}: BestellungTabProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = vorgang.steuerung
  const { receiptItems, inputItems } = toBestellungData(products, mengen)
  const totalCents = calculateTotalPrice(receiptItems)

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  const bestellungAufgenommen = () => {
    reset()
    onErfolg('Bestellung wurde aufgenommen.')
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
          <BestellungAbschluss
            variant="spalte"
            tisch={tisch}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={totalCents}
            vorgang={vorgang}
            bestellungAufgenommen={bestellungAufgenommen}
          />
        }
      />
    )
  }

  return (
    <>
      <BestellungDrawer
        tisch={tisch}
        receiptItems={receiptItems}
        positionen={inputItems}
        totalCents={totalCents}
        anzahl={calculateAnzahl(inputItems)}
        vorgang={vorgang}
        bestellungAufgenommen={bestellungAufgenommen}
      />
      {productList}
    </>
  )
}
