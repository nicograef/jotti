import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'

import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { DockActionButton } from '../table/DockActionButton'
import type { ReceiptPosition } from '../table/Receipt'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'

interface VerkaufPositionInput {
  produktId: number
  varianteId: number
  menge: number
}

interface DirektverkaufDrawerProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  receiptItems: ReceiptPosition[]
  positionen: VerkaufPositionInput[]
  anzahl: number
  totalCents: number
  verkaufAbgeschlossen: () => void
}

// Handy-Container (unter lg): Dock-Aktionsbutton als Trigger plus
// Bottom-Sheet-Drawer, der den gemeinsamen Abschluss-Inhalt trägt. Ab lg rendert
// die Fläche stattdessen die feste Abschluss-Spalte (siehe Direktverkauf).
export function DirektverkaufDrawer(props: DirektverkaufDrawerProps) {
  const [open, setOpen] = useState(false)
  const noPositionenSelected = props.positionen.length === 0

  const onOpenChange = (isOpen: boolean) => {
    setOpen(noPositionenSelected ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Kassieren"
          anzahl={props.anzahl}
          summeCents={props.totalCents}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <DirektverkaufAbschluss
        variant="sheet"
        backend={props.backend}
        receiptItems={props.receiptItems}
        positionen={props.positionen}
        totalCents={props.totalCents}
        verkaufAbgeschlossen={() => {
          setOpen(false)
          props.verkaufAbgeschlossen()
        }}
      />
    </Drawer>
  )
}
