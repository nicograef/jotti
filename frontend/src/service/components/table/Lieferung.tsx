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
import { useTischUngeliefert } from '../../table/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { LieferungDrawer } from './LieferungDrawer'

interface LieferungProps {
  backend: Pick<TischBackend, 'produkteLiefern'>
  tisch: Tisch
  onProdukteGeliefert: () => void
}

export function Lieferung({
  tisch,
  backend,
  onProdukteGeliefert,
}: LieferungProps) {
  const { positionen, loading, reload } = useTischUngeliefert(tisch.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const undeliveredQuantities: Record<number, number> = {}
  positionen.forEach((position) => {
    undeliveredQuantities[position.id] = position.menge
  })

  const onAdd = (positionId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[positionId] || 0
      if (currentQuantity >= (undeliveredQuantities[positionId] || 0))
        return prev
      return {
        ...prev,
        [positionId]: currentQuantity + 1,
      }
    })
  }

  const onRemove = (positionId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[positionId] || 0
      if (currentQuantity <= 0) return prev
      return {
        ...prev,
        [positionId]: currentQuantity - 1,
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
        quantities={quantities}
        produkteGeliefert={() => {
          setQuantities({})
          toast.success(`Lieferung wurde registriert.`)
          onProdukteGeliefert()
          reload()
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
                key={position.id}
                position={position}
                menge={quantities[position.id] || 0}
                undeliveredQuantity={undeliveredQuantities[position.id] || 0}
                onAdd={() => {
                  onAdd(position.id)
                }}
                onRemove={() => {
                  onRemove(position.id)
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
  undeliveredQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  undeliveredQuantity,
  onAdd,
  onRemove,
}: PositionItemProps) {
  return (
    <Item key={position.id} variant="outline">
      <ItemContent>
        <ItemTitle>{position.name}</ItemTitle>
        <ItemDescription>
          noch {undeliveredQuantity - menge} zu liefern
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
