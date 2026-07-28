import { ChevronDown, CircleCheck } from 'lucide-react'
import { useState } from 'react'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { useErstAufbau } from '@/hooks/use-erst-aufbau'
import type { MengenSteuerung } from '@/hooks/use-mengen'
import { useIsMobile } from '@/hooks/use-mobile'
import { AuthSingleton } from '@/lib/Auth'
import {
  cn,
  formatAlleAuswaehlenLabel,
  formatEuro,
  formatPositionName,
} from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { ServiceSplitLayout } from '../ServiceSplitLayout'
import { Stepper } from '../Stepper'
import { selectPositionen } from './drawerUtils'
import { ZahlungAbschluss } from './ZahlungAbschluss'
import { ZahlungDrawer } from './ZahlungDrawer'

interface ZahlungProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  positionen: Position[]
  // Kassieren-Auswahl (Position-ID → Menge, gedeckelt auf die unbezahlte
  // Menge), von TablePage gehoben, damit sie den Tab-Wechsel überlebt.
  mengenSteuerung: MengenSteuerung<string>
  // Idempotenz-Schlüssel dieser Zusammenstellung, aus demselben Grund wie die
  // Auswahl von TablePage gehoben: Ein hier gehaltener Schlüssel wechselte beim
  // Tab-Wechsel und machte aus einem Wiederholversuch eine zweite Buchung.
  vorgangId: string
  // Meldet die erfolgreiche Zahlung samt Bestätigungstext an die Seite, die den
  // Erfolgs-Pop hostet (früher ein toast.success plus direkter Refetch).
  onErfolg: (nachricht: string) => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`). Räumt Auswahl und Tischzustand ab; beides liegt
  // in TablePage, deshalb kommt der Handler von dort.
  onVorgangBereitsGebucht: () => void
}

export function Zahlung({
  tisch,
  backend,
  positionen,
  mengenSteuerung,
  vorgangId,
  onErfolg,
  onVorgangBereitsGebucht,
}: ZahlungProps) {
  const isMobile = useIsMobile()
  const [andereOffen, setAndereOffen] = useState(false)
  // Positionen treten nur beim ersten Aufbau gestaffelt ein; nach einer Zahlung
  // (Refetch) bleiben die verbleibenden Zeilen unbewegt.
  const erstAufbau = useErstAufbau(true)

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
  } = mengenSteuerung

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
  // Die ausgewählten Positionen (Menge = Auswahl) für Beleg und Nutzlast der
  // Abschluss-Spalte; auswahlSumme ist deren Gesamtsumme.
  const positionenToPay = selectPositionen(positionen, mengen)

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
      unbezahlteMenge={unbezahlteMengen[position.positionId] || 0}
      eintrittIndex={eintrittIndex}
      onAdd={() => {
        onAdd(position.positionId)
      }}
      onRemove={() => {
        onRemove(position.positionId)
      }}
    />
  )

  const zahlungKassiert = () => {
    reset()
    onErfolg('Zahlung erfolgreich.')
  }

  const auswahl = (
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
                'meine',
              )}
        </button>
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
          ) : andereAusgewaehlteAnzahl > 0 ? (
            <p className="text-[13px] font-medium text-primary/80">
              {andereAusgewaehlteAnzahl} ausgewählt ·{' '}
              {formatEuro(andereAusgewaehlteSumme)}
            </p>
          ) : (
            <p className="text-[13px] text-muted-foreground">
              {anderePositionenNamen} · {formatEuro(anderePositionenSumme)}
            </p>
          )}
        </div>
      )}
    </div>
  )

  // Ab lg: offene Positionen links, Zahlungsübersicht rechts. Der extrahierte
  // Abschluss-Inhalt mountet genau einmal (isMobile entscheidet den Zweig).
  if (!isMobile) {
    return (
      <ServiceSplitLayout
        auswahl={auswahl}
        abschluss={
          <ZahlungAbschluss
            variant="spalte"
            backend={backend}
            tisch={tisch}
            positionenToPay={positionenToPay}
            totalCents={auswahlSumme}
            restNachZahlungCents={restNachZahlung}
            vorgangId={vorgangId}
            zahlungKassiert={zahlungKassiert}
            vorgangBereitsGebucht={onVorgangBereitsGebucht}
          />
        }
      />
    )
  }

  // Unter lg: unverändert Dock-Aktionsbutton (plus Restbetrag im Dock-Slot) und
  // Bottom-Sheet-Drawer.
  return (
    <>
      <ZahlungDrawer
        backend={backend}
        tisch={tisch}
        unbezahltePositionen={positionen}
        mengen={mengen}
        restNachZahlungCents={restNachZahlung}
        vorgangId={vorgangId}
        zahlungKassiert={zahlungKassiert}
        vorgangBereitsGebucht={onVorgangBereitsGebucht}
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
  // Position in der Eintritts-Staffelung (0-basiert) oder `undefined` ohne
  // animierten Eintritt (z. B. Fremdpositionen oder nach einem Refetch).
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
  // Beim Mount erfasst, damit ein Refetch die Staffelung nicht mittendrin abreißt.
  const [eintritt] = useState(eintrittIndex)

  return (
    <Item
      key={position.positionId}
      variant="outline"
      // Listen-Eintritt (Handoff): fadeUp 450 ms, 60 ms Stagger, weiche Kurve,
      // nur beim ersten Aufbau. Verzögerung dynamisch → inline.
      style={
        eintritt === undefined
          ? undefined
          : { animationDelay: `${(eintritt * 60).toString()}ms` }
      }
      className={cn(
        ausgewaehlt && 'border-primary/60 bg-primary/5',
        eintritt !== undefined &&
          'animate-fade-up [animation-duration:450ms] [animation-timing-function:cubic-bezier(0.2,0.7,0.3,1)]',
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
