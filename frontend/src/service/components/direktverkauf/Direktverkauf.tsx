import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
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
  // Das Erstladen der Produkte ist gescheitert (kein brauchbarer Cache-Stand).
  // Ein gescheiterter Hintergrund-Refetch setzt die Flagge nicht: Die zuletzt
  // geladenen Produkte bleiben stehen, die Meldung trägt der zentrale
  // Fehler-Toast.
  productsError: boolean
  // Lädt die Produkte nach einem Ladefehler erneut.
  onErneutVersuchen: () => void
  // Meldet den abgeschlossenen Verkauf samt Bestätigungstext an die Seite, die
  // den Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg?: (nachricht: string) => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`). Lädt die Historie neu, damit der Helfer den
  // tatsächlichen Stand sieht; die Auswahl leert diese Komponente selbst.
  onVorgangBereitsGebucht: () => void
}

export function Direktverkauf({
  backend,
  products,
  productsLoading,
  productsError,
  onErneutVersuchen,
  onErfolg,
  onVorgangBereitsGebucht,
}: DirektverkaufProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = useMengen<number>()

  const { receiptItems, inputItems } = toBestellungData(products, mengen)
  const total = calculateTotalPrice(receiptItems)
  const anzahl = inputItems.reduce((sum, item) => sum + item.menge, 0)

  // Eine leere Produktliste behauptet, es gebe nichts zu verkaufen — bei einem
  // Ladefehler ist das falsch.
  if (productsError) {
    return (
      <LadefehlerAlert
        titel="Produkte konnten nicht geladen werden"
        onErneutVersuchen={onErneutVersuchen}
        className="mt-4"
      />
    )
  }

  if (productsLoading) {
    return <ProductListSkeleton />
  }

  const verkaufAbgeschlossen = () => {
    reset()
    onErfolg?.('Verkauf abgeschlossen.')
  }

  // Die Auswahl muss auch hier leer werden: Erst ihr Leerzustand rotiert die
  // verkaufId, sonst liefe der nächste Versuch wieder in denselben 409.
  const vorgangBereitsGebucht = () => {
    reset()
    onVorgangBereitsGebucht()
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
            vorgangBereitsGebucht={vorgangBereitsGebucht}
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
        vorgangBereitsGebucht={vorgangBereitsGebucht}
      />
      {productList}
    </>
  )
}
