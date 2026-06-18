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
import { useMengen } from '@/hooks/use-mengen'
import { AuthSingleton } from '@/lib/Auth'
import { formatCents, formatPositionName } from '@/lib/utils'

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
  const [alleAnzeigen, setAlleAnzeigen] = useState(false)

  const unbezahlteMengen: Record<string, number> = {}
  positionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })

  const {
    mengen,
    add: onAdd,
    remove: onRemove,
    reset,
  } = useMengen<string>((positionId) => unbezahlteMengen[positionId] || 0)

  const meinePositionen = positionen.filter(
    (position) => position.bestellerUserId === AuthSingleton.userId,
  )
  const anderePositionen = positionen.filter(
    (position) => position.bestellerUserId !== AuthSingleton.userId,
  )

  const renderPosition = (position: Position, showBesteller: boolean) => (
    <PositionItem
      key={position.positionId}
      position={position}
      showBesteller={showBesteller}
      menge={mengen[position.positionId] || 0}
      unbezahlteMenge={unbezahlteMengen[position.positionId] || 0}
      onAdd={() => {
        onAdd(position.positionId)
      }}
      onRemove={() => {
        onRemove(position.positionId)
      }}
    />
  )

  return (
    <>
      {AuthSingleton.canCancel && (
        <div className={saldoCents < 0 ? 'mb-4' : 'mb-4 w-fit'}>
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
      <ZahlungDrawer
        backend={backend}
        tisch={tisch}
        unbezahltePositionen={positionen}
        mengen={mengen}
        zahlungKassiert={() => {
          reset()
          toast.success(`Zahlung erfolgreich.`)
          onZahlungKassiert()
        }}
      />
      {loading ? (
        <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
          {Array.from({ length: 6 }).map((_, index) => (
            // eslint-disable-next-line react-x/no-array-index-key
            <PositionItemSkeleton key={index} />
          ))}
        </ItemGroup>
      ) : (
        <div className="my-4 space-y-3">
          <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3">
            {meinePositionen.map((position) => renderPosition(position, false))}
          </ItemGroup>
          {anderePositionen.length > 0 && (
            <div className="space-y-2">
              <Button
                variant="outline"
                className="w-full"
                onClick={() => {
                  setAlleAnzeigen((offen) => !offen)
                }}
              >
                {alleAnzeigen
                  ? 'Weniger anzeigen'
                  : `Alle anzeigen (${anderePositionen.length.toString()} von anderen)`}
              </Button>
              {alleAnzeigen && (
                <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3">
                  {anderePositionen.map((position) =>
                    renderPosition(position, true),
                  )}
                </ItemGroup>
              )}
            </div>
          )}
        </div>
      )}
    </>
  )
}

interface PositionItemProps {
  position: Position
  menge: number
  unbezahlteMenge: number
  showBesteller: boolean
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  unbezahlteMenge,
  showBesteller,
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
          <span className="font-bold">
            {formatCents(position.einzelpreis)}&nbsp;€
          </span>
          &nbsp; &ndash; &nbsp;noch {unbezahlteMenge - menge} unbezahlt
          {showBesteller && (
            <span className="block text-muted-foreground">
              von {position.bestellerName}
            </span>
          )}
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
