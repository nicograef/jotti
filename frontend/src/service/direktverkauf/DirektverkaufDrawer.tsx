import { useState } from 'react'

import { Drawer, DrawerTrigger } from '@/components/ui/drawer'
import type { Vorgang } from '@/hooks/use-vorgang'
import type { Produkt } from '@/lib/produktSchemas'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import { DockActionButton } from '../table/DockActionButton'
import type { ReceiptPosition } from '../table/Receipt'
import type { VerkaufPositionInput } from './Direktverkauf'
import { DirektverkaufAbschluss } from './DirektverkaufAbschluss'

interface DirektverkaufDrawerProps {
  pfandProdukte: Produkt[]
  pfandVorgang: PfandVorgang
  receiptItems: ReceiptPosition[]
  positionen: VerkaufPositionInput[]
  anzahl: number
  totalCents: number
  vorgang: Vorgang<number>
  verkaufAbgeschlossen: () => void
  verkaufVerworfen: () => void
  pfandAusgezahlt: () => void
}

// Handy-Container (unter lg); ab lg rendert Direktverkauf stattdessen die feste
// Abschluss-Spalte.
export function DirektverkaufDrawer(props: DirektverkaufDrawerProps) {
  const [open, setOpen] = useState(false)
  const noPositionenSelected = props.positionen.length === 0
  // Ohne Auswahl öffnet der Drawer nur für die Pfandrückgabe.
  const nurPfand = noPositionenSelected && props.pfandProdukte.length > 0
  const gesperrt = noPositionenSelected && !nurPfand

  const onOpenChange = (isOpen: boolean) => {
    setOpen(gesperrt ? false : isOpen)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerTrigger asChild>
        <DockActionButton
          label={nurPfand ? 'Pfand zurück' : 'Kassieren'}
          anzahl={props.anzahl}
          summeCents={props.totalCents}
          disabled={gesperrt}
        />
      </DrawerTrigger>
      <DirektverkaufAbschluss
        variant="sheet"
        pfandProdukte={props.pfandProdukte}
        pfandVorgang={props.pfandVorgang}
        receiptItems={props.receiptItems}
        positionen={props.positionen}
        totalCents={props.totalCents}
        vorgang={props.vorgang}
        verkaufAbgeschlossen={() => {
          setOpen(false)
          props.verkaufAbgeschlossen()
        }}
        verkaufVerworfen={() => {
          setOpen(false)
          props.verkaufVerworfen()
        }}
        pfandAusgezahlt={() => {
          setOpen(false)
          props.pfandAusgezahlt()
        }}
      />
    </Drawer>
  )
}
