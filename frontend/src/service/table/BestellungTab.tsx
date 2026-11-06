import { useIsMobile } from '@/hooks/use-mobile'
import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { BestellungAbschluss } from './BestellungAbschluss'
import { BestellungDrawer } from './BestellungDrawer'
import { calculateTotalPrice, toBestellungData } from './drawerUtils'
import { ProductList, ProductListSkeleton } from './ProductList'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'

interface BestellungTabProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  products: Produkt[]
  productsLoading: boolean
  // Bestell-Korb (Variante-ID → Menge) samt bestellungId, von TablePage gehoben.
  vorgang: Vorgang<number>
  onErfolg: (nachricht: string) => void
}

export function BestellungTab({
  backend,
  tisch,
  products,
  productsLoading,
  vorgang,
  onErfolg,
}: BestellungTabProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = vorgang.steuerung

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
    const { receiptItems, inputItems } = toBestellungData(products, mengen)
    return (
      <ServiceSplitLayout
        auswahl={productList}
        abschluss={
          <BestellungAbschluss
            variant="spalte"
            backend={backend}
            tisch={tisch}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={calculateTotalPrice(receiptItems)}
            schluesselFuer={vorgang.schluesselFuer}
            bestellungAufgenommen={bestellungAufgenommen}
          />
        }
      />
    )
  }

  return (
    <>
      <BestellungDrawer
        backend={backend}
        tisch={tisch}
        products={products}
        mengen={mengen}
        schluesselFuer={vorgang.schluesselFuer}
        bestellungAufgenommen={bestellungAufgenommen}
      />
      {productList}
    </>
  )
}
