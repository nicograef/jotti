import { useIsMobile } from '@/hooks/use-mobile'
import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import {
  calculateAnzahl,
  calculateTotalPrice,
  toBestellungData,
} from '../drawerUtils'
import type { PfandVorgang } from '../pfand/pfandAuswahl'
import { ProductList, ProductListSkeleton } from '../ProductList'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufTabProps {
  products: Produkt[]
  pfandProdukte: Produkt[]
  // Von DirektverkaufPage gehoben, damit Korb, verkaufId, Pfand und rueckgabeId einen Blick in die Historie überstehen.
  vorgang: Vorgang<number>
  pfandVorgang: PfandVorgang
  productsLoading: boolean
  onErfolg: (nachricht: string) => void
}

export function DirektverkaufTab({
  products,
  pfandProdukte,
  vorgang,
  pfandVorgang,
  productsLoading,
  onErfolg,
}: DirektverkaufTabProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = vorgang.steuerung

  const { receiptItems, inputItems } = toBestellungData(
    products,
    mengen,
    vorgang.positionsKommentare,
  )
  const total = calculateTotalPrice(receiptItems)
  const anzahl = calculateAnzahl(inputItems)

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
            pfandProdukte={pfandProdukte}
            pfandVorgang={pfandVorgang}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={total}
            vorgang={vorgang}
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
        pfandProdukte={pfandProdukte}
        pfandVorgang={pfandVorgang}
        receiptItems={receiptItems}
        positionen={inputItems}
        anzahl={anzahl}
        totalCents={total}
        vorgang={vorgang}
        verkaufAbgeschlossen={verkaufAbgeschlossen}
        verkaufVerworfen={reset}
        pfandAusgezahlt={pfandAusgezahlt}
      />
      {productList}
    </>
  )
}
