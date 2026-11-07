import type { ReactNode } from 'react'

import {
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { formatRelativeTime } from '@/lib/utils'

import type { Bestellung } from './Bestellung'
import type { Umbuchung } from './Umbuchung'

// Bei einem Umbuchungs-Zugang ist `kommentar` der Richtungs-Autotext
// („Umbuchung von Tisch X") und damit der Vorgangstitel.
function quelleTitel(quelle: Bestellung | Umbuchung): string {
  return quelle.art === 'bestellung' ? 'Bestellung' : quelle.kommentar
}

function quelleZeitpunkt(quelle: Bestellung | Umbuchung): string {
  return quelle.art === 'bestellung' ? quelle.aufgenommenAm : quelle.umgebuchtAm
}

// children ist die drawer-spezifische Beschreibung unter dem Kopf.
export function QuelleDrawerHeader({
  quelle,
  children,
}: {
  quelle: Bestellung | Umbuchung
  children: ReactNode
}) {
  return (
    <DrawerHeader className="mx-auto w-full max-w-sm">
      <DrawerTitle>
        {quelleTitel(quelle)} · {formatRelativeTime(quelleZeitpunkt(quelle))} ·{' '}
        {quelle.userName}
      </DrawerTitle>
      <DrawerDescription>{children}</DrawerDescription>
    </DrawerHeader>
  )
}
