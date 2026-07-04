import { ArrowRightLeft, Eye, X } from 'lucide-react'
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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { formatCents } from '@/lib/utils'

import { belegDruckenMitNachfassen, meldeBelegStatus } from '../../beleg'
import type { Ausgabe } from '../../table/Ausgabe'
import type { Bestellung } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import type { Umbuchung } from '../../table/Umbuchung'
import type { Zahlung } from '../../table/Zahlung'
import { Kommentar } from './CommentField'
import { toReceiptItems } from './drawerUtils'
import { HistorieStornierungDrawer } from './HistorieStornierungDrawer'
import { HistorieUmbuchungDrawer } from './HistorieUmbuchungDrawer'
import { Receipt, type ReceiptPosition } from './Receipt'

type HistorieEintrag = Bestellung | Zahlung | Stornierung | Umbuchung | Ausgabe

// Quelle ist ein Eintrag, der Positionen auf den Tisch bringt — eine Bestellung oder
// der Zugang einer Umbuchung. Nur diese tragen stornier-/umbuchbare Positionen.
type Quelle = Bestellung | Umbuchung

interface TischHistorieProps {
  historie: HistorieEintrag[]
  historieLoading: boolean
  tisch: Tisch
  backend: Pick<
    TischBackend,
    'stornierungErteilen' | 'bestellungUmbuchen' | 'belegDrucken'
  >
  onStornierungErteilt: () => void
  onBestellungUmgebucht: () => void
}

export function TischHistorie({
  historie,
  historieLoading,
  tisch,
  backend,
  onStornierungErteilt,
  onBestellungUmgebucht,
}: TischHistorieProps) {
  const [detail, setDetail] = useState<HistorieEintrag | null>(null)
  const [stornierenQuelle, setStornierenQuelle] = useState<Quelle | null>(null)
  const [umbuchenQuelle, setUmbuchenQuelle] = useState<Quelle | null>(null)
  const { loading: belegDruckenLoading, run: runBelegDrucken } =
    useActionSubmit({
      actionLabel: 'Beleg drucken',
      byCode: {
        kassenbeleg_drucker_nicht_konfiguriert:
          'Kein Kassenbeleg-Drucker konfiguriert. Bitte in den Admin-Einstellungen hinterlegen.',
        zahlung_not_found: 'Die ausgewählte Zahlung wurde nicht gefunden.',
      },
    })

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historieLoading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : historie.map((item) => {
              switch (item.art) {
                case 'zahlung':
                  return (
                    <HistoryItem
                      key={item.id}
                      title={`Zahlung -${formatCents(item.gesamtZahlungCents)} €`}
                      date={item.kassiertAm}
                      userName={item.userName}
                      kommentar={item.kommentar}
                      onClick={() => {
                        setDetail(item)
                      }}
                    />
                  )
                case 'bestellung':
                  return (
                    <HistoryItem
                      key={item.id}
                      title={`Bestellung +${formatCents(item.gesamtPreisCents)} €`}
                      date={item.aufgenommenAm}
                      userName={item.userName}
                      kommentar={item.kommentar}
                      onClick={() => {
                        setDetail(item)
                      }}
                      onStornieren={
                        AuthSingleton.canCancel &&
                        item.stornierbarePositionen.length > 0
                          ? () => {
                              setStornierenQuelle(item)
                            }
                          : undefined
                      }
                      onUmbuchen={
                        AuthSingleton.canRebook &&
                        item.umbuchbarePositionen.length > 0
                          ? () => {
                              setUmbuchenQuelle(item)
                            }
                          : undefined
                      }
                    />
                  )
                case 'umbuchung': {
                  const istZugang = item.tischId === item.zielTischId
                  return (
                    <HistoryItem
                      key={item.id}
                      title={`Umbuchung ${istZugang ? '+' : '-'}${formatCents(item.gesamtCents)} €`}
                      date={item.umgebuchtAm}
                      userName={item.userName}
                      kommentar={item.kommentar}
                      onClick={() => {
                        setDetail(item)
                      }}
                      onStornieren={
                        AuthSingleton.canCancel &&
                        item.stornierbarePositionen.length > 0
                          ? () => {
                              setStornierenQuelle(item)
                            }
                          : undefined
                      }
                      onUmbuchen={
                        AuthSingleton.canRebook &&
                        item.umbuchbarePositionen.length > 0
                          ? () => {
                              setUmbuchenQuelle(item)
                            }
                          : undefined
                      }
                    />
                  )
                }
                case 'stornierung':
                  return (
                    <HistoryItem
                      key={item.id}
                      title={`Stornierung -${formatCents(item.gesamtStornierungCents)} €`}
                      date={item.storniertAm}
                      userName={item.userName}
                      kommentar={item.kommentar}
                      onClick={() => {
                        setDetail(item)
                      }}
                    />
                  )
                case 'ausgabe':
                  return (
                    <HistoryItem
                      key={item.id}
                      title="Ausgabe"
                      date={item.ausgegebenAm}
                      userName={item.userName}
                      kommentar={item.kommentar}
                      onClick={() => {
                        setDetail(item)
                      }}
                    />
                  )
              }
            })}
      </ItemGroup>
      {detail && (
        <Details
          {...detailView(detail)}
          id={detail.id}
          userName={detail.userName}
          kommentar={detail.kommentar}
          onClose={() => {
            setDetail(null)
          }}
          primaryAction={
            detail.art === 'zahlung'
              ? {
                  label: 'Beleg drucken',
                  loading: belegDruckenLoading,
                  onAction: () => {
                    void runBelegDrucken(async () => {
                      const status = await belegDruckenMitNachfassen(() =>
                        backend.belegDrucken(tisch.id, detail.id),
                      )
                      meldeBelegStatus(
                        status,
                        'Beleg in die Druckwarteschlange eingereiht.',
                      )
                    })
                  },
                }
              : undefined
          }
        />
      )}
      {stornierenQuelle && (
        <HistorieStornierungDrawer
          backend={backend}
          tisch={tisch}
          vorgangId={stornierenQuelle.id}
          positionen={stornierenQuelle.stornierbarePositionen}
          onClose={() => {
            setStornierenQuelle(null)
          }}
          onStornierungErteilt={() => {
            setStornierenQuelle(null)
            onStornierungErteilt()
          }}
        />
      )}
      {umbuchenQuelle && (
        <HistorieUmbuchungDrawer
          backend={backend}
          tisch={tisch}
          vorgangId={umbuchenQuelle.id}
          positionen={umbuchenQuelle.umbuchbarePositionen}
          onClose={() => {
            setUmbuchenQuelle(null)
          }}
          onBestellungUmgebucht={() => {
            setUmbuchenQuelle(null)
            onBestellungUmgebucht()
          }}
        />
      )}
    </>
  )
}

