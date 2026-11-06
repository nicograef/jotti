import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'

import type { BestellPositionInput } from './Bestellung'
import { BestellungAbschluss } from './BestellungAbschluss'
import { DockActionButton } from './DockActionButton'
import type { ReceiptPosition } from './Receipt'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'

interface BestellungDrawerProps {
  backend: Pick<TischBackend, 'bestellungAufnehmen'>
  tisch: Tisch
  // Von BestellungTab einmal abgeleitet, wie für die Spalte.
  receiptItems: ReceiptPosition[]
  positionen: BestellPositionInput[]
  totalCents: number
  anzahl: number
  schluesselFuer: (inhalt: unknown) => string
  bestellungAufgenommen: () => void
}

// Handy-Container (unter lg); ab lg rendert Bestellung stattdessen die feste
// Abschluss-Spalte.
export function BestellungDrawer(props: BestellungDrawerProps) {
  const [open, setOpen] = useState(false)
  const noPositionenSelected = props.positionen.length === 0

  const onOpenChange = (isOpen: boolean) => {
    setOpen(noPositionenSelected ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Bestellung überprüfen"
          anzahl={props.anzahl}
          summeCents={props.totalCents}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <BestellungAbschluss
        variant="sheet"
        backend={props.backend}
        tisch={props.tisch}
        receiptItems={props.receiptItems}
        positionen={props.positionen}
        totalCents={props.totalCents}
        schluesselFuer={props.schluesselFuer}
        bestellungAufgenommen={() => {
          setOpen(false)
          props.bestellungAufgenommen()
        }}
      />
    </Drawer>
  )
}
