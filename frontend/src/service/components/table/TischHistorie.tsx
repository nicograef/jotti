import { Eye, X } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
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

import type { Ausgabe } from '../../table/Ausgabe'
import type { Bestellung } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import type { Zahlung } from '../../table/Zahlung'
import { Kommentar } from './CommentField'
import { toReceiptItems } from './drawerUtils'
import { HistorieStornierungDrawer } from './HistorieStornierungDrawer'
import { Receipt, type ReceiptPosition } from './Receipt'

interface TischHistorieProps {
  historie: (Bestellung | Zahlung | Stornierung | Ausgabe)[]
  historieLoading: boolean
  userId: number | null
  tisch: Tisch
  backend: Pick<TischBackend, 'stornierungErteilen'>
  onStornierungErteilt: () => void
}

const initialBestellungState: {
  bestellung: Bestellung | null
  open: boolean
} = {
  bestellung: null,
  open: false,
}

const initialZahlungState: {
  zahlung: Zahlung | null
  open: boolean
} = {
  zahlung: null,
  open: false,
}

const initialStornierungState: {
  stornierung: Stornierung | null
  open: boolean
} = {
  stornierung: null,
  open: false,
}

const initialAusgabeState: {
  ausgabe: Ausgabe | null
  open: boolean
} = {
  ausgabe: null,
  open: false,
}

