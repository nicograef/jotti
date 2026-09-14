import type { ReactNode } from 'react'

import {
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { formatEuro, formatRelativeTime } from '@/lib/utils'

import type { Bestellung } from '../../table/Bestellung'
import type { Umbuchung } from '../../table/Umbuchung'
import { quelleTitel, quelleZeitpunkt } from './drawerUtils'

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
