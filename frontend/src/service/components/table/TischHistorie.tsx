import type { LucideIcon } from 'lucide-react'
import {
  ArrowRightLeft,
  Banknote,
  ChevronRight,
  Plus,
  RotateCcw,
} from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { ItemGroup } from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { cn, formatCents, formatRelativeTime } from '@/lib/utils'

import { belegDruckenMitNachfassen, meldeBelegStatus } from '../../beleg'
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

type HistorieEintrag = Bestellung | Zahlung | Stornierung | Umbuchung

// Quelle ist ein Eintrag, der Positionen auf den Tisch bringt — eine Bestellung oder
// der Zugang einer Umbuchung. Nur diese tragen stornier-/umbuchbare Positionen.
type Quelle = Bestellung | Umbuchung

// Betragsfarbe: Zugänge (Bestellung, Umbuchungs-Zugang) emerald, kassenwirksame
// Storni rot, Zahlung und Umbuchungs-Abgang neutral.
type Betragsfarbe = 'zugang' | 'storno' | 'neutral'

interface Zeilenmodell {
  icon: LucideIcon
  iconWrapper: string
  title: string
  betrag: string
  betragFarbe: Betragsfarbe
  date: string
  userName: string
  kommentar: string
}

function zeilenmodell(item: HistorieEintrag): Zeilenmodell {
  switch (item.art) {
    case 'bestellung':
      return {
        icon: Plus,
        iconWrapper: 'bg-primary/10 text-primary',
        title: 'Bestellung',
        betrag: `+${formatCents(item.gesamtPreisCents)} €`,
        betragFarbe: 'zugang',
        date: item.aufgenommenAm,
        userName: item.userName,
        kommentar: item.kommentar,
      }
    case 'zahlung':
      return {
        icon: Banknote,
        iconWrapper: 'bg-muted text-muted-foreground',
        title: 'Zahlung',
        betrag: `-${formatCents(item.gesamtZahlungCents)} €`,
        betragFarbe: 'neutral',
        date: item.kassiertAm,
        userName: item.userName,
        kommentar: item.kommentar,
      }
    case 'umbuchung': {
      const istZugang = item.tischId === item.zielTischId
      return {
        icon: ArrowRightLeft,
        iconWrapper: 'bg-muted text-muted-foreground',
        // Der Autotext („Umbuchung von/auf Tisch X") ist selbst der Titel.
        title: item.kommentar,
        betrag: `${istZugang ? '+' : '-'}${formatCents(item.gesamtCents)} €`,
        betragFarbe: istZugang ? 'zugang' : 'neutral',
        date: item.umgebuchtAm,
        userName: item.userName,
        kommentar: '',
      }
    }
    case 'stornierung':
      return {
        icon: RotateCcw,
        iconWrapper: 'bg-destructive/10 text-destructive',
        title: item.barRueckgabe ? 'Warenrücknahme' : 'Korrektur',
        betrag: `-${formatCents(item.gesamtStornierungCents)} €`,
        betragFarbe: 'storno',
        date: item.storniertAm,
        userName: item.userName,
        kommentar: item.kommentar,
      }
  }
}

