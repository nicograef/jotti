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
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AuszahlungDrawer } from './AuszahlungDrawer'
import { ZahlungDrawer } from './ZahlungDrawer'

interface ZahlungProps {
  backend: Pick<TischBackend, 'zahlungKassieren' | 'auszahlungLeisten'>
  tisch: Tisch
  positionen: Position[]
  saldoCents: number
  loading: boolean
  onZahlungKassiert: () => void
  onAuszahlungGeleistet: () => void
}

export function Zahlung({
  tisch,
  backend,
  positionen,
  saldoCents,
  loading,
  onZahlungKassiert,
  onAuszahlungGeleistet,
}: ZahlungProps) {
  const [mengen, setMengen] = useState<Record<string, number>>({})

  const unbezahlteMengen: Record<string, number> = {}
  positionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })

  const onAdd = (positionId: string) => {
    setMengen((prev) => {
      const aktuelleMenge = prev[positionId] || 0
      if (aktuelleMenge >= (unbezahlteMengen[positionId] || 0)) return prev
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
      <div className="flex gap-2">
        {AuthSingleton.canCancel && (
          <div className={saldoCents < 0 ? 'flex-1' : 'flex-none'}>
            <AuszahlungDrawer
              backend={backend}
              tisch={tisch}
              saldoCents={saldoCents}
              auszahlungGeleistet={() => {
                toast.success(`Auszahlung erfolgreich.`)
                onAuszahlungGeleistet()
              }}
            />
          </div>
        )}
        <div className="flex-1">
          <ZahlungDrawer
            backend={backend}
            tisch={tisch}
            unbezahltePositionen={positionen}
            mengen={mengen}
            zahlungKassiert={() => {
              setMengen({})
              toast.success(`Zahlung erfolgreich.`)
              onZahlungKassiert()
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
                key={position.positionId}
                position={position}
                menge={mengen[position.positionId] || 0}
                unbezahlteMenge={unbezahlteMengen[position.positionId] || 0}
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
  unbezahlteMenge: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  unbezahlteMenge,
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
          <span className="font-bold">
            {formatCents(position.einzelpreis)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;noch {unbezahlteMenge - menge} unbezahlt
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
