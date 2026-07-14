import { ChevronDown, CircleCheck } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { useMengen } from '@/hooks/use-mengen'
import { AuthSingleton } from '@/lib/Auth'
import {
  cn,
  formatAlleAuswaehlenLabel,
  formatCents,
  formatPositionName,
} from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { Stepper } from '../Stepper'
import { ZahlungDrawer } from './ZahlungDrawer'

interface ZahlungProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  positionen: Position[]
  onZahlungKassiert: () => void
}

export function Zahlung({
  tisch,
  backend,
  positionen,
  onZahlungKassiert,
}: ZahlungProps) {
  const [andereOffen, setAndereOffen] = useState(false)

  const unbezahlteMengen: Record<string, number> = {}
  positionen.forEach((position) => {
    unbezahlteMengen[position.positionId] = position.menge
  })

  const {
    mengen,
    add: onAdd,
    remove: onRemove,
    reset,
    setAll,
  } = useMengen<string>((positionId) => unbezahlteMengen[positionId] || 0)

  const meinePositionen = positionen.filter(
    (position) => position.bestellerUserId === AuthSingleton.userId,
  )
  const anderePositionen = positionen.filter(
    (position) => position.bestellerUserId !== AuthSingleton.userId,
  )

  const auswahlSumme = positionen.reduce(
    (summe, position) =>
      summe + (mengen[position.positionId] || 0) * position.einzelpreisCents,
    0,
  )
  const restNachZahlung = tisch.saldoCents - auswahlSumme

  const alleEigenenVollAusgewaehlt =
    meinePositionen.length > 0 &&
    meinePositionen.every(
      (position) => (mengen[position.positionId] || 0) === position.menge,
    )

  const alleAuswaehlen = () => {
    if (alleEigenenVollAusgewaehlt) {
      reset()
      return
    }
    const naechste: Record<string, number> = {}
    meinePositionen.forEach((position) => {
      naechste[position.positionId] = position.menge
    })
    setAll(naechste)
  }

  const eigeneUnbezahltGesamt = meinePositionen.reduce(
    (summe, position) => summe + position.menge * position.einzelpreisCents,
    0,
  )

  const anderePositionenSumme = anderePositionen.reduce(
    (summe, position) => summe + position.menge * position.einzelpreisCents,
    0,
  )
  const anderePositionenNamen = anderePositionen
    .map((position) =>
      formatPositionName(position.produktName, position.varianteName),
    )
    .join(', ')
  const andereAusgewaehlteAnzahl = anderePositionen.reduce(
    (anzahl, position) => anzahl + (mengen[position.positionId] || 0),
    0,
  )
  const andereAusgewaehlteSumme = anderePositionen.reduce(
    (summe, position) =>
      summe + (mengen[position.positionId] || 0) * position.einzelpreisCents,
    0,
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
      <ZahlungDrawer
        backend={backend}
        tisch={tisch}
        unbezahltePositionen={positionen}
        mengen={mengen}
        restNachZahlungCents={restNachZahlung}
        zahlungKassiert={() => {
          reset()
          toast.success(`Zahlung erfolgreich.`)
          onZahlungKassiert()
        }}
      />
      <div className="my-4 space-y-3">
        {meinePositionen.length > 0 && (
          <button
            type="button"
            onClick={alleAuswaehlen}
            className="flex h-11 w-full items-center justify-center gap-2 rounded-lg border border-primary/50 bg-primary/5 px-4 text-sm font-medium text-primary"
          >
            <CircleCheck className="size-5" />
            {alleEigenenVollAusgewaehlt
              ? 'Auswahl aufheben'
              : formatAlleAuswaehlenLabel(
                  meinePositionen.length,
                  eigeneUnbezahltGesamt,
                )}
          </button>
        )}
        <ItemGroup className="grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3">
          {meinePositionen.map((position) => renderPosition(position, false))}
        </ItemGroup>
        {anderePositionen.length > 0 && (
          <div className="space-y-2">
            <button
              type="button"
              onClick={() => {
                setAndereOffen((offen) => !offen)
              }}
              className="flex w-full items-center justify-between gap-2 py-1 text-left"
              aria-expanded={andereOffen}
            >
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Von anderen · {anderePositionen.length}
              </span>
              <ChevronDown
                className={cn(
                  'size-4 text-muted-foreground transition-transform',
                  andereOffen && 'rotate-180',
                )}
              />
            </button>
            {andereOffen ? (
              <ItemGroup className="grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3">
                {anderePositionen.map((position) =>
                  renderPosition(position, true),
                )}
              </ItemGroup>
            ) : andereAusgewaehlteAnzahl > 0 ? (
              <p className="text-[13px] font-medium text-primary/80">
                {andereAusgewaehlteAnzahl} ausgewählt ·{' '}
                {formatCents(andereAusgewaehlteSumme)}&nbsp;€
              </p>
            ) : (
              <p className="text-[13px] text-muted-foreground">
                {anderePositionenNamen} · {formatCents(anderePositionenSumme)}
                &nbsp;€
              </p>
            )}
          </div>
        )}
      </div>
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
  const ausgewaehlt = menge > 0
  const zeilenSumme = menge * position.einzelpreisCents
  const unbezahltSumme = unbezahlteMenge * position.einzelpreisCents

  return (
    <Item
      key={position.positionId}
      variant="outline"
      className={cn(ausgewaehlt && 'border-primary/60 bg-primary/5')}
    >
      <ItemContent>
        <ItemTitle className="text-[15px]">
          {formatPositionName(position.produktName, position.varianteName)}
        </ItemTitle>
        <ItemDescription
          className={cn(
            'text-[13px]',
            ausgewaehlt && 'font-medium text-primary/80',
          )}
        >
          {ausgewaehlt
            ? `${menge.toString()} von ${unbezahlteMenge.toString()} ausgewählt · ${formatCents(zeilenSumme)}\u00A0€`
            : `${unbezahlteMenge.toString()} unbezahlt · ${formatCents(unbezahltSumme)}\u00A0€`}
          {showBesteller && (
            <span className="block text-muted-foreground">
              von {position.bestellerName}
            </span>
          )}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Stepper
          menge={menge}
          onAdd={onAdd}
          onRemove={onRemove}
          addLabel="Produkt hinzufügen"
          removeLabel="Produkt entfernen"
          addDisabled={menge >= unbezahlteMenge}
        />
      </ItemActions>
    </Item>
  )
}