interface TischHistorieProps {
  historie: HistorieEintrag[]
  historieLoading: boolean
  tisch: Tisch
  backend: Pick<
    TischBackend,
    | 'stornierungErteilen'
    | 'bestellungUmbuchen'
    | 'belegDrucken'
    | 'stornobelegDrucken'
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
        stornierung_not_found: 'Die Stornierung wurde nicht gefunden.',
      },
    })

  const stornobelegAnfordern = (stornierungId: string) => {
    void runBelegDrucken(async () => {
      const status = await belegDruckenMitNachfassen(() =>
        backend.stornobelegDrucken(tisch.id, stornierungId),
      )
      meldeBelegStatus(
        status,
        'Stornobeleg in die Druckwarteschlange eingereiht.',
      )
    })
  }

  return (
    <>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historieLoading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <ItemSkeleton key={index} />
            ))
          : historie.map((item) => (
              <HistoryRow
                key={item.id}
                {...zeilenmodell(item)}
                onClick={() => {
                  setDetail(item)
                }}
              />
            ))}
      </ItemGroup>
      {detail &&
        (() => {
          const canStornieren =
            (detail.art === 'bestellung' || detail.art === 'umbuchung') &&
            AuthSingleton.canCancel &&
            detail.stornierbarePositionen.length > 0
          const canUmbuchen =
            (detail.art === 'bestellung' || detail.art === 'umbuchung') &&
            AuthSingleton.canRebook &&
            detail.umbuchbarePositionen.length > 0
          const zeile = zeilenmodell(detail)
          return (
            <Details
              {...detailView(detail)}
              title={zeile.title}
              userName={detail.userName}
              tischName={tisch.name}
              kommentar={zeile.kommentar}
              onClose={() => {
                setDetail(null)
              }}
              onStornieren={
                canStornieren
                  ? () => {
                      setDetail(null)
                      setStornierenQuelle(detail)
                    }
                  : undefined
              }
              onUmbuchen={
                canUmbuchen
                  ? () => {
                      setDetail(null)
                      setUmbuchenQuelle(detail)
                    }
                  : undefined
              }
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
                  : // Nur die kassenwirksame Warenrücknahme (Bargeld zurück)
                    // erzeugt einen Stornobeleg; die geldneutrale Korrektur nicht.
                    detail.art === 'stornierung' && detail.barRueckgabe
                    ? {
                        label: 'Stornobeleg drucken',
                        loading: belegDruckenLoading,
                        onAction: () => {
                          stornobelegAnfordern(detail.id)
                        },
                      }
                    : undefined
              }
            />
          )
        })()}
      {stornierenQuelle && (
        <HistorieStornierungDrawer
          backend={backend}
          tisch={tisch}
          quelle={stornierenQuelle}
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
          quelle={umbuchenQuelle}
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

const betragKlassen: Record<Betragsfarbe, string> = {
  zugang: 'text-emerald-600',
  storno: 'text-destructive',
  neutral: 'text-foreground',
}

function HistoryRow({
  icon: Icon,
  iconWrapper,
  title,
  betrag,
  betragFarbe,
  date,
  userName,
  kommentar,
  onClick,
}: Zeilenmodell & { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:bg-muted/50"
    >
      <span
        className={cn(
          'flex size-10 shrink-0 items-center justify-center rounded-full [&_svg]:size-5',
          iconWrapper,
        )}
      >
        <Icon />
      </span>
      <span className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="truncate text-[15px] font-medium">{title}</span>
        <span className="text-sm text-muted-foreground">
          {formatRelativeTime(date)} · {userName}
          {kommentar && ` · „${kommentar}“`}
        </span>
      </span>
      <span
        className={cn(
          'shrink-0 text-[15px] font-bold tabular-nums',
          betragKlassen[betragFarbe],
        )}
      >
        {betrag}
      </span>
      <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
    </button>
  )
}

function ItemSkeleton() {
  return (
    <div className="flex items-center gap-3 rounded-md border px-3 py-3">
      <Skeleton className="size-10 shrink-0 rounded-full" />
      <div className="flex flex-1 flex-col gap-1">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-48" />
      </div>
      <Skeleton className="h-4 w-16" />
    </div>
  )
}

// detailView maps a history entry to the fields the detail drawer renders. Titel
// und Kommentar liefert stattdessen zeilenmodell(), damit Zeile und Detail nie
// auseinanderlaufen (Umbuchung: Autotext als Titel, kein Kommentar-Widget).
function detailView(eintrag: HistorieEintrag): {
  date: string
  positionen?: ReceiptPosition[]
  totalPrice?: number
} {
  switch (eintrag.art) {
    case 'bestellung':
      return {
        date: eintrag.aufgenommenAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtPreisCents,
      }
    case 'zahlung':
      return {
        date: eintrag.kassiertAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtZahlungCents,
      }
    case 'stornierung':
      return {
        date: eintrag.storniertAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtStornierungCents,
      }
    case 'umbuchung':
      return {
        date: eintrag.umgebuchtAm,
        positionen: toReceiptItems(eintrag.positionen),
        totalPrice: eintrag.gesamtCents,
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
  date,
  userName,
  tischName,
  kommentar,
  positionen,
  totalPrice,
  primaryAction,
  onStornieren,
  onUmbuchen,
}: {
  onClose: () => void
  title: string
  date: string
  userName: string
  tischName: string
  kommentar: string
  positionen?: ReceiptPosition[]
  totalPrice?: number
  primaryAction?: PrimaryAction
  onStornieren?: () => void
  onUmbuchen?: () => void
}) {
  return (
    <Drawer
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DrawerContent>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>
            {title} · {formatRelativeTime(date)} · {userName}
          </DrawerTitle>
          <DrawerDescription>
            {tischName} ·{' '}
            {new Date(date).toLocaleString('de-DE', {
              day: 'numeric',
              month: 'numeric',
              year: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
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
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
          {(onUmbuchen ?? onStornieren) && (
            <div className="flex gap-2">
              {onUmbuchen && (
                <Button
                  variant="outline"
                  className="flex-1"
                  disabled={primaryAction?.loading}
                  onClick={onUmbuchen}
                >
                  <ArrowRightLeft /> Umbuchen
                </Button>
              )}
              {onStornieren && (
                <Button
                  variant="outline"
                  className="flex-1 border-destructive/40 text-destructive"
                  disabled={primaryAction?.loading}
                  onClick={onStornieren}
                >
                  <RotateCcw /> Stornieren…
                </Button>
              )}
            </div>
          )}
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
      </DrawerContent>
    </Drawer>
  )
}
