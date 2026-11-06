import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import type { PfandBackend } from '../pfand/PfandBackend'
import { DockActionSlot } from '../ServiceDock'
import type { Position } from './Bestellung'
import { DockActionButton } from './DockActionButton'
import { RestbetragZeile } from './RestbetragZeile'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'
import { ZahlungAbschluss } from './ZahlungAbschluss'

interface ZahlungDrawerProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  pfandBackend: Pick<PfandBackend, 'pfandZurueckgeben'>
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  tisch: Tisch
  // Von ZahlungTab einmal abgeleitet, wie für die Spalte.
  positionenToPay: Position[]
  totalCents: number
  anzahl: number
  restNachZahlungCents: number
  zahlungKassiert: () => void
  pfandAusgezahlt: () => void
}

// Handy-Container (unter lg): Dock-Aktionsbutton, Restbetrag-Zeile im Dock-Slot,
// Bottom-Sheet-Drawer. Ab lg trägt die Abschluss-Spalte (siehe ZahlungTab) beides.
export function ZahlungDrawer(props: ZahlungDrawerProps) {
  const [open, setOpen] = useState(false)
  const noPositionenSelected = props.positionenToPay.length === 0
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
          anzahl={props.anzahl}
          summeCents={props.totalCents}
          disabled={gesperrt}
        />
      </DrawerTrigger>
      <ZahlungAbschluss
        variant="sheet"
        backend={props.backend}
        pfandBackend={props.pfandBackend}
        pfandProdukte={props.pfandProdukte}
        pfandVorgang={props.pfandVorgang}
        tisch={props.tisch}
        positionenToPay={props.positionenToPay}
        totalCents={props.totalCents}
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
