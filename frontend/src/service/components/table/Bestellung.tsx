import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import type { MengenSteuerung } from '@/hooks/use-mengen'
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
  // Das Erstladen der Produkte ist gescheitert (kein brauchbarer Cache-Stand).
  // Ein gescheiterter Hintergrund-Refetch setzt die Flagge nicht: Die zuletzt
  // geladenen Produkte bleiben stehen, die Meldung trägt der zentrale
  // Fehler-Toast.
  productsError: boolean
  // Lädt die Produkte nach einem Ladefehler erneut.
  onErneutVersuchen: () => void
  // Bestell-Korb (Variante-ID → Menge), von TablePage gehoben, damit die
  // Auswahl das Aus- und Wiedereinhängen der Tab-Inhalte überlebt.
  mengenSteuerung: MengenSteuerung<number>
  // Idempotenz-Schlüssel dieser Zusammenstellung, aus demselben Grund wie der
  // Korb von TablePage gehoben: Ein hier gehaltener Schlüssel wechselte beim
  // Tab-Wechsel und machte aus einem Wiederholversuch eine zweite Buchung.
  bestellungId: string
  // Meldet die erfolgreiche Buchung samt Bestätigungstext an die Seite, die den
  // Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg: (nachricht: string) => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`). Räumt Korb und Tischzustand ab; beides liegt
  // in TablePage, deshalb kommt der Handler von dort.
  onVorgangBereitsGebucht: () => void
}

export function Bestellung({
  backend,
  tisch,
  products,
  productsLoading,
  productsError,
  onErneutVersuchen,
  mengenSteuerung,
  bestellungId,
  onErfolg,
  onVorgangBereitsGebucht,
}: BestellungProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = mengenSteuerung

  // Eine leere Produktliste behauptet, es gebe nichts zu bestellen — bei einem
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
            bestellungId={bestellungId}
            bestellungAufgenommen={bestellungAufgenommen}
            vorgangBereitsGebucht={onVorgangBereitsGebucht}
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
        bestellungId={bestellungId}
        bestellungAufgenommen={bestellungAufgenommen}
        vorgangBereitsGebucht={onVorgangBereitsGebucht}
      />
      {productList}
    </>
  )
}
