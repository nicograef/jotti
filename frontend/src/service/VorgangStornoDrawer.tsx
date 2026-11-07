import {
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'

import type { GebuchtePosition, PositionRef } from './schemas'
import { StornoDrawer } from './StornoDrawer'

interface VorgangStornoDrawerProps {
  // Der Titel lautet „<bezeichnung> stornieren“.
  bezeichnung: string
  offenePositionen: GebuchtePosition[]
  stornieren: (
    schluessel: string,
    positionen: PositionRef[],
    kommentar: string,
  ) => Promise<void>
  onClose: () => void
  onStorniert: () => void
}

export function VorgangStornoDrawer({
  bezeichnung,
  offenePositionen,
  ...rest
}: VorgangStornoDrawerProps) {
  return (
    <StornoDrawer
      {...rest}
      positionen={offenePositionen}
      kopf={
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>{bezeichnung} stornieren</DrawerTitle>
          <DrawerDescription>
            Die zu stornierenden Positionen auswählen.
          </DrawerDescription>
        </DrawerHeader>
      }
    />
  )
}
