import type { LucideIcon } from 'lucide-react'
import {
  ArrowRightLeft,
  Banknote,
  ChevronRight,
  Plus,
  RotateCcw,
} from 'lucide-react'
import type { Dispatch, SetStateAction } from 'react'
import { useState } from 'react'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
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
import { useActionSubmit } from '@/hooks/use-action-submit'
import { AuthSingleton } from '@/lib/Auth'
import { cn, formatEuro, formatRelativeTime } from '@/lib/utils'

import { belegDruckenMitNachfassen, meldeBelegStatus } from '../../beleg'
import type { Bestellung } from '../../table/Bestellung'
import type { Stornierung } from '../../table/Stornierung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import type { Umbuchung } from '../../table/Umbuchung'
import type { Zahlung } from '../../table/Zahlung'
import { Kommentar } from './CommentField'
import { toReceiptItems } from './drawerUtils'
import { HistorieRowSkeleton } from './HistorieRowSkeleton'
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
  // Beleg-Daten für den Detail-Drawer; jeder Eintrag wird hier einmalig
  // gemappt (Zeile und Detail teilen dasselbe Modell).
  positionen: ReceiptPosition[]
  totalPrice: number
}

function zeilenmodell(item: HistorieEintrag): Zeilenmodell {
  switch (item.art) {
    case 'bestellung':
      return {
        icon: Plus,
        iconWrapper: 'bg-primary/10 text-primary',
        title: 'Bestellung',
        betrag: `+${formatEuro(item.gesamtPreisCents)}`,
        betragFarbe: 'zugang',
        date: item.aufgenommenAm,
        userName: item.userName,
        kommentar: item.kommentar,
        positionen: toReceiptItems(item.positionen),
        totalPrice: item.gesamtPreisCents,
      }
    case 'zahlung':
      return {
        icon: Banknote,
        iconWrapper: 'bg-muted text-muted-foreground',
        title: 'Zahlung',
        betrag: `-${formatEuro(item.gesamtZahlungCents)}`,
        betragFarbe: 'neutral',
        date: item.kassiertAm,
        userName: item.userName,
        kommentar: item.kommentar,
        positionen: toReceiptItems(item.positionen),
        totalPrice: item.gesamtZahlungCents,
      }
    case 'umbuchung': {
      const istZugang = item.tischId === item.zielTischId
      return {
        icon: ArrowRightLeft,
        iconWrapper: 'bg-muted text-muted-foreground',
        // Der Autotext („Umbuchung von/auf Tisch X") ist selbst der Titel; das
        // optionale Benutzerkommentar erscheint in Unterzeile und Detail.
        title: item.kommentar,
        betrag: `${istZugang ? '+' : '-'}${formatEuro(item.gesamtCents)}`,
        betragFarbe: istZugang ? 'zugang' : 'neutral',
        date: item.umgebuchtAm,
        userName: item.userName,
        kommentar: item.benutzerKommentar,
        positionen: toReceiptItems(item.positionen),
        totalPrice: item.gesamtCents,
      }
    }
    case 'stornierung':
      return {
        icon: RotateCcw,
        iconWrapper: 'bg-destructive/10 text-destructive',
        title: item.barRueckgabe ? 'Warenrücknahme' : 'Korrektur',
        betrag: `-${formatEuro(item.gesamtStornierungCents)}`,
        betragFarbe: 'storno',
        date: item.storniertAm,
        userName: item.userName,
        kommentar: item.kommentar,
        positionen: toReceiptItems(item.positionen),
        totalPrice: item.gesamtStornierungCents,
      }
  }
}

interface TischHistorieProps {
  historie: HistorieEintrag[]
  historieLoading: boolean
  // Das Erstladen der Historie ist gescheitert (kein brauchbarer Cache-Stand).
  // Ein gescheiterter Hintergrund-Refetch setzt die Flagge nicht: Die zuletzt
  // geladene Historie bleibt stehen, die Meldung trägt der zentrale
  // Fehler-Toast.
  historieError: boolean
  // Lädt die Tischdaten nach einem Ladefehler erneut.
  onErneutVersuchen: () => void
  tisch: Tisch
  backend: Pick<
    TischBackend,
    | 'stornierungErteilen'
    | 'bestellungUmbuchen'
    | 'belegDrucken'
    | 'stornobelegDrucken'
  >
  // Buchungserfolg (Stornierung/Umbuchung) meldet der Aufrufer über den
  // Erfolgs-Pop; der nachgelagerte Refetch läuft dort beim Schließen.
  onErfolg: (nachricht: string) => void
}

