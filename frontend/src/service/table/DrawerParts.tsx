import type { ReactNode } from 'react'

import {
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { formatEuro, formatRelativeTime } from '@/lib/utils'

import type { Bestellung } from './Bestellung'
import { quelleTitel, quelleZeitpunkt } from './drawerUtils'
import type { Umbuchung } from './Umbuchung'

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

// betrag ist in Cent.
export function GesamtZeile({
  label,
  betrag,
}: {
  label: string
  betrag: number
}) {
  return (
    <div className="flex justify-between border-t-2 pt-2 font-bold">
      <div>{label}</div>
      <div>{formatEuro(betrag)}</div>
    </div>
  )
}
