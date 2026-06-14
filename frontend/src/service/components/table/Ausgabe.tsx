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
import { formatPositionName } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AusgabeDrawer } from './AusgabeDrawer'

interface AusgabeProps {
  backend: Pick<TischBackend, 'ausgabeBestaetigen'>
  tisch: Tisch
  positionen: Position[]
  loading: boolean
  onAusgabeBestaetigt: () => void
}

export function Ausgabe({
  tisch,
  backend,
  positionen,
  loading,
  onAusgabeBestaetigt,
}: AusgabeProps) {
  const [mengen, setMengen] = useState<Record<string, number>>({})

  const ausstehendeMengen: Record<string, number> = {}
  positionen.forEach((position) => {
    ausstehendeMengen[position.positionId] = position.menge
  })

  const onAdd = (positionId: string) => {
    setMengen((prev) => {
      const aktuelleMenge = prev[positionId] || 0
      if (aktuelleMenge >= (ausstehendeMengen[positionId] || 0)) return prev
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
      <AusgabeDrawer
        backend={backend}
        tisch={tisch}
        ausstehendePositionen={positionen}
        mengen={mengen}
        ausgabeBestaetigt={() => {
          setMengen({})
          toast.success(`Ausgabe wurde bestätigt.`)
          onAusgabeBestaetigt()
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
                ausstehendeMenge={ausstehendeMengen[position.positionId] || 0}
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
  ausstehendeMenge: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  ausstehendeMenge,
  onAdd,
  onRemove,
}: PositionItemProps) {
  return (
    <Item key={position.positionId} variant="outline">
      <ItemContent>
        <ItemTitle>
          {formatPositionName(position.produktName, position.varianteName)}
        </ItemTitle>
        <ItemDescription>
          noch {ausstehendeMenge - menge} ausstehend
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon"
          variant="outline"
          className="rounded-full min-h-12 min-w-12"
          aria-label="Produkt entfernen"
          onClick={onRemove}
        >
          <Minus />
        </Button>
        <span className="text-lg mx-2">{menge}</span>
        <Button
          size="icon"
          variant="outline"
          className="rounded-full min-h-12 min-w-12"
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
        <span className="text-lg mx-2">&nbsp;</span>
        <Plus />
      </ItemActions>
    </Item>
  )
}
