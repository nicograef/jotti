import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { LieferungDrawer } from './LieferungDrawer'

interface LieferungProps {
  backend: Pick<TischBackend, 'produkteLiefern'>
  tisch: Tisch
  positionen: Position[]
  loading: boolean
  onProdukteGeliefert: () => void
}

export function Lieferung({
  tisch,
  backend,
  positionen,
  loading,
  onProdukteGeliefert,
}: LieferungProps) {
  const [mengen, setMengen] = useState<Record<string, number>>({})

  const ungelieferteMengen: Record<string, number> = {}
  positionen.forEach((position) => {
    ungelieferteMengen[position.positionId] = position.menge
  })

  const onAdd = (positionId: string) => {
    setMengen((prev) => {
      const aktuelleMenge = prev[positionId] || 0
      if (aktuelleMenge >= (ungelieferteMengen[positionId] || 0)) return prev
      return {
        ...prev,
        [positionId]: aktuelleMenge + 1,
      }
    })
  }

  const onRemove = (positionId: string) => {
    setMengen((prev) => {
      const aktuelleMenge = prev[positionId] || 0
      if (aktuelleMenge <= 0) return prev
      return {
        ...prev,
        [positionId]: aktuelleMenge - 1,
      }
    })
  }

  return (
    <>
      {' '}
      <LieferungDrawer
        backend={backend}
        tisch={tisch}
        ungeliefertePositionen={positionen}
        mengen={mengen}
        produkteGeliefert={() => {
          setMengen({})
          toast.success(`Lieferung wurde registriert.`)
          onProdukteGeliefert()
        }}
      />
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 mt-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <PositionItemSkeleton key={index} />
            ))
          : positionen.map((position) => (
              <PositionItem
                key={position.positionId}
                position={position}
                menge={mengen[position.positionId] || 0}
                ungelieferteMenge={ungelieferteMengen[position.positionId] || 0}
                onAdd={() => {
                  onAdd(position.positionId)
                }}
                onRemove={() => {
                  onRemove(position.positionId)
                }}
              />
            ))}
      </ItemGroup>
    </>
  )
}

interface PositionItemProps {
  position: Position
  menge: number
  ungelieferteMenge: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  ungelieferteMenge,
  onAdd,
  onRemove,
}: PositionItemProps) {
  return (
    <Item key={position.positionId} variant="outline">
      <ItemContent>
        <ItemTitle>
          {position.produktName} {position.varianteName}
        </ItemTitle>
        <ItemDescription>
          noch {ungelieferteMenge - menge} zu liefern
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Produkt entfernen"
          onClick={onRemove}
        >
          <Minus />
        </Button>
        <span className="text-lg mx-1">{menge}</span>
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full"
          aria-label="Produkt hinzufügen"
          onClick={onAdd}
        >
          <Plus />
        </Button>
      </ItemActions>
    </Item>
  )
}

function PositionItemSkeleton() {
  return (
    <Item variant="outline">
      <ItemContent>
        <Skeleton className="h-4 w-24" />
      </ItemContent>
      <ItemActions>
        <Minus />
        <span className="text-lg mx-1">&nbsp;</span>
        <Plus />
      </ItemActions>
    </Item>
  )
}