export function TischHistorie({
  historie,
  historieLoading,
  userId,
  tisch,
  backend,
  onStornierungErteilt,
}: TischHistorieProps) {
  const [bestellung, setBestellung] = useState(initialBestellungState)
  const [zahlung, setZahlung] = useState(initialZahlungState)
  const [stornierung, setStornierung] = useState(initialStornierungState)
  const [ausgabe, setAusgabe] = useState(initialAusgabeState)
  const [stornierenBestellung, setStornierenBestellung] =
    useState<Bestellung | null>(null)

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historieLoading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : historie.map((item) => {
              if (Object.prototype.hasOwnProperty.call(item, 'kassiertAm')) {
                const zahlung = item as Zahlung
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Zahlung -${formatCents(zahlung.gesamtZahlungCents)} €`}
                    date={zahlung.kassiertAm}
                    isFromUser={userId === zahlung.userId}
                    kommentar={zahlung.kommentar}
                    onClick={() => {
                      setZahlung({ zahlung, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'aufgenommenAm')
              ) {
                const bestellung = item as Bestellung
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Bestellung +${formatCents(bestellung.gesamtPreisCents)} €`}
                    date={bestellung.aufgenommenAm}
                    isFromUser={userId === bestellung.userId}
                    kommentar={bestellung.kommentar}
                    onClick={() => {
                      setBestellung({ bestellung, open: true })
                    }}
                    onStornieren={
                      AuthSingleton.canCancel &&
                      getStornierbarePositionen(bestellung, historie).length > 0
                        ? () => {
                            setStornierenBestellung(bestellung)
                          }
                        : undefined
                    }
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'storniertAm')
              ) {
                const stornierung = item as Stornierung
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Stornierung -${formatCents(stornierung.gesamtStornierungCents)} €`}
                    date={stornierung.storniertAm}
                    isFromUser={userId === stornierung.userId}
                    kommentar={stornierung.kommentar}
                    onClick={() => {
                      setStornierung({ stornierung, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'ausgegebenAm')
              ) {
                const ausgabe = item as Ausgabe
                return (
                  <HistoryItem
                    key={item.id}
                    title="Ausgabe"
                    date={ausgabe.ausgegebenAm}
                    isFromUser={userId === ausgabe.userId}
                    kommentar={ausgabe.kommentar}
                    onClick={() => {
                      setAusgabe({ ausgabe, open: true })
                    }}
                  />
                )
              } else {
                return null
              }
            })}
      </ItemGroup>
      {bestellung.bestellung && (
        <Details
          title="Bestellung"
          id={bestellung.bestellung.id}
          isFromUser={userId === bestellung.bestellung.userId}
          open={bestellung.open}
          onClose={() => {
            setBestellung(initialBestellungState)
          }}
          date={bestellung.bestellung.aufgenommenAm}
          kommentar={bestellung.bestellung.kommentar}
          positionen={toReceiptItems(bestellung.bestellung.positionen)}
          totalPrice={bestellung.bestellung.gesamtPreisCents}
        />
      )}
      {zahlung.zahlung && (
        <Details
          title="Zahlung"
          id={zahlung.zahlung.id}
          isFromUser={userId === zahlung.zahlung.userId}
          open={zahlung.open}
          onClose={() => {
            setZahlung(initialZahlungState)
          }}
          date={zahlung.zahlung.kassiertAm}
          kommentar={zahlung.zahlung.kommentar}
          positionen={toReceiptItems(zahlung.zahlung.positionen)}
          totalPrice={zahlung.zahlung.gesamtZahlungCents}
        />
      )}
      {stornierung.stornierung && (
        <Details
          title="Stornierung"
          id={stornierung.stornierung.id}
          isFromUser={userId === stornierung.stornierung.userId}
          open={stornierung.open}
          onClose={() => {
            setStornierung(initialStornierungState)
          }}
          date={stornierung.stornierung.storniertAm}
          kommentar={stornierung.stornierung.kommentar}
          positionen={toReceiptItems(stornierung.stornierung.positionen)}
          totalPrice={stornierung.stornierung.gesamtStornierungCents}
        />
      )}
      {ausgabe.ausgabe && (
        <Details
          title="Ausgabe"
          id={ausgabe.ausgabe.id}
          isFromUser={userId === ausgabe.ausgabe.userId}
          open={ausgabe.open}
          onClose={() => {
            setAusgabe(initialAusgabeState)
          }}
          date={ausgabe.ausgabe.ausgegebenAm}
          kommentar={ausgabe.ausgabe.kommentar}
          positionen={toReceiptItems(ausgabe.ausgabe.positionen)}
        />
      )}
      {stornierenBestellung && (
        <HistorieStornierungDrawer
          backend={backend}
          tisch={tisch}
          bestellung={stornierenBestellung}
          positionen={getStornierbarePositionen(stornierenBestellung, historie)}
          onClose={() => {
            setStornierenBestellung(null)
          }}
          onStornierungErteilt={() => {
            setStornierenBestellung(null)
            onStornierungErteilt()
          }}
        />
      )}
    </>
  )
}

function getStornierbarePositionen(
  bestellung: Bestellung,
  historie: (Bestellung | Zahlung | Stornierung | Ausgabe)[],
) {
  const stornierteMengen = new Map<string, number>()

  historie.forEach((item) => {
    if (Object.prototype.hasOwnProperty.call(item, 'storniertAm')) {
      const stornierung = item as Stornierung
      stornierung.positionen.forEach((position) => {
        const bisherigeMenge = stornierteMengen.get(position.positionId) ?? 0
        stornierteMengen.set(
          position.positionId,
          bisherigeMenge + position.menge,
        )
      })
    }
  })

  return bestellung.positionen.flatMap((position) => {
    const verbleibendeMenge =
      position.menge - (stornierteMengen.get(position.positionId) ?? 0)
    if (verbleibendeMenge <= 0) {
      return []
    }

    return [{ ...position, menge: verbleibendeMenge }]
  })
}

function HistoryItem({
  title,
  date,
  isFromUser,
  kommentar,
  onClick,
  onStornieren,
}: {
  title: string
  date: string
  isFromUser: boolean
  kommentar: string
  onClick: () => void
  onStornieren?: () => void
}) {
  return (
    <Item variant="outline" className={isFromUser ? 'border-primary' : ''}>
      <ItemContent>
        <ItemTitle>{title}</ItemTitle>
        <ItemDescription>
          {new Date(date).toLocaleString('de-DE')}
          {kommentar && (
            <>
              <br />
              {kommentar}
            </>
          )}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        {onStornieren && (
          <Button
            size="icon-sm"
            variant="destructive"
            className="rounded-full cursor-pointer"
            aria-label="Stornieren"
            onClick={onStornieren}
          >
            <X />
          </Button>
        )}
        <Button
          size="icon-sm"
          variant="outline"
          className="rounded-full cursor-pointer"
          aria-label="Details anzeigen"
          onClick={onClick}
        >
          <Eye />
        </Button>
      </ItemActions>
    </Item>
  )
}

function ItemSkeleton() {
  return (
    <Item variant="outline">
      <ItemContent>
        <ItemTitle>
          <Skeleton className="h-6 w-32" />
        </ItemTitle>
        <Skeleton className="h-4 w-48" />
      </ItemContent>
      <ItemActions>
        <Skeleton className="h-8 w-8 rounded-full" />
      </ItemActions>
    </Item>
  )
}

function Details({
  open,
  onClose,
  title,
  id,
  date,
  isFromUser,
  kommentar,
  positionen,
  totalPrice,
  summary,
}: {
  open: boolean
  onClose: () => void
  title: string
  id: string
  date: string
  isFromUser: boolean
  kommentar: string
  positionen?: ReceiptPosition[]
  totalPrice?: number
  summary?: string
}) {
  return (
    <Drawer
      open={open}
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>
              {title} {id.slice(0, 8)}
            </DrawerTitle>
            <DrawerDescription>
              {isFromUser ? 'Du am ' : ''}
              {new Date(date).toLocaleDateString('de-DE')} um{' '}
              {new Date(date).toLocaleTimeString('de-DE')} Uhr
            </DrawerDescription>
          </DrawerHeader>
          {positionen ? (
            <Receipt positionen={positionen} totalPrice={totalPrice} />
          ) : (
            <div className="px-4 py-2 space-y-1">
              {summary && <p className="text-muted-foreground">{summary}</p>}
              {totalPrice !== undefined && (
                <p className="font-bold">{formatCents(totalPrice)}&nbsp;€</p>
              )}
            </div>
          )}
          {kommentar && (
            <div className="px-4">
              <Kommentar value={kommentar} />
            </div>
          )}
          <DrawerFooter>
            <DrawerClose asChild>
              <Button variant="outline">Schließen</Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