function HistoryItem({
  title,
  date,
  userName,
  kommentar,
  onClick,
  onStornieren,
  onUmbuchen,
}: {
  title: string
  date: string
  userName: string
  kommentar: string
  onClick: () => void
  onStornieren?: () => void
  onUmbuchen?: () => void
}) {
  return (
    <Item variant="outline">
      <ItemContent>
        <ItemTitle>{title}</ItemTitle>
        <ItemDescription>
          <span className="block">von {userName}</span>
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
        {onUmbuchen && (
          <Button
            size="icon-sm"
            variant="outline"
            className="rounded-full cursor-pointer"
            aria-label="Umbuchen"
            onClick={onUmbuchen}
          >
            <ArrowRightLeft />
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

// detailView maps a history entry to the fields the detail drawer renders.
function detailView(eintrag: HistorieEintrag): {
  title: string
  date: string
  positionen?: ReceiptPosition[]
  totalPrice?: number
} {
  switch (eintrag.art) {
    case 'bestellung':
      return {
        title: 'Bestellung',
        date: eintrag.aufgenommenAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtPreisCents,
      }
    case 'zahlung':
      return {
        title: 'Zahlung',
        date: eintrag.kassiertAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtZahlungCents,
      }
    case 'stornierung':
      return {
        title: 'Stornierung',
        date: eintrag.storniertAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtStornierungCents,
      }
    case 'umbuchung':
      return {
        title: 'Umbuchung',
        date: eintrag.umgebuchtAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtCents,
      }
    case 'ausgabe':
      return {
        title: 'Ausgabe',
        date: eintrag.ausgegebenAm,
        positionen: toReceiptItems(eintrag.positionen),
      }
  }
}

interface PrimaryAction {
  label: string
  loading: boolean
  onAction: () => void
}

function Details({
  onClose,
  title,
  id,
  date,
  userName,
  kommentar,
  positionen,
  totalPrice,
  primaryAction,
}: {
  onClose: () => void
  title: string
  id: string
  date: string
  userName: string
  kommentar: string
  positionen?: ReceiptPosition[]
  totalPrice?: number
  primaryAction?: PrimaryAction
}) {
  return (
    <Drawer
      open
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
              von {userName} am {new Date(date).toLocaleDateString('de-DE')} um{' '}
              {new Date(date).toLocaleTimeString('de-DE')} Uhr
            </DrawerDescription>
          </DrawerHeader>
          {positionen ? (
            <Receipt positionen={positionen} totalPrice={totalPrice} />
          ) : (
            totalPrice !== undefined && (
              <div className="px-4 py-2">
                <p className="font-bold">{formatCents(totalPrice)}&nbsp;€</p>
              </div>
            )
          )}
          {kommentar && (
            <div className="px-4">
              <Kommentar value={kommentar} />
            </div>
          )}
          <DrawerFooter>
            {primaryAction && (
              <Button
                onClick={primaryAction.onAction}
                disabled={primaryAction.loading}
              >
                {primaryAction.loading ? 'Drucke…' : primaryAction.label}
              </Button>
            )}
            <DrawerClose asChild>
              <Button variant="outline" disabled={primaryAction?.loading}>
                Schließen
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
