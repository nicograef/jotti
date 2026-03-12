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
import { AuthSingleton } from '@/lib/Auth'
import { formatCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import { useTischUnbezahlt } from '../../table/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { StornierungDrawer } from './StornierungDrawer'
import { ZahlungDrawer } from './ZahlungDrawer'

interface ZahlungProps {
  backend: Pick<TischBackend, 'zahlungRegistrieren' | 'produkteStornieren'>
  tisch: Tisch
  onZahlungRegistriert: () => void
  onProdukteStorniert: () => void
}

export function Zahlung({
  tisch,
  backend,
  onZahlungRegistriert,
  onProdukteStorniert,
}: ZahlungProps) {
  const { positionen, loading, reload } = useTischUnbezahlt(tisch.id)
  const [quantities, setQuantities] = useState<Record<number, number>>({})

  const unpaidQuantities: Record<number, number> = {}
  positionen.forEach((position) => {
    unpaidQuantities[position.id] = position.menge
  })

  const onAdd = (positionId: number) => {
    setQuantities((prev) => {
      const currentQuantity = prev[positionId] || 0
      if (currentQuantity >= (unpaidQuantities[positionId] || 0)) return prev
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
      <div className="flex gap-2">
        {AuthSingleton.canCancel && (
          <div className="flex-1">
            <StornierungDrawer
              backend={backend}
              tisch={tisch}
              unbezahltePositionen={positionen}
              quantities={quantities}
              produkteStorniert={() => {
                setQuantities({})
                toast.success(`Stornierung erfolgreich.`)
                onProdukteStorniert()
                reload()
              }}
            />
          </div>
        )}
        <div className="flex-1">
          <ZahlungDrawer
            backend={backend}
            tisch={tisch}
            unbezahltePositionen={positionen}
            quantities={quantities}
            zahlungRegistriert={() => {
              setQuantities({})
              toast.success(`Zahlung erfolgreich.`)
              onZahlungRegistriert()
              reload()
            }}
          />
        </div>
      </div>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
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
                unpaidQuantity={unpaidQuantities[position.id] || 0}
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
  unpaidQuantity: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  unpaidQuantity,
  onAdd,
  onRemove,
}: PositionItemProps) {
  return (
    <Item key={position.id} variant="outline">
      <ItemContent>
        <ItemTitle>{position.name}</ItemTitle>
        <ItemDescription>
          <span className="font-bold">
            {formatCents(position.preisCents)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;noch {unpaidQuantity - menge} unbezahlt
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
