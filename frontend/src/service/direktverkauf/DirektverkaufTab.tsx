import { useIsMobile } from '@/hooks/use-mobile'
import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { PfandBackend } from '../pfand/PfandBackend'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import {
  calculateAnzahl,
  calculateTotalPrice,
  toBestellungData,
} from '../table/drawerUtils'
import { ProductList, ProductListSkeleton } from '../table/ProductList'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'
import type { DirektverkaufBackend } from './DirektverkaufBackend'
import { DirektverkaufDrawer } from './DirektverkaufDrawer'

interface DirektverkaufTabProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  products: Produkt[]
  pfandProdukte: Produkt[]
  // Von DirektverkaufPage gehoben, damit Korb, verkaufId, Pfand und rueckgabeId einen Blick in die Historie überstehen.
  vorgang: Vorgang<number>
  pfandVorgang: PfandVorgang
  productsLoading: boolean
  onErfolg: (nachricht: string) => void
}

export function DirektverkaufTab({
  backend,
  pfandBackend,
  products,
  pfandProdukte,
  vorgang,
  pfandVorgang,
  productsLoading,
  onErfolg,
}: DirektverkaufTabProps) {
  const isMobile = useIsMobile()
  const { mengen, add, remove, reset } = vorgang.steuerung

  const { receiptItems, inputItems } = toBestellungData(products, mengen)
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
            backend={backend}
            pfandBackend={pfandBackend}
            pfandProdukte={pfandProdukte}
            pfandVorgang={pfandVorgang}
            receiptItems={receiptItems}
            positionen={inputItems}
            totalCents={total}
            schluesselFuer={vorgang.schluesselFuer}
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
        schluesselFuer={vorgang.schluesselFuer}
        verkaufAbgeschlossen={verkaufAbgeschlossen}
        verkaufVerworfen={reset}
        pfandAusgezahlt={pfandAusgezahlt}
      />
      {productList}
    </>
  )
}
