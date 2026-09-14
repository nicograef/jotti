import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'
import type { Produkt } from '@/lib/produktSchemas'

import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { BestellungAbschluss } from './BestellungAbschluss'
import { DockActionButton } from './DockActionButton'
import { calculateTotalPrice, toBestellungData } from './drawerUtils'

interface BestellungDrawerProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  products: Produkt[]
  mengen: Record<number, number>
  bestellungAufgenommen: () => void
}

// Handy-Container (unter lg); ab lg rendert Bestellung stattdessen die feste
// Abschluss-Spalte.
export function BestellungDrawer(props: BestellungDrawerProps) {
  const [open, setOpen] = useState(false)
  const { receiptItems, inputItems } = toBestellungData(
    props.products,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(receiptItems)
  const anzahl = inputItems.reduce((sum, item) => sum + item.menge, 0)
  const noPositionenSelected = inputItems.length === 0

  const onOpenChange = (isOpen: boolean) => {
    setOpen(noPositionenSelected ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Bestellung überprüfen"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <BestellungAbschluss
        variant="sheet"
        backend={props.backend}
        tisch={props.tisch}
        receiptItems={receiptItems}
        positionen={inputItems}
        totalCents={totalPrice}
        bestellungAufgenommen={() => {
          setOpen(false)
          props.bestellungAufgenommen()
        }}
      />
    </Drawer>
  )
}
