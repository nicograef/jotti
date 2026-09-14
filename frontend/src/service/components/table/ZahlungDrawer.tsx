import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { DockActionSlot } from '../ServiceDock'
import { DockActionButton } from './DockActionButton'
import { calculateTotalPrice, selectPositionen } from './drawerUtils'
import { RestbetragZeile } from './RestbetragZeile'
import { ZahlungAbschluss } from './ZahlungAbschluss'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  unbezahltePositionen: Position[]
  mengen: Record<string, number>
  restNachZahlungCents: number
  zahlungKassiert: () => void
}

// Handy-Container (unter lg): Dock-Aktionsbutton, Restbetrag-Zeile im Dock-Slot,
// Bottom-Sheet-Drawer. Ab lg trägt die Abschluss-Spalte (siehe Zahlung) beides.
export function ZahlungDrawer(props: ZahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const positionenToPay = selectPositionen(
    props.unbezahltePositionen,
    props.mengen,
  )
  const totalPrice = calculateTotalPrice(positionenToPay)
  const anzahl = positionenToPay.reduce((sum, p) => sum + p.menge, 0)
  const noPositionenSelected = positionenToPay.length === 0

  const onOpenChange = (isOpen: boolean) => {
    setOpen(noPositionenSelected ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DockActionSlot>
        <RestbetragZeile cents={props.restNachZahlungCents} />
      </DockActionSlot>
      <DrawerTrigger asChild>
        <DockActionButton
          label="Kassieren"
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={noPositionenSelected}
        />
      </DrawerTrigger>
      <ZahlungAbschluss
        variant="sheet"
        backend={props.backend}
        tisch={props.tisch}
        positionenToPay={positionenToPay}
        totalCents={totalPrice}
        restNachZahlungCents={props.restNachZahlungCents}
        zahlungKassiert={() => {
          setOpen(false)
          props.zahlungKassiert()
        }}
      />
    </Drawer>
  )
}
