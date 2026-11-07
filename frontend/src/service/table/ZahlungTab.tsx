import { ChevronDown } from 'lucide-react'
import { useState } from 'react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { useEintritt } from '@/hooks/use-eintritt'
import { useErstAufbau } from '@/hooks/use-erst-aufbau'
import { useIsMobile } from '@/hooks/use-mobile'
import type { Vorgang } from '@/hooks/use-vorgang'
import { AuthSingleton } from '@/lib/Auth'
import type { Produkt } from '@/lib/produktSchemas'
import { cn, formatEuro, formatPositionName } from '@/lib/utils'

import type { PfandVorgang } from '../pfand/pfandAuswahl'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { Stepper } from '../Stepper'
import { AlleAuswaehlenButton } from './AlleAuswaehlenButton'
import type { Position } from './Bestellung'
import {
  calculateAnzahl,
  calculateTotalPrice,
  selectPositionen,
} from './drawerUtils'
import type { Tisch } from './Tisch'
import { ZahlungAbschluss } from './ZahlungAbschluss'
import { ZahlungDrawer } from './ZahlungDrawer'

interface ZahlungTabProps {
  pfandProdukte: Produkt[]
  // Wie die Kassieren-Auswahl von TablePage gehoben und je Tisch zurückgesetzt.
  pfandVorgang: PfandVorgang
  tisch: Tisch
  positionen: Position[]
  // Kassieren-Auswahl (Position-ID → Menge, gedeckelt auf die unbezahlte
  // Menge) samt Kommentar, von TablePage gehoben, damit sie den Tab-Wechsel überlebt.
  vorgang: Vorgang<string>
  onErfolg: (nachricht: string) => void
}

export function ZahlungTab({
  tisch,
  pfandProdukte,
  pfandVorgang,
  positionen,
  vorgang,
  onErfolg,
}: ZahlungTabProps) {
  const isMobile = useIsMobile()
  const [andereOffen, setAndereOffen] = useState(false)
  const erstAufbau = useErstAufbau(true)

  const { mengen, add, remove, reset } = vorgang.steuerung

  const meinePositionen: Position[] = []
  const anderePositionen: Position[] = []
  for (const position of positionen) {
    if (position.bestellerUserId === AuthSingleton.userId) {
      meinePositionen.push(position)
    } else {
      anderePositionen.push(position)
    }
  }

  const positionenToPay = selectPositionen(positionen, mengen)
  const auswahlSumme = calculateTotalPrice(positionenToPay)
  const auswahlAnzahl = calculateAnzahl(positionenToPay)
  const restNachZahlung = tisch.saldoCents - auswahlSumme

  const anderePositionenNamen = anderePositionen
    .map((position) =>
      formatPositionName(position.produktName, position.varianteName),
    )
    .join(', ')
  const andereAusgewaehlt = selectPositionen(anderePositionen, mengen)

  const renderPosition = (
    position: Position,
    showBesteller: boolean,
    eintrittIndex?: number,
  ) => (
    <PositionItem
      key={position.positionId}
      position={position}
      showBesteller={showBesteller}
      menge={mengen[position.positionId] || 0}
      unbezahlteMenge={position.menge}
      eintrittIndex={eintrittIndex}
      onAdd={() => {
        add(position.positionId)
      }}
      onRemove={() => {
        remove(position.positionId)
      }}
    />
  )

  const zahlungKassiert = () => {
    reset()
    onErfolg('Zahlung erfolgreich.')
  }

  const pfandAusgezahlt = () => {
    onErfolg('Pfand ausgezahlt.')
  }

  const auswahl = (
    <div className="my-4 space-y-3">
      {meinePositionen.length > 0 && (
        <AlleAuswaehlenButton
          positionen={meinePositionen}
          steuerung={vorgang.steuerung}
          variante="meine"
        />
      )}
      <ItemGroup className="grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3">
        {meinePositionen.map((position, index) =>
          renderPosition(position, false, erstAufbau ? index : undefined),
        )}
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
          ) : andereAusgewaehlt.length > 0 ? (
            <p className="text-[13px] font-medium text-primary/80">
              {calculateAnzahl(andereAusgewaehlt)} ausgewählt ·{' '}
              {formatEuro(calculateTotalPrice(andereAusgewaehlt))}
            </p>
          ) : (
            <p className="text-[13px] text-muted-foreground">
              {anderePositionenNamen} ·{' '}
              {formatEuro(calculateTotalPrice(anderePositionen))}
            </p>
          )}
        </div>
      )}
    </div>
  )

  // Ab lg: offene Positionen links, Abschluss-Spalte rechts; der Abschluss-Inhalt
  // mountet genau einmal.
  if (!isMobile) {
    return (
      <ServiceSplitLayout
        auswahl={auswahl}
        abschluss={
          <ZahlungAbschluss
            variant="spalte"
            pfandProdukte={pfandProdukte}
            pfandVorgang={pfandVorgang}
            tisch={tisch}
            positionenToPay={positionenToPay}
            totalCents={auswahlSumme}
            restNachZahlungCents={restNachZahlung}
            kommentar={vorgang.kommentar}
            setKommentar={vorgang.setKommentar}
            zahlungKassiert={zahlungKassiert}
            pfandAusgezahlt={pfandAusgezahlt}
          />
        }
      />
    )
  }

  return (
    <>
      <ZahlungDrawer
        pfandProdukte={pfandProdukte}
        pfandVorgang={pfandVorgang}
        tisch={tisch}
        positionenToPay={positionenToPay}
        totalCents={auswahlSumme}
        anzahl={auswahlAnzahl}
        restNachZahlungCents={restNachZahlung}
        kommentar={vorgang.kommentar}
        setKommentar={vorgang.setKommentar}
        zahlungKassiert={zahlungKassiert}
        pfandAusgezahlt={pfandAusgezahlt}
      />
      {auswahl}
    </>
  )
}

interface PositionItemProps {
  position: Position
  menge: number
  unbezahlteMenge: number
  showBesteller: boolean
  // Position in der Eintritts-Staffelung; `undefined` = kein animierter Eintritt.
  eintrittIndex?: number
  onAdd: () => void
  onRemove: () => void
}

function PositionItem({
  position,
  menge,
  unbezahlteMenge,
  showBesteller,
  eintrittIndex,
  onAdd,
  onRemove,
}: PositionItemProps) {
  const ausgewaehlt = menge > 0
  const zeilenSumme = menge * position.einzelpreisCents
  const unbezahltSumme = unbezahlteMenge * position.einzelpreisCents
  const eintritt = useEintritt(eintrittIndex)

  return (
    <Item
      key={position.positionId}
      variant="outline"
      style={eintritt.style}
      className={cn(
        ausgewaehlt && 'border-primary/60 bg-primary/5',
        eintritt.className,
      )}
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
            ? `${menge.toString()} von ${unbezahlteMenge.toString()} ausgewählt · ${formatEuro(zeilenSumme)}`
            : `${unbezahlteMenge.toString()} unbezahlt · ${formatEuro(unbezahltSumme)}`}
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
