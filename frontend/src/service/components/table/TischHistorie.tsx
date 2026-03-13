import { Eye } from 'lucide-react'
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
import { formatCents } from '@/lib/utils'

import type { Bestellung } from '../../table/Bestellung'
import { useTischHistorie } from '../../table/hooks'
import type { Lieferung } from '../../table/Lieferung'
import type { Stornierung } from '../../table/Stornierung'
import type { Zahlung } from '../../table/Zahlung'
import { Kommentar } from './CommentField'
import { toReceiptItems } from './drawerUtils'
import { Receipt, type ReceiptPosition } from './Receipt'

interface TischHistorieProps {
  tischId: number
  userId: number | null
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

const initialLieferungState: {
  lieferung: Lieferung | null
  open: boolean
} = {
  lieferung: null,
  open: false,
}

export function TischHistorie({ tischId, userId }: TischHistorieProps) {
  const { loading, historie } = useTischHistorie(tischId)
  const [bestellung, setBestellung] = useState(initialBestellungState)
  const [zahlung, setZahlung] = useState(initialZahlungState)
  const [stornierung, setStornierung] = useState(initialStornierungState)
  const [lieferung, setLieferung] = useState(initialLieferungState)

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {loading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : historie.map((item) => {
              if (Object.prototype.hasOwnProperty.call(item, 'registriertAm')) {
                const zahlung = item as Zahlung
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Zahlung -${formatCents(zahlung.gesamtZahlungCents)} €`}
                    date={zahlung.registriertAm}
                    isFromUser={userId === zahlung.userId}
                    kommentar={zahlung.kommentar}
                    onClick={() => {
                      setZahlung({ zahlung, open: true })
                    }}
                  />
                )
              } else if (
                Object.prototype.hasOwnProperty.call(item, 'aufgegebenAm')
              ) {
                const bestellung = item as Bestellung
                return (
                  <HistoryItem
                    key={item.id}
                    title={`Bestellung +${formatCents(bestellung.gesamtPreisCents)} €`}
                    date={bestellung.aufgegebenAm}
                    isFromUser={userId === bestellung.userId}
                    kommentar={bestellung.kommentar}
                    onClick={() => {
                      setBestellung({ bestellung, open: true })
                    }}
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
                Object.prototype.hasOwnProperty.call(item, 'geliefertAm')
              ) {
                const lieferung = item as Lieferung
                return (
                  <HistoryItem
                    key={item.id}
                    title="Auslieferung"
                    date={lieferung.geliefertAm}
                    isFromUser={userId === lieferung.userId}
                    kommentar={lieferung.kommentar}
                    onClick={() => {
                      setLieferung({ lieferung, open: true })
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
          date={bestellung.bestellung.aufgegebenAm}
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
          date={zahlung.zahlung.registriertAm}
          kommentar={zahlung.zahlung.kommentar}
          summary={`${zahlung.zahlung.positionen.length.toString()} Positionen`}
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
          summary={`${stornierung.stornierung.positionen.length.toString()} Positionen`}
          totalPrice={stornierung.stornierung.gesamtStornierungCents}
        />
      )}
      {lieferung.lieferung && (
        <Details
          title="Auslieferung"
          id={lieferung.lieferung.id}
          isFromUser={userId === lieferung.lieferung.userId}
          open={lieferung.open}
          onClose={() => {
            setLieferung(initialLieferungState)
          }}
          date={lieferung.lieferung.geliefertAm}
          kommentar={lieferung.lieferung.kommentar}
          summary={`${lieferung.lieferung.positionen.length.toString()} Positionen ausgeliefert`}
        />
      )}
    </>
  )
}

function HistoryItem({
  title,
  date,
  isFromUser,
  kommentar,
  onClick,
}: {
  title: string
  date: string
  isFromUser: boolean
  kommentar: string
  onClick: () => void
}) {
  return (
    <Item variant="outline" className={isFromUser ? 'border-primary' : ''}>
      <ItemContent>
        <ItemTitle>{title}</ItemTitle>
        <ItemDescription>
          {new Date(date).toLocaleString()}
          {kommentar && (
            <>
              <br />
              {kommentar}
            </>
          )}
        </ItemDescription>
      </ItemContent>
      <ItemActions>
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
              {new Date(date).toLocaleDateString()} um{' '}
              {new Date(date).toLocaleTimeString()} Uhr
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