export function TischHistorie({
  historie,
  historieLoading,
  historieError,
  onErneutVersuchen,
  tisch,
  backend,
  onErfolg,
}: TischHistorieProps) {
  const [detail, setDetail] = useState<HistorieEintrag | null>(null)
  const [stornierenQuelle, setStornierenQuelle] = useState<Quelle | null>(null)
  const [umbuchenQuelle, setUmbuchenQuelle] = useState<Quelle | null>(null)

  // Eine leere Historie behauptet, an diesem Tisch sei noch nichts gebucht —
  // bei einem Ladefehler ist das falsch.
  if (historieError) {
    return (
      <LadefehlerAlert
        titel="Historie konnte nicht geladen werden"
        onErneutVersuchen={onErneutVersuchen}
        className="mt-4"
      />
    )
  }

  return (
    <>
      <ItemGroup className="grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
        {historieLoading
          ? Array.from({ length: 6 }).map((_, index) => (
              // eslint-disable-next-line react-x/no-array-index-key
              <HistorieRowSkeleton key={index} />
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
      {detail && (
        <HistorieDetail
          detail={detail}
          zeile={zeilenmodell(detail)}
          tisch={tisch}
          backend={backend}
          setDetail={setDetail}
          setStornierenQuelle={setStornierenQuelle}
          setUmbuchenQuelle={setUmbuchenQuelle}
        />
      )}
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
            onErfolg('Stornierung gebucht.')
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
          onBestellungUmgebucht={(zielName) => {
            setUmbuchenQuelle(null)
            onErfolg(`Auf ${zielName} umgebucht.`)
          }}
        />
      )}
    </>
  )
}

// HistorieDetail rendert den Detail-Drawer eines Historien-Eintrags: leitet die
// Aktions-Berechtigungen intern ab und kümmert sich um den Belegdruck (Zahlung →
// Kassenbeleg, Warenrücknahme → Stornobeleg). Das Zeilenmodell kommt vom
// Aufrufer, damit Zeile und Detail denselben Eintrag genau einmal mappen.
function HistorieDetail({
  detail,
  zeile,
  tisch,
  backend,
  setDetail,
  setStornierenQuelle,
  setUmbuchenQuelle,
}: {
  detail: HistorieEintrag
  zeile: Zeilenmodell
  tisch: Tisch
  backend: Pick<TischBackend, 'belegDrucken' | 'stornobelegDrucken'>
  setDetail: Dispatch<SetStateAction<HistorieEintrag | null>>
  setStornierenQuelle: Dispatch<SetStateAction<Quelle | null>>
  setUmbuchenQuelle: Dispatch<SetStateAction<Quelle | null>>
}) {
  const { loading: belegDruckenLoading, run: runBelegDrucken } =
    useActionSubmit({
      actionLabel: 'Kassenbeleg drucken',
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

  const canStornieren =
    (detail.art === 'bestellung' || detail.art === 'umbuchung') &&
    AuthSingleton.canCancel &&
    detail.stornierbarePositionen.length > 0
  const canUmbuchen =
    (detail.art === 'bestellung' || detail.art === 'umbuchung') &&
    AuthSingleton.canRebook &&
    detail.umbuchbarePositionen.length > 0

  return (
    <Details
      title={zeile.title}
      date={zeile.date}
      positionen={zeile.positionen}
      totalPrice={zeile.totalPrice}
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
              label: 'Kassenbeleg drucken',
              loading: belegDruckenLoading,
              onAction: () => {
                void runBelegDrucken(async () => {
                  const status = await belegDruckenMitNachfassen(() =>
                    backend.belegDrucken(tisch.id, detail.id),
                  )
                  meldeBelegStatus(
                    status,
                    'Kassenbeleg in die Druckwarteschlange eingereiht.',
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
                <p className="font-bold">{formatEuro(totalPrice)}</p>
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
                  onClick={onUmbuchen}
                >
                  <ArrowRightLeft /> Umbuchen
                </Button>
              )}
              {onStornieren && (
                <Button
                  variant="outline"
                  className="flex-1 border-destructive/40 text-destructive"
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
