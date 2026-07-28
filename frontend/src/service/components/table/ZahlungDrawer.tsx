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
  // Idempotenz-Schlüssel dieser Zusammenstellung, von TablePage gehoben.
  vorgangId: string
  zahlungKassiert: () => void
  vorgangBereitsGebucht: () => void
}

// Handy-Container (unter lg): Dock-Aktionsbutton als Trigger, die Restbetrag-
// Zeile im Dock-Slot plus Bottom-Sheet-Drawer mit dem gemeinsamen
// Abschluss-Inhalt. Ab lg rendert die Fläche stattdessen die feste
// Abschluss-Spalte (siehe Zahlung), die die Restbetrag-Zeile selbst trägt.
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
        vorgangId={props.vorgangId}
        zahlungKassiert={() => {
          setOpen(false)
          props.zahlungKassiert()
        }}
        vorgangBereitsGebucht={() => {
          setOpen(false)
          props.vorgangBereitsGebucht()
        }}
      />
    </Drawer>
  )
}
