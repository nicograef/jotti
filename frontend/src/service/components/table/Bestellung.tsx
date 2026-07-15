import { useMengen } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'

import type { Produkt } from '../../product/Produkt'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { BestellungAbschluss } from './BestellungAbschluss'
import { BestellungDrawer } from './BestellungDrawer'
import { calculateTotalPrice, toBestellungData } from './drawerUtils'
import { ProductList, ProductListSkeleton } from './ProductList'

interface BestellungProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  products: Produkt[]
  productsLoading: boolean
  // Meldet die erfolgreiche Buchung samt Bestätigungstext an die Seite, die den
  // Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg: (nachricht: string) => void
}

export function Bestellung({
  backend,
  tisch,
  products,
  productsLoading,
  onErfolg,
}: BestellungProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = useMengen<number>()

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

  // Ab lg: feste Abschluss-Spalte rechts, Produkte links. Der extrahierte
  // Abschluss-Inhalt mountet genau einmal (isMobile entscheidet den Zweig).
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
            bestellungAufgenommen={bestellungAufgenommen}
          />
        }
      />
    )
  }

  // Unter lg: unverändert Dock-Aktionsbutton plus Bottom-Sheet-Drawer.
  return (
    <>
      <BestellungDrawer
        backend={backend}
        tisch={tisch}
        products={products}
        mengen={mengen}
        bestellungAufgenommen={bestellungAufgenommen}
      />
      {productList}
    </>
  )
}
