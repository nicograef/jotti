import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandBackend } from '../pfand/PfandBackend'
import { DockActionSlot } from '../ServiceDock'
import type { Position } from './Bestellung'
import { DockActionButton } from './DockActionButton'
import { calculateTotalPrice, selectPositionen } from './drawerUtils'
import { RestbetragZeile } from './RestbetragZeile'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'
import { ZahlungAbschluss } from './ZahlungAbschluss'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  pfandProdukte: Produkt[]
  tisch: Tisch
  unbezahltePositionen: Position[]
  mengen: Record<string, number>
  restNachZahlungCents: number
  zahlungKassiert: () => void
  pfandAusgezahlt: () => void
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
  // Ohne Auswahl öffnet der Drawer nur für die Pfandrückgabe.
  const nurPfand = noPositionenSelected && props.pfandProdukte.length > 0
  const gesperrt = noPositionenSelected && !nurPfand

  const onOpenChange = (isOpen: boolean) => {
    setOpen(gesperrt ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DockActionSlot>
        <RestbetragZeile cents={props.restNachZahlungCents} />
      </DockActionSlot>
      <DrawerTrigger asChild>
        <DockActionButton
          label={nurPfand ? 'Pfand zurück' : 'Kassieren'}
          anzahl={anzahl}
          summeCents={totalPrice}
          disabled={gesperrt}
        />
      </DrawerTrigger>
      <ZahlungAbschluss
        variant="sheet"
        backend={props.backend}
        pfandBackend={props.pfandBackend}
        pfandProdukte={props.pfandProdukte}
        tisch={props.tisch}
        positionenToPay={positionenToPay}
        totalCents={totalPrice}
        restNachZahlungCents={props.restNachZahlungCents}
        zahlungKassiert={() => {
          setOpen(false)
          props.zahlungKassiert()
        }}
        pfandAusgezahlt={() => {
          setOpen(false)
          props.pfandAusgezahlt()
        }}
      />
    </Drawer>
  )
}
